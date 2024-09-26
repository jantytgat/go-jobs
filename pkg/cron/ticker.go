package cron

import (
	"context"
	"sync"
	"time"
)

// NewTicker creates a cron ticker that will send the current time to the output channel when the schedule is due.
// It can be controlled using the Start() and Stop() functions, or by cancelling the parent context
func NewTicker(ctx context.Context, s Schedule, chOut chan<- time.Time) *Ticker {
	interval := 1 * time.Second
	t := &Ticker{
		ctx:      ctx,
		schedule: s,
		interval: interval,
		chOut:    chOut,
	}
	return t
}

type Ticker struct {
	ctx    context.Context
	cancel context.CancelFunc

	schedule Schedule
	interval time.Duration
	ticker   *time.Ticker

	chOut chan<- time.Time

	mux sync.RWMutex
}

func (t *Ticker) Start() {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.cancel != nil {
		return
	}
	tickerCtx, tickerCancel := context.WithCancel(t.ctx)
	t.cancel = tickerCancel
	t.ticker = time.NewTicker(t.interval)
	go t.tick(tickerCtx)
}

func (t *Ticker) Stop() {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if t.cancel == nil {
		return
	}
	t.cancel()
}

func (t *Ticker) tick(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			t.mux.Lock()
			t.ticker.Stop()
			t.ticker = nil
			t.cancel = nil
			t.mux.Unlock()
			return
		case trigger := <-t.ticker.C:
			if t.schedule.IsDue(trigger) {
				t.chOut <- trigger
			}
		}
	}
}
