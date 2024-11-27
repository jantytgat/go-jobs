package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jantytgat/go-jobs/pkg/task"
	"github.com/jantytgat/go-jobs/pkg/taskLibrary"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// hp := library.PrintTaskHandlerPool(ctx, time.Duration(5)*time.Second)
	hp := task.NewHandlerPool(ctx, taskLibrary.PrintTaskHandler(time.Duration(1)*time.Second), 0)
	chResult := make(chan task.HandlerResult, 2000)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-chResult:
				if result.Error != nil {
					fmt.Println(result)
				}
			}
		}
	}(ctx)

	timer := time.After(time.Duration(1) * time.Second)
	var exit bool
	var i int
	for {
		if exit {
			break
		}

		select {
		case <-timer:
			exit = true
		default:
			i++
			hp.Input <- task.HandlerTask{
				Task:     taskLibrary.PrintTask{Message: fmt.Sprintf("Task %04d", i)},
				Pipeline: nil,
				ChResult: chResult,
			}
		}
	}
	fmt.Println("Shutting down")
	cancel()
	time.Sleep(time.Duration(500) * time.Millisecond)
	fmt.Println("Cancel is called")
	stats := hp.Statistics()

	fmt.Printf("A: %d / I: %d / R: %d / M: %d / T: %d\r\n", stats.ActiveWorkers, stats.IdleWorkers, stats.Workers, stats.MaxWorkers, stats.TasksProcessed)
}
