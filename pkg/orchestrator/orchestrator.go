package orchestrator

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/task"
)

func New(maxRunners int, opts ...Option) *Orchestrator {
	chScheduler := make(chan schedulerMessage, maxRunners)
	chTick := make(chan schedulerTick, maxRunners)
	chResults := make(chan job.Result, maxRunners)
	s := newScheduler(chScheduler, chTick)

	chDispatcher := make(chan dispatcherMessage, maxRunners)

	o := &Orchestrator{
		scheduler:    s,
		dispatcher:   newDispatcher(maxRunners, chDispatcher, chResults),
		chScheduler:  chScheduler,
		chDispatcher: chDispatcher,
		chTick:       chTick,
		chResults:    chResults,
		mux:          sync.Mutex{},
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.logger == nil {
		o.logger = slog.Default()
	}

	if o.Catalog == nil {
		o.Catalog = job.NewMemoryCatalog()
	}

	if o.Handlers == nil {
		o.Handlers = task.NewHandlerRepository()
	}

	if o.queue == nil {
		o.queue = NewMemoryQueue()
	}

	return o
}

type Orchestrator struct {
	ctx          context.Context
	cancelFunc   context.CancelFunc
	scheduler    *scheduler  // manages tickers for job schedule
	queue        Queue       // jobs to be queued for execution
	dispatcher   *dispatcher // manages job runners
	logger       *slog.Logger
	Catalog      job.Catalog             // contains jobs
	Handlers     *task.HandlerRepository // contains task handlers
	chScheduler  chan schedulerMessage   // channel to send updates to the scheduler
	chDispatcher chan dispatcherMessage  // channel to send jobs to dispatcher
	chResults    chan job.Result         // channel to get results from dispatcher
	chTick       chan schedulerTick      // channel to receive ticks from scheduler

	mux sync.Mutex
}

func (o *Orchestrator) Start(ctx context.Context) {
	o.mux.Lock()
	defer o.mux.Unlock()

	var oCtx context.Context
	oCtx, o.cancelFunc = context.WithCancel(ctx)

	o.scheduler.Start(oCtx)
	o.dispatcher.Start(oCtx, o.logger)

	if o.scheduler.IsRunning() {
		go o.storeResults(oCtx)
		go o.sendToDispatcher(oCtx)
		go o.sendToScheduler(oCtx)
		go o.listenToTickers(oCtx)
	}

}

func (o *Orchestrator) Statistics() Statistics {
	o.mux.Lock()
	defer o.mux.Unlock()

	handlerPoolStats := o.Handlers.Statistics()
	queueLength := o.queue.Length()

	return Statistics{
		HandlerPoolStatistics: handlerPoolStats,
		QueueLength:           queueLength,
	}
}

func (o *Orchestrator) Stop() {
	o.cancelFunc()
}

func (o *Orchestrator) sendToScheduler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			jobs := o.Catalog.All()
			if len(jobs) == 0 {
				time.Sleep(1 * time.Second)
				break
			}

			for _, j := range jobs {
				o.chScheduler <- schedulerMessage{
					uuid:     j.Uuid,
					enabled:  j.Enabled,
					schedule: j.Schedule,
				}
			}
			time.Sleep(10 * time.Second) // TODO increase sleep time?
		}
	}
}

func (o *Orchestrator) listenToTickers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-o.chTick:
			o.queue.Push(t)
		}
	}
}

func (o *Orchestrator) sendToDispatcher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var t schedulerTick
			var err error
			if t, err = o.queue.Pop(); err != nil {
				break
			}

			var j job.Job
			if j, err = o.Catalog.Get(t.uuid); err != nil {
				// TODO handle errors to job catalog when job cannot be found
				time.Sleep(5 * time.Second) // back off from catalog before getting next item off the queue
				break
			}
			o.chDispatcher <- dispatcherMessage{
				job:               j,
				handlerRepository: o.Handlers,
				trigger:           t.time,
			}
		}
	}
}

func (o *Orchestrator) storeResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-o.chResults:
			go o.Catalog.AddResult(r) // make sure results are read from the channel as fast as possible
		}
	}
}
