package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/jantytgat/go-jobs/pkg/task"
	"github.com/jantytgat/go-jobs/pkg/taskLibrary"
)

func main() {
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(logHandler)

	r := task.NewHandlerRepository("sequencePing")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tasks := make([]task.Task, 0)
	for i := 0; i < 1; i++ {
		tasks = append(tasks, []task.Task{
			taskLibrary.ExecTask{
				Program: "ping",
				Path:    "/sbin",
				Args:    []string{"-c", "10", "127.0.0.1"},
			},
			taskLibrary.LogTask{Message: fmt.Sprintf("Task %d", i), Level: slog.LevelDebug},
		}...)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := task.ExecuteSequence(ctx, logger, tasks, r); err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()
	fmt.Println(r.Statistics())
}
