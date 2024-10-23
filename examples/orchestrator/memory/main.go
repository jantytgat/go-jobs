package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/cron"
	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/orchestrator"
	"github.com/jantytgat/go-jobs/pkg/task"
	"github.com/jantytgat/go-jobs/pkg/taskLibrary"
)

func main() {
	var err error
	var o *orchestrator.Orchestrator

	c := job.NewMemoryCatalog()
	r := task.NewHandlerRepository()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = r.RegisterHandlerPools([]*task.HandlerPool{
		task.NewHandlerPool(ctx, taskLibrary.EmptyTaskHandler(5*time.Second), 2000),
		task.NewHandlerPool(ctx, taskLibrary.LogTaskHandler(5*time.Second), 2000),
	}); err != nil {
		panic(err)
	}
	maxJobs := runtime.NumCPU()
	maxJobs = 5000
	if o, err = orchestrator.New(logger, maxJobs, orchestrator.WithCatalog(c), orchestrator.WithHandlerRepository(r)); err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 50000; i++ {
			var schedule cron.Schedule
			if i%2 == 0 {
				schedule, _ = cron.NewSchedule("*/2 * * * * *")
			} else if i%3 == 0 {
				schedule, _ = cron.NewSchedule("*/3 * * * * *")
			} else if i%5 == 0 {
				schedule, _ = cron.NewSchedule("*/5 * * * * *")
			} else {
				schedule, _ = cron.NewSchedule("* * * * * *")
			}
			schedule = cron.EverySecond()
			t := make([]task.Task, 0)
			t = append(t, taskLibrary.LogTask{Message: fmt.Sprintf("Hello %d", i)})
			t = append(t, taskLibrary.EmptyTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("this %d", i)})
			// t = append(t, taskLibrary.EmptyTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("is %d", i)})
			// t = append(t, taskLibrary.EmptyTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("my %d", i)})
			// t = append(t, taskLibrary.EmptyTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("message %d", i)})
			// t = append(t, taskLibrary.EmptyTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("Goodbye %d", i)})
			j := job.New(uuid.New(), "sequenceJob", schedule, t, job.WithRunLimit(1))
			if err = c.Add(j); err != nil {
				panic(err)
			}
		}
	}(wg)
	wg.Wait()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
			default:
				fmt.Println(c.Statistics(), o.Statistics())
				time.Sleep(1 * time.Second)
			}
		}

	}(ctx)

	fmt.Println("STARTING")
	_ = o.Start(ctx)
	time.Sleep(10 * time.Second)
	fmt.Println("STOPPING")
	o.Stop()
	cancel()
	time.Sleep(1 * time.Second)
	fmt.Println("FINAL STATS")
	fmt.Println(c.Statistics(), o.Statistics())
}
