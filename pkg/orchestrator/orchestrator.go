package orchestrator

import (
	"context"
	"sync"

	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/task"
)

func NewOrchestrator(ctx context.Context, catalog job.Catalog, maxRunners int) *Orchestrator {
	chScheduler := make(chan schedulerUpdate)
	s, chTick := newScheduler(ctx, chScheduler)
	r, chRunner := NewRunner(ctx, maxRunners)
	return &Orchestrator{
		ctx:         ctx,
		catalog:     catalog,
		handlerRepo: task.NewHandlerRepository(),
		scheduler:   s,
		runner:      r,
		chScheduler: chScheduler,
		chRunner:    chRunner,
		chTick:      chTick,
		mux:         sync.RWMutex{},
	}
}

type Orchestrator struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	catalog     job.Catalog             // contains jobs
	handlerRepo *task.HandlerRepository // contains task handlers
	scheduler   *scheduler              // manages tickers for job schedule
	runner      *runner                 // manages job runners
	chScheduler chan schedulerUpdate    // channel to send updates to the scheduler
	chRunner    chan tick               // channel to send jobs to runner
	chTick      chan tick               // channel to receive ticks from scheduler
	queue       []tick                  // jobs to be queued for execution

	mux sync.RWMutex
}

func (o *Orchestrator) Start() {
	ctx, cancel := context.WithCancel(o.ctx)

	o.mux.Lock()
	o.cancelFunc = cancel
	o.mux.Unlock()

	go o.sendToRunner(ctx)
	go o.listenForTicks(ctx)
}

func (o *Orchestrator) Stop() {
	o.cancelFunc()
}

func (o *Orchestrator) Statistics() Statistics {
	o.mux.RLock()
	defer o.mux.RUnlock()

	handlerPoolStats := o.handlerRepo.Statistics()
	queueLength := len(o.queue)

	return Statistics{
		HandlerPoolStatistics: handlerPoolStats,
		QueueLength:           queueLength,
	}
}

func (o *Orchestrator) listenForTicks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-o.chTick:
			o.mux.Lock()
			o.queue = append(o.queue, t)
			o.mux.Unlock()
		}
	}
}

func (o *Orchestrator) sendToRunner(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			o.mux.Lock()
			if len(o.queue) == 0 {
				o.mux.Unlock()
				break
			}
			t := o.queue[0]
			o.queue = o.queue[1:]
			o.mux.Unlock()

			o.chRunner <- t
		}
	}
}
