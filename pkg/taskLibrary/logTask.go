package taskLibrary

import (
	"context"
	"log/slog"
	"time"

	"github.com/jantytgat/go-jobs/pkg/task"
)

const (
	logTaskName                  = "LogTask"
	logTaskMaxConcurrency        = 0
	logTaskHandlerDefaultTimeout = time.Duration(1) * time.Second
)

type LogTask struct {
	Level   slog.Level
	Message string
}

func (t LogTask) Name() string {
	return logTaskName
}

func (t LogTask) DefaultHandler() task.Handler {
	return LogTaskHandler(logTaskHandlerDefaultTimeout)
}

func (t LogTask) DefaultHandlerPool(ctx context.Context) *task.HandlerPool {
	return LogTaskHandlerPool(ctx, logTaskHandlerDefaultTimeout)
}

func (t LogTask) Handler(timeout time.Duration) task.Handler {
	return LogTaskHandler(timeout)
}

func (t LogTask) HandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return LogTaskHandlerPool(ctx, timeout)
}

func LogTaskHandler(timeout time.Duration) task.Handler {
	return task.NewHandler(logTaskName, timeout, handleLogTask)
}

func LogTaskHandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return task.NewHandlerPool(ctx, task.NewHandler(logTaskName, timeout, handleLogTask), logTaskMaxConcurrency)
}

func handleLogTask(ctx context.Context, t task.Task, p *task.Pipeline) error {
	lt := t.(LogTask)
	p.Logger(t).LogAttrs(ctx, lt.Level, lt.Message)
	return nil
}
