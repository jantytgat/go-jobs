package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"


	"github.com/jantytgat/go-jobs/pkg/cron"
	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/orchestrator"
	"github.com/jantytgat/go-jobs/pkg/task"
	"github.com/jantytgat/go-jobs/pkg/taskLibrary"
)

func main() {

	var err error
	var o *orchestrator.Orchestrator

	reg := prometheus.NewRegistry()
	// Add go runtime metrics and process collectors.
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
	)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		fmt.Println(http.ListenAndServe("localhost:6061", mux))
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	maxWorkers := runtime.NumCPU()
	maxJobs := runtime.NumCPU() * 2

	// maxJobs = 10
	if o, err = orchestrator.New(logger, "example", maxJobs, orchestrator.WithPrometheusRegistry(reg)); err != nil {
		panic(err)
	}
	if err = o.Handlers.RegisterHandlerPools([]*task.HandlerPool{
		task.NewHandlerPool(ctx, taskLibrary.EmptyTaskHandler(5*time.Second), maxWorkers*2, task.WithHandlerPoolPrometheusRegister(reg), task.WithHandlerPoolRecycling(200)),
		task.NewHandlerPool(ctx, taskLibrary.LogTaskHandler(5*time.Second), maxWorkers*3, task.WithHandlerPoolPrometheusRegister(reg), task.WithHandlerPoolRecycling(300)),
		task.NewHandlerPool(ctx, taskLibrary.EmptyErrorTaskHandler(5*time.Second), maxWorkers*4, task.WithHandlerPoolPrometheusRegister(reg), task.WithHandlerPoolRecycling(400)),

	}); err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
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
			t = append(t, taskLibrary.LogTask{Message: fmt.Sprintf("Hello %d", i), Level: slog.LevelDebug})
			t = append(t, taskLibrary.EmptyTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("this %d", i)})
			t = append(t, taskLibrary.EmptyErrorTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("is %d", i)})
			// t = append(t, taskLibrary.EmptyTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("my %d", i)})
			// t = append(t, taskLibrary.EmptyTask{})
			// t = append(t, taskLibrary.PrintTask{Message: fmt.Sprintf("message %d", i)})
			// t = append(t, taskLibrary.EmptyTask{})
			t = append(t, taskLibrary.LogTask{Message: fmt.Sprintf("Goodbye %d", i), Level: slog.LevelDebug})
			var j job.Job
			if i%2 == 0 {
				j = job.New(uuid.New(), fmt.Sprintf("%s-%d", "sequenceJob", i), schedule, t, job.WithRunLimit(20))
			} else if i%3 == 0 {
				j = job.New(uuid.New(), fmt.Sprintf("%s-%d", "sequenceJob", i), schedule, t, job.WithRunLimit(30))

			} else {
				j = job.New(uuid.New(), fmt.Sprintf("%s-%d", "sequenceJob", i), schedule, t)
			}

			if err = o.Catalog.Add(j); err != nil {
				panic(err)
			}
		}
	}(wg)
	wg.Wait()
	go func(ctx context.Context) {
		i := 0
		for {
			select {
			case <-ctx.Done():
			default:
				fmt.Println(o.Catalog.Statistics(), o.Statistics())
				err = prometheus.WriteToTextfile(fmt.Sprintf("%s_%d.prom", "test", i), reg)
				if err != nil {
					fmt.Printf("failed to write prometheus metrics: %v\n", err)
				}
				i++
				time.Sleep(1 * time.Second)
			}
		}

	}(ctx)

	fmt.Println("STARTING")
	_ = o.Start(ctx)
	time.Sleep(60 * time.Second)

	fmt.Println("STOPPING")
	o.Stop()
	cancel()
	time.Sleep(1 * time.Second)
	fmt.Println("FINAL STATS")
	fmt.Println(o.Catalog.Statistics(), o.Statistics())

	err = prometheus.WriteToTextfile("test.prom", reg)
	if err != nil {
		fmt.Printf("failed to write prometheus metrics: %v\n", err)
	}
}
