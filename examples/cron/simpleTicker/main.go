package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jantytgat/go-jobs/pkg/cron"
)

func main() {
	schedule := cron.EverySecond()
	chOut := make(chan time.Time)
	ticker := cron.NewTicker(schedule, chOut)

	ctx, cancel := context.WithCancel(context.Background())
	stop := context.AfterFunc(ctx, func() {
		fmt.Println("Closing channel")
		time.Sleep(1 * time.Second)
		close(chOut)
	})
	defer stop()
	defer cancel()

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
	fmt.Println("Starting ticker")
	_ = ticker.Start(ctx)

	// stop ticker after 5 seconds
	fmt.Println("Stopping ticker after 2 seconds")
	time.Sleep(2 * time.Second)
	_ = ticker.Stop()

	// try stopping ticker second time
	fmt.Println("Stopping ticker a second time after 1 second")
	time.Sleep(1 * time.Second)
	_ = ticker.Stop()

	// Restart ticker after 2 seconds
	fmt.Println("Starting ticker again in 2 seconds")
	time.Sleep(2 * time.Second)
	_ = ticker.Start(ctx)

	// Cancel the root context after 5 seconds
	time.Sleep(5 * time.Second)
}
