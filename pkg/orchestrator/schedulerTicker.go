package orchestrator

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/cron"
)

func newSchedulerTicker(uuid uuid.UUID, schedule cron.Schedule) *schedulerTicker {
	return &schedulerTicker{
		Uuid:     uuid,
		schedule: schedule,
		chTime:   make(chan time.Time),
	}
}

type schedulerTicker struct {
	Uuid         uuid.UUID
	schedule     cron.Schedule
	chTime       chan time.Time
	ticker       *cron.Ticker
	tickerCancel context.CancelFunc
	mux          sync.Mutex
}

// TODO Start return error?
func (s *schedulerTicker) Start(ctx context.Context, chTick chan SchedulerTick) {
	go s.tick(ctx, chTick)

	// Wait until the ticker has started
	for {
		if s.isRunning() {
			return
		}
	}
}

func (s *schedulerTicker) Stop() {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.tickerCancel != nil {
		s.tickerCancel()
	}
}

func (s *schedulerTicker) isRunning() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.ticker != nil {
		return s.ticker.IsRunning()
	}
	return false
}

// Listen on the scheduler channel chTime for triggers from the tickers.
// When a time tick is received, create a scheduler SchedulerTick and forward it
func (s *schedulerTicker) tick(ctx context.Context, chTick chan SchedulerTick) {
	s.mux.Lock()
	var tickerCtx context.Context
	tickerCtx, s.tickerCancel = context.WithCancel(ctx)
	s.ticker = cron.NewTicker(s.schedule, s.chTime)
	err := s.ticker.Start(tickerCtx)
	s.mux.Unlock()

	if err != nil {
		return
	}

	for {
		select {
		case <-tickerCtx.Done():
			return
		case t := <-s.chTime:
			chTick <- SchedulerTick{
				uuid: s.Uuid,
				time: t,
			}
		}
	}
}
