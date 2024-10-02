package library

import (
	"context"
	"fmt"
	"time"

	"github.com/jantytgat/go-jobs/pkg/task"
)

const (
	printTaskName                  = "PrintTask"
	printTaskMaxConcurrency        = 0
	printTaskHandlerDefaultTimeout = time.Duration(1) * time.Second
)

type PrintTask struct {
	Message string
}

func (t PrintTask) Name() string {
	return printTaskName
}

func (t PrintTask) DefaultHandler() task.Handler {
	return PrintTaskHandler(printTaskHandlerDefaultTimeout)
}

func (t PrintTask) DefaultHandlerPool(ctx context.Context) *task.HandlerPool {
	return PrintTaskHandlerPool(ctx, printTaskHandlerDefaultTimeout)
}

func (t PrintTask) Handler(timeout time.Duration) task.Handler {
	return PrintTaskHandler(timeout)
}

func (t PrintTask) HandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return PrintTaskHandlerPool(ctx, timeout)
}

func PrintTaskHandler(timeout time.Duration) task.Handler {
	return task.NewHandler(printTaskName, timeout, handlePrintTask)
}

func PrintTaskHandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return task.NewHandlerPool(ctx, task.NewHandler(printTaskName, timeout, handlePrintTask), printTaskMaxConcurrency)
}

func handlePrintTask(ctx context.Context, t task.Task, p *task.Pipeline) error {
	pt := t.(PrintTask)
	fmt.Printf("%s\n", pt.Message)
	return nil
}
