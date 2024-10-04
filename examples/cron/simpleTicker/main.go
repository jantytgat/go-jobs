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

	schedule := cron.EverySecond()
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

	// start ticker
	ticker.Start()

	// stop ticker after 5 seconds
	time.Sleep(5 * time.Second)
	ticker.Stop()

	// Restart ticker after 2 seconds
	time.Sleep(2 * time.Second)
	ticker.Start()

	// Cancel the root context after 10 seconds
	time.Sleep(10 * time.Second)
	cancel()
}
