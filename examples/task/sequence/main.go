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

	r := task.NewHandlerRepository()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tasks := make([]task.Task, 0)
	for i := 0; i < 20000; i++ {
		tasks = append(tasks, []task.Task{
			taskLibrary.EmptyTask{},
			taskLibrary.LogTask{Message: fmt.Sprintf("Task %d", i), Level: slog.LevelDebug},
		}...)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runSequence(ctx, logger, r, tasks)
		}()
	}
	wg.Wait()
	fmt.Println(r.Statistics())
}

func runSequence(ctx context.Context, logger *slog.Logger, r *task.HandlerRepository, tasks []task.Task) {
	s := task.NewSequence(logger, tasks)
	if err := s.Execute(ctx, r); err != nil {
		panic(err)
	}
}
