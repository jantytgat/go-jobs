package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jantytgat/go-jobs/pkg/cron"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedule, err := cron.NewSchedule("*/2 * * * * *")
	if err != nil {
		panic(err)
	}
	chOut := make(chan time.Time)
	ticker := cron.NewTicker(ctx, schedule, chOut)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-chOut:
				fmt.Println(t)
			}
		}
	}()

	// Start ticker
	ticker.Start()
	time.Sleep(20*time.Second + 1)
	cancel()
}
