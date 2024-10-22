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
func NewTicker(s Schedule, chTrigger chan<- time.Time) *Ticker {
	t := &Ticker{
		schedule:  s,
		chTrigger: chTrigger,
	}
	return t
}

// Ticker creates a cron ticker that will send the current time to the output channel when the schedule is due.
// It can be controlled using the start() and stop() functions, or by cancelling the parent context.
type Ticker struct {
	tickerCancel context.CancelFunc

	schedule  Schedule
	chTrigger chan<- time.Time

	mux sync.Mutex
}

func (t *Ticker) IsRunning() bool {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.tickerCancel != nil {
		return true
	}
	return false
}

// Start the cron ticker.
func (t *Ticker) Start(ctx context.Context) error {
	if t.IsRunning() {
		return fmt.Errorf("ticker already started")
	}

	t.mux.Lock()
	defer t.mux.Unlock()

	var tickerCtx context.Context
	tickerCtx, t.tickerCancel = context.WithCancel(ctx)
	go t.tick(tickerCtx, t.schedule, t.chTrigger)

	return nil
}

// Stop the cron ticker.
func (t *Ticker) Stop() error {
	if !t.IsRunning() {
		return fmt.Errorf("ticker already stopped")
	}

	t.mux.Lock()
	defer t.mux.Unlock()
	t.tickerCancel()
	return nil
}

func (t *Ticker) resetCancelFunc() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.tickerCancel = nil
}

// tick is the function called by start to initiate the goroutine
func (t *Ticker) tick(ctx context.Context, s Schedule, chTrigger chan<- time.Time) {
	ticker := time.NewTicker(tickerInterval)
	var exit bool
	for {
		if exit {
			break
		}

		select {
		case <-ctx.Done():
			ticker.Stop()
			exit = true
		case trigger := <-ticker.C:
			if s.IsDue(trigger) {
				chTrigger <- trigger
			}
		}
	}
	t.resetCancelFunc()
}
