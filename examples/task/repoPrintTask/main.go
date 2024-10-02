package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jantytgat/go-jobs/pkg/library"
	"github.com/jantytgat/go-jobs/pkg/task"
)

func main() {
	r := task.NewHandlerRepository()
	ctx, cancel := context.WithCancel(context.Background())
	chResult := make(chan task.HandlerTaskResult, 2000)

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

	go func(ctx context.Context) {
		timer := time.After(time.Duration(5) * time.Second)
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
				if err := r.Execute(ctx, task.HandlerTask{
					Task:     library.EmptyTask{Message: fmt.Sprintf("Task %04d", i)},
					Pipeline: nil,
					ChResult: chResult,
				}); err != nil {
					fmt.Println("EMPTY: ", err)
				}
			}
		}
	}(ctx)

	timer := time.After(time.Duration(5) * time.Second)
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
			if err := r.Execute(ctx, task.HandlerTask{
				Task:     library.PrintTask{Message: fmt.Sprintf("Task %04d", i)},
				Pipeline: nil,
				ChResult: chResult,
			}); err != nil {
				fmt.Println("PRINT: ", err)
			}
		}
	}
	cancel()
	time.Sleep(time.Duration(500) * time.Millisecond)
	fmt.Println("Cancel is called")
	fmt.Println(r.Statistics())
}
