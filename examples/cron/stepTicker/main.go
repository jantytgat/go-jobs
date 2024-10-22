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

	maxTickers := 1000
	tickers := make([]*cron.Ticker, maxTickers)

	for i := 0; i < maxTickers; i++ {
		tickers[i] = cron.NewTicker(schedule, chOut)
	}

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
	for i := 0; i < maxTickers; i++ {
		_ = tickers[i].Start(ctx)
	}

	time.Sleep(20*time.Second + 1)
	cancel()
}
