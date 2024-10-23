package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/task"
)

func newDispatcher(logger *slog.Logger, maxRunners int, chDispatcher chan dispatcherMessage, chResults chan job.Result) *dispatcher {
	if maxRunners < 1 {
		maxRunners = 1
	}
	return &dispatcher{
		maxRunners:   maxRunners,
		chDispatcher: chDispatcher,
		chResults:    chResults,
		dispatchers:  make(map[int]context.CancelFunc),
		logger:       logger,
	}
}

type dispatcher struct {
	maxRunners   int
	cancel       context.CancelFunc
	chDispatcher chan dispatcherMessage
	chResults    chan job.Result
	dispatchers  map[int]context.CancelFunc
	logger       *slog.Logger
	mux          sync.Mutex
}

func (d *dispatcher) IsRunning() bool {
	d.mux.Lock()
	defer d.mux.Unlock()

	if d.cancel != nil {
		for i := 0; i < d.maxRunners; i++ {
			if _, found := d.dispatchers[i]; !found {
				return false
			}
		}
		return true
	}
	return false
}

func (d *dispatcher) Start(ctx context.Context) error {
	var runnerCtx context.Context
	runnerCtx, d.cancel = context.WithCancel(ctx)

	go d.run(runnerCtx)

	startCtx, startCancel := context.WithTimeout(ctx, 1*time.Second)
	defer startCancel()

	for {
		select {
		case <-startCtx.Done():
			d.logger.LogAttrs(ctx, slog.LevelError, "dispatcher start timeout")
			return fmt.Errorf("dispatcher start timeout")
		default:
			if d.IsRunning() { // TODO add timeout to start before returning an error?
				d.logger.LogAttrs(ctx, slog.LevelDebug, "dispatcher has started")
				return nil
			}
		}
	}
}

func (d *dispatcher) Stop() {
	d.cancel()
}

func (d *dispatcher) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for i := 0; i < d.maxRunners; i++ {
				d.mux.Lock()
				_, found := d.dispatchers[i]
				d.mux.Unlock()

				if !found {
					d.newInstance(ctx, i)
				}

			}
			// time.Sleep(1 * time.Second)
		}
	}
}

func (d *dispatcher) newInstance(ctx context.Context, id int) {
	d.mux.Lock()
	defer d.mux.Unlock()

	dispatchCtx, dispatchCancel := context.WithCancel(ctx)
	d.dispatchers[id] = dispatchCancel
	go d.runInstance(dispatchCtx, id)
}

func (d *dispatcher) removeInstance(id int) {
	d.mux.Lock()
	defer d.mux.Unlock()

	delete(d.dispatchers, id)
}

func (d *dispatcher) runInstance(ctx context.Context, id int) {
	d.logger.LogAttrs(ctx, slog.LevelDebug, "starting dispatcher instance", slog.Group("dispatcher", slog.Int("id", id)))
	defer d.logger.LogAttrs(ctx, slog.LevelDebug, "dispatcher instance stopped", slog.Group("dispatcher", slog.Int("id", id)))
	defer d.removeInstance(id)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-d.chDispatcher:
			startTime := time.Now()
			l := d.logger.WithGroup("job").With(slog.Int("dispatcher_id", id), slog.String("id", msg.job.Uuid.String()))
			runUuid := uuid.New()
			l.LogAttrs(ctx, slog.LevelInfo, "job starting", slog.String("instance", runUuid.String()))
			taskResults, err := task.ExecuteSequence(ctx, l, msg.job.Tasks, msg.handlerRepository)
			l.LogAttrs(ctx, slog.LevelInfo, "job finished", slog.String("instance", runUuid.String()))
			duration := time.Since(startTime)
			result := job.Result{
				Uuid:        msg.job.Uuid,
				RunUuid:     uuid.New(),
				Trigger:     msg.trigger,
				RunTime:     duration,
				TaskResults: taskResults,
				Error:       err,
			}
			d.chResults <- result
		}
	}
}
