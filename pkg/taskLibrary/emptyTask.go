package taskLibrary

import (
	"context"
	"time"

	"github.com/jantytgat/go-jobs/pkg/task"
)

const (
	emptyTaskName                  = "EmptyTask"
	emptyTaskMaxConcurrency        = 0
	emptyTaskHandlerDefaultTimeout = time.Duration(1) * time.Second
)

type EmptyTask struct{}

func (t EmptyTask) Name() string {
	return emptyTaskName
}

func (t EmptyTask) DefaultHandler() task.Handler {
	return EmptyTaskHandler(emptyTaskHandlerDefaultTimeout)
}

func (t EmptyTask) DefaultHandlerPool(ctx context.Context) *task.HandlerPool {
	return EmptyTaskHandlerPool(ctx, emptyTaskHandlerDefaultTimeout)
}

func (t EmptyTask) Handler(timeout time.Duration) task.Handler {
	return EmptyTaskHandler(timeout)
}

func (t EmptyTask) HandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return EmptyTaskHandlerPool(ctx, timeout)
}

func EmptyTaskHandler(timeout time.Duration) task.Handler {
	return task.NewHandler(emptyTaskName, timeout, handleEmptyTask)
}

func EmptyTaskHandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return task.NewHandlerPool(ctx, task.NewHandler(emptyTaskName, timeout, handleEmptyTask), emptyTaskMaxConcurrency)
}

func handleEmptyTask(ctx context.Context, t task.Task, p *task.Pipeline) error {
	return nil
}
