package orchestrator

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/task"
)

func newDispatcher(maxRunners int, chDispatcher chan dispatcherMessage, chResults chan job.Result) *dispatcher {
	if maxRunners < 1 {
		maxRunners = 1
	}
	return &dispatcher{
		maxRunners:   maxRunners,
		chDispatcher: chDispatcher,
		chResults:    chResults,
	}
}

type dispatcher struct {
	maxRunners   int
	cancel       context.CancelFunc
	chDispatcher chan dispatcherMessage
	chResults    chan job.Result
	mux          sync.Mutex
}

func (d *dispatcher) Start(ctx context.Context, logger *slog.Logger) {
	var runnerCtx context.Context
	runnerCtx, d.cancel = context.WithCancel(ctx)

	for i := 0; i < d.maxRunners; i++ {
		go d.dispatch(runnerCtx, logger)
	}
}

func (d *dispatcher) dispatch(ctx context.Context, logger *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-d.chDispatcher:
			startTime := time.Now()
			taskResults, err := task.ExecuteSequence(ctx, logger, msg.job.Tasks, msg.handlerRepository)
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
