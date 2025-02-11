package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/cron"
)

func newScheduler(logger *slog.Logger, chIn chan schedulerMessage, chOut chan SchedulerTick) *scheduler {
	s := &scheduler{
		chIn:    chIn,
		chOut:   chOut,
		tickers: make(map[uuid.UUID]*schedulerTicker),
		logger:  logger.WithGroup("scheduler"),
	}
	return s
}

type scheduler struct {
	chIn             chan schedulerMessage
	chOut            chan SchedulerTick
	listenCtx        context.Context
	listenCancelFunc context.CancelFunc
	tickers          map[uuid.UUID]*schedulerTicker
	logger           *slog.Logger
	mux              sync.Mutex
}

func (s *scheduler) IsRunning() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.listenCancelFunc != nil {
		return true
	}
	return false
}

func (s *scheduler) Start(ctx context.Context) error {
	go s.listen(ctx)
	startCtx, startCancel := context.WithTimeout(ctx, 1*time.Second)
	defer startCancel()
	for {
		select {
		case <-startCtx.Done():
			s.logger.LogAttrs(ctx, slog.LevelError, "scheduler start timeout")
			return fmt.Errorf("scheduler start timeout")
		default:
			if s.IsRunning() { // TODO add timeout to start before returning an error?
				s.logger.LogAttrs(ctx, slog.LevelDebug, "scheduler has started")
				return nil
			}
		}
	}
}

func (s *scheduler) Stop(ctx context.Context) {
	s.mux.Lock()
	if s.listenCancelFunc == nil {
		s.logger.LogAttrs(context.Background(), slog.LevelWarn, "scheduler has stopped already")
		return
	}
	s.mux.Unlock()

	s.logger.LogAttrs(ctx, slog.LevelDebug, "dispatcher stopping")
	s.listenCancelFunc()

	s.mux.Lock()
	s.listenCancelFunc = nil
	s.mux.Unlock()
}

func (s *scheduler) getTicker(uuid uuid.UUID) *schedulerTicker {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.tickers[uuid]
}

func (s *scheduler) handleUpdate(u schedulerMessage) {
	tickerExists := s.tickerExists(u.uuid)

	if !tickerExists {
		switch u.enabled {
		case true:
			s.startTicker(u.uuid, u.schedule)
			return
		case false:
			return
		}
	}

	// The ticker exists and must be disabled
	if !u.enabled {
		s.stopAndRemoveTicker(u.uuid)
		return
	}

	// The ticker exists but the schedule has changed
	ticker := s.getTicker(u.uuid)
	if ticker != nil && ticker.schedule.String() != u.schedule.String() {
		s.updateTicker(u.uuid, u.schedule)
		return
	}
}

func (s *scheduler) isRunning() bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.listenCancelFunc == nil {
		return false
	}
	return true
}

// Start the scheduler to listen for updates from schedulerTickers
// If the scheduleTicker cannot be found, add a new one to the map.
// For each ticker, the scheduler channel chOut is passed so the tickers can send a trigger to the orchestrator when
// an item must be queued.
func (s *scheduler) listen(ctx context.Context) {
	s.logger.LogAttrs(ctx, slog.LevelDebug, "scheduler starting")
	defer s.logger.LogAttrs(ctx, slog.LevelDebug, "scheduler stopped")
	s.mux.Lock()
	s.listenCtx, s.listenCancelFunc = context.WithCancel(ctx)
	s.mux.Unlock()

	for {
		select {
		case <-s.listenCtx.Done():
			// All tickers will be stopped as well as their context is based on s.listenCtx
			return
		case u := <-s.chIn:
			go s.handleUpdate(u)
		}
	}
}

func (s *scheduler) startTicker(uuid uuid.UUID, schedule cron.Schedule) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.logger.LogAttrs(s.listenCtx, slog.LevelDebug, "starting ticker", slog.Group("job", slog.String("id", uuid.String()), slog.String("schedule", schedule.String())))
	s.tickers[uuid] = newSchedulerTicker(uuid, schedule)
	s.tickers[uuid].Start(s.listenCtx, s.chOut)
}

func (s *scheduler) stopAndRemoveTicker(uuid uuid.UUID) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, found := s.tickers[uuid]; found {
		s.logger.LogAttrs(s.listenCtx, slog.LevelDebug, "stopping ticker", slog.Group("job", slog.String("id", uuid.String())))
		s.tickers[uuid].Stop()
		delete(s.tickers, uuid)
	}
}

func (s *scheduler) tickerExists(uuid uuid.UUID) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, found := s.tickers[uuid]; found {
		return true
	}
	return false
}

func (s *scheduler) updateTicker(uuid uuid.UUID, schedule cron.Schedule) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.tickers[uuid].Stop()
	s.tickers[uuid].schedule = schedule
	s.tickers[uuid].Start(s.listenCtx, s.chOut)
	s.logger.LogAttrs(s.listenCtx, slog.LevelDebug, "updated ticker", slog.Group("job", slog.String("id", uuid.String()), slog.String("schedule", s.tickers[uuid].schedule.String())))

}
