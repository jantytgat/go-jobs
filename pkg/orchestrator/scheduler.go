package orchestrator

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/cron"
)

func newScheduler(ctx context.Context, chJob chan schedulerUpdate) (*scheduler, chan tick) {
	schedulerCtx, schedulerCancel := context.WithCancel(ctx)
	chTick := make(chan tick)
	s := &scheduler{
		ctx:      schedulerCtx,
		cancel:   schedulerCancel,
		chUpdate: chJob,
		chTick:   chTick,
	}
	s.tickerCtx, s.tickerCancel = context.WithCancel(schedulerCtx)
	go s.start(schedulerCtx)
	return s, chTick
}

type scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	chUpdate chan schedulerUpdate
	chTick   chan tick

	tickerCtx    context.Context
	tickerCancel context.CancelFunc

	tickers map[uuid.UUID]*cron.Ticker

	mux sync.RWMutex
}

func (s *scheduler) listen(ctx context.Context, uuid uuid.UUID, chTick chan time.Time) {
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-chTick:
			s.chTick <- tick{
				uuid: uuid,
				time: t,
			}
		}
	}
}

func (s *scheduler) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done(): // stop all tickers
			s.tickerCancel()
			return
		case u := <-s.chUpdate:
			var t *cron.Ticker
			var found bool

			// look up the ticker for the job uuid
			s.mux.RLock()
			t, found = s.tickers[u.uuid]
			s.mux.RUnlock()

			if !found {
				chTick := make(chan time.Time)
				go s.listen(ctx, u.uuid, chTick) // start goroutine to listen for ticker events

				s.mux.Lock()
				// add ticker to scheduler when it does not exist
				t = cron.NewTicker(s.tickerCtx, u.schedule, chTick)
				s.tickers[u.uuid] = t
				s.mux.Unlock()
			}

			// Update the ticker state
			switch u.enabled {
			case true:
				t.Start()
			case false:
				t.Stop()
			}
		}
	}
}
