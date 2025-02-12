package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jantytgat/go-jobs/pkg/task"
	"github.com/jantytgat/go-jobs/pkg/taskLibrary"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// hp := library.PrintTaskHandlerPool(ctx, time.Duration(5)*time.Second)
	hp := task.NewHandlerPool(ctx, taskLibrary.PrintTaskHandler(time.Duration(1)*time.Second), 24, task.WithHandlerPoolRecycling(5000))
	chResult := make(chan task.HandlerResult, 2000)

	resultsCtx, resultsCancel := context.WithCancel(context.Background())
	defer resultsCancel()
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
	}(resultsCtx)

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
			hp.ChPoolInput <- task.HandlerTask{
				Task:     taskLibrary.PrintTask{Message: fmt.Sprintf("Task %04d", i)},
				Pipeline: nil,
				ChResult: chResult,
			}
		}
	}
	fmt.Printf("Shutting down after %d tasks\n", i)
	cancel()

	if hp.IsRunning() {
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	stats := hp.Statistics()

	output, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(output))
	// fmt.Printf("A: %.0f / I: %.0f / R: %.0f / M: %.0f / TS: %.0f / TC: %.0f\r\n", stats.ActiveWorkers, stats.IdleWorkers, stats.Workers, stats.MaxWorkers, stats.TasksProcessedStatusSuccess, stats.TasksProcessedStatusCanceled)
	fmt.Printf("All tasks completed: %t (%d = %.0f)\n", float64(i) == stats.TasksProcessedStatusSuccess, i, stats.TasksProcessedStatusSuccess)
}
