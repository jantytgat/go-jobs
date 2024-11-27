package taskLibrary

import (
	"context"
	"fmt"
	"time"

	"github.com/jantytgat/go-jobs/pkg/task"
)

const (
	emptyErrorTaskName                  = "EmptyErrorTask"
	emptyErrorTaskMaxConcurrency        = 0
	emptyErrorTaskHandlerDefaultTimeout = time.Duration(1) * time.Second
)

type EmptyErrorTask struct{}

func (t EmptyErrorTask) Name() string {
	return emptyErrorTaskName
}

func (t EmptyErrorTask) DefaultHandler() task.Handler {
	return EmptyErrorTaskHandler(emptyErrorTaskHandlerDefaultTimeout)
}

func (t EmptyErrorTask) DefaultHandlerPool(ctx context.Context) *task.HandlerPool {
	return EmptyErrorTaskHandlerPool(ctx, emptyErrorTaskHandlerDefaultTimeout)
}

func (t EmptyErrorTask) Handler(timeout time.Duration) task.Handler {
	return EmptyErrorTaskHandler(timeout)
}

func (t EmptyErrorTask) HandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return EmptyErrorTaskHandlerPool(ctx, timeout)
}

func EmptyErrorTaskHandler(timeout time.Duration) task.Handler {
	return task.NewHandler(emptyErrorTaskName, timeout, handleEmptyErrorTask)
}

func EmptyErrorTaskHandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return task.NewHandlerPool(ctx, task.NewHandler(emptyErrorTaskName, timeout, handleEmptyErrorTask), emptyErrorTaskMaxConcurrency)
}

func handleEmptyErrorTask(ctx context.Context, t task.Task, p *task.Pipeline) error {
	return fmt.Errorf("%s", emptyErrorTaskName)
}
