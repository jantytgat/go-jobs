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
		runners:      make(map[int]context.CancelFunc),
		logger:       logger,
	}
}

type dispatcher struct {
	maxRunners   int
	cancelFunc   context.CancelFunc
	chDispatcher chan dispatcherMessage
	chResults    chan job.Result
	runners      map[int]context.CancelFunc
	logger       *slog.Logger
	mux          sync.Mutex
}

func (d *dispatcher) IsRunning() bool {
	d.mux.Lock()
	defer d.mux.Unlock()

	if d.cancelFunc != nil {
		for i := 0; i < d.maxRunners; i++ {
			if _, found := d.runners[i]; !found {
				return false
			}
		}
		return true
	}
	return false
}

func (d *dispatcher) Start(ctx context.Context) error {
	var dispatchCtx context.Context
	dispatchCtx, d.cancelFunc = context.WithCancel(ctx)

	go d.run(dispatchCtx)

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

func (d *dispatcher) Stop(ctx context.Context) {
	d.mux.Lock()
	if d.cancelFunc == nil {
		d.logger.LogAttrs(ctx, slog.LevelWarn, "dispatcher has stopped already")
		return
	}
	d.mux.Unlock()

	d.logger.LogAttrs(ctx, slog.LevelDebug, "dispatcher stopping")
	d.cancelFunc()

	d.mux.Lock()
	d.cancelFunc = nil
	d.mux.Unlock()
}

func (d *dispatcher) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			d.mux.Lock()
			for i := 0; i < d.maxRunners; i++ {
				if _, found := d.runners[i]; !found {
					runnerCtx, runnerCancel := context.WithCancel(ctx)
					d.runners[i] = runnerCancel
					go d.launchRunner(runnerCtx, i)
				}
			}
			d.mux.Unlock()
		}
	}
}

func (d *dispatcher) deleteRunner(id int) {
	d.mux.Lock()
	defer d.mux.Unlock()

	delete(d.runners, id)
}

func (d *dispatcher) launchRunner(ctx context.Context, id int) {
	d.logger.LogAttrs(ctx, slog.LevelDebug, "starting runner", slog.Group("runner", slog.Int("id", id)))
	defer d.logger.LogAttrs(ctx, slog.LevelDebug, "stopping runner", slog.Group("runner", slog.Int("id", id)))
	defer d.deleteRunner(id)

	d.logger.LogAttrs(ctx, slog.LevelDebug, "waiting for job", slog.Group("runner", slog.Int("id", id)))
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
			TriggerTime: msg.triggerTime,
			RunTime:     duration,
			TaskResults: taskResults,
			Error:       err,
		}
		d.chResults <- result
	}
}
