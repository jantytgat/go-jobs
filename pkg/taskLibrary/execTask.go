package taskLibrary

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jantytgat/go-jobs/pkg/task"
)

const (
	execTaskName                  = "ExecTask"
	execTaskMaxConcurrency        = 0
	execTaskHandlerDefaultTimeout = time.Duration(60) * time.Second
)

type ExecTask struct {
	Program string
	Path    string
	Args    []string
}

func (t ExecTask) Name() string {
	return t.name()
}

func (t ExecTask) DefaultHandler() task.Handler {
	return ExecTaskHandler(execTaskHandlerDefaultTimeout)
}

func (t ExecTask) DefaultHandlerPool(ctx context.Context) *task.HandlerPool {
	return ExecTaskHandlerPool(ctx, t.name(), execTaskHandlerDefaultTimeout)
}

func (t ExecTask) Handler(timeout time.Duration) task.Handler {
	return ExecTaskHandler(timeout)
}

func (t ExecTask) HandlerPool(ctx context.Context, timeout time.Duration) *task.HandlerPool {
	return ExecTaskHandlerPool(ctx, t.name(), timeout)
}

func (t ExecTask) name() string {
	return strings.Join([]string{
		execTaskName,
		t.Program,
	}, "_")
}

func ExecTaskHandler(timeout time.Duration) task.Handler {
	return task.NewHandler(execTaskName, timeout, handleExecTask)
}

func ExecTaskHandlerPool(ctx context.Context, name string, timeout time.Duration) *task.HandlerPool {
	return task.NewHandlerPool(ctx, task.NewHandler(name, timeout, handleExecTask), execTaskMaxConcurrency)
}

func handleExecTask(ctx context.Context, t task.Task, p *task.Pipeline) error {
	et := t.(ExecTask)
	out := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, strings.Join([]string{et.Path, et.Program}, string(os.PathSeparator)), et.Args...)
	cmd.Stdout = out

	var err error
	if err = cmd.Start(); err != nil {
		fmt.Println(err)
		return err
	}
	if err = cmd.Wait(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(out.Bytes()))
	return nil
}
