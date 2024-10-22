package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

func newScheduler(chIn chan schedulerMessage, chOut chan schedulerTick) *scheduler {
	s := &scheduler{
		chIn:    chIn,
		chOut:   chOut,
		tickers: make(map[uuid.UUID]*schedulerTicker),
	}
	return s
}

type scheduler struct {
	chIn             chan schedulerMessage
	chOut            chan schedulerTick
	listenCtx        context.Context
	listenCancelFunc context.CancelFunc
	tickers          map[uuid.UUID]*schedulerTicker
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

func (s *scheduler) Start(ctx context.Context) {
	go s.listen(ctx)
	for {
		if s.IsRunning() {
			return
		}
	}
}

func (s *scheduler) Stop() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.listenCancelFunc == nil {
		return fmt.Errorf("scheduler has already been stopped")
	}
	s.listenCancelFunc()
	s.listenCancelFunc = nil
	return nil
}

func (s *scheduler) handleUpdate(u schedulerMessage) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.tickerExists(u.uuid) {
		s.tickers[u.uuid] = newSchedulerTicker(u.uuid, u.schedule)
	}

	switch u.enabled {
	case true:
		if !s.tickers[u.uuid].isRunning() {
			go s.tickers[u.uuid].Start(s.listenCtx, s.chOut)
		}
	case false:
		// Stop the individual ticker, then remove it from the map
		if s.tickers[u.uuid].isRunning() {
			s.tickers[u.uuid].Stop()
		}
		delete(s.tickers, u.uuid)
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

func (s *scheduler) tickerExists(uuid uuid.UUID) bool {
	var found bool
	_, found = s.tickers[uuid]
	return found
}
