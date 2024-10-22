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

	c := job.NewMemoryCatalog()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	o := orchestrator.New(runtime.NumCPU(), orchestrator.WithCatalog(c), orchestrator.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))))
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 100; i++ {
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

			t := make([]task.Task, 0)
			t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("Hello %d", i)})
			t = append(t, taskLibrary.EmptyTask{})
			t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("this %d", i)})
			t = append(t, taskLibrary.EmptyTask{})
			t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("is %d", i)})
			t = append(t, taskLibrary.EmptyTask{})
			t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("my %d", i)})
			t = append(t, taskLibrary.EmptyTask{})
			t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("message %d", i)})
			t = append(t, taskLibrary.EmptyTask{})
			t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("Goodbye %d", i)})
			t = append(t, taskLibrary.LogTask{Message: fmt.Sprintf("Done %d", i), Level: slog.LevelInfo})
			j := job.New(uuid.New(), "sequenceJob", schedule, t, job.WithRunLimit(1))
			if err = c.Add(j); err != nil {
				panic(err)
			}
		}
	}(wg)
	wg.Wait()
	fmt.Println(c.Statistics(), o.Statistics())
	time.Sleep(2 * time.Second)
	fmt.Println("STARTING")
	o.Start(ctx)
	time.Sleep(20 * time.Second)
	o.Stop()
	fmt.Println("FINAL STATS")
	fmt.Println(c.Statistics(), o.Statistics())
}
