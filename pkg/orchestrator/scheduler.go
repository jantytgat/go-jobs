package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/cron"
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

func (s *scheduler) addTicker(uuid uuid.UUID, schedule cron.Schedule) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.tickers[uuid] = newSchedulerTicker(uuid, schedule)
}

func (s *scheduler) deleteTicker(uuid uuid.UUID) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.tickers, uuid)
}

func (s *scheduler) handleUpdate(u schedulerMessage) {
	// Only add new ticker if the uuid does not exist and the ticker must be enabled.
	// If the ticker does not exist and it should be disabled, exit the function as no further action is required.
	if !s.tickerExists(u.uuid) && u.enabled {
		s.addTicker(u.uuid, u.schedule)
	} else if !u.enabled {
		return
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
		s.deleteTicker(u.uuid)
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
