package cron

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const tickerInterval = 1 * time.Second

// NewTicker returns a cron Ticker based on the input schedule s.
// Ticket will send the time to channel chTrigger when the schedule is due.
func NewTicker(ctx context.Context, s Schedule, chTrigger chan<- time.Time) *Ticker {
	t := &Ticker{
		ctx:       ctx,
		schedule:  s,
		chTrigger: chTrigger,
	}
	return t
}

// Ticker creates a cron ticker that will send the current time to the output channel when the schedule is due.
// It can be controlled using the Start() and Stop() functions, or by cancelling the parent context.
type Ticker struct {
	ctx          context.Context
	tickerCancel context.CancelFunc

	schedule  Schedule
	chTrigger chan<- time.Time

	mux sync.RWMutex
}

// Start the cron ticker.
func (t *Ticker) Start() error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.tickerCancel != nil { // if tickerCancel is not nil, the tick goroutine is already running
		return fmt.Errorf("ticker already started")
	}

	tickerCtx, tickerCancel := context.WithCancel(t.ctx)
	t.tickerCancel = tickerCancel

	go t.monitor()
	go tick(tickerCtx, t.schedule, t.chTrigger)
	return nil
}

// Stop the cron ticker.
func (t *Ticker) Stop() error {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.tickerCancel == nil { // if tickerCancel is nil, Start() has never been called
		return fmt.Errorf("ticker already stopped")
	}

	t.tickerCancel()
	t.tickerCancel = nil // Reset tickerCancel so the goroutine can be started again
	return nil
}

func (t *Ticker) monitor() {
	for {
		select {
		case <-t.ctx.Done():
			_ = t.Stop()
			return
		}
	}
}

// tick is the function called by Start to initiate the goroutine
func tick(ctx context.Context, s Schedule, chTrigger chan<- time.Time) {
	ticker := time.NewTicker(tickerInterval)
	var exit bool
	for {
		if exit {
			break
		}

		select {
		case <-ctx.Done():
			exit = true
		case trigger := <-ticker.C:
			if s.IsDue(trigger) {
				chTrigger <- trigger
			}
		}
	}
}
