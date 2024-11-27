package orchestrator

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/task"
)

func New(logger *slog.Logger, name string, maxRunners int, opts ...Option) (*Orchestrator, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}
	logger = logger.WithGroup("orchestrator")

	chScheduler := make(chan schedulerMessage, maxRunners)
	chTick := make(chan schedulerTick, maxRunners)
	chResults := make(chan job.Result, maxRunners)
	chDispatcher := make(chan dispatcherMessage, maxRunners)

	if name == "" {
		name = "orchestrator"
	}

	o := &Orchestrator{
		name:         name,
		scheduler:    newScheduler(logger, chScheduler, chTick),
		dispatcher:   newDispatcher(logger, maxRunners, chDispatcher, chResults),
		chScheduler:  chScheduler,
		chDispatcher: chDispatcher,
		chTick:       chTick,
		chResults:    chResults,
		logger:       logger,
		maxRunners:   maxRunners,
		mux:          sync.Mutex{},
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.reg == nil {
		o.reg = prometheus.NewRegistry()
	}

	if o.Catalog == nil {
		o.Catalog = job.NewMemoryCatalog()
	}

	if o.Handlers == nil {
		o.Handlers = task.NewHandlerRepository(name, task.WithHandlerRepositoryPrometheusRegister(o.reg))
	}

	if o.queue == nil {
		o.queue = NewMemoryQueue()
	}

	return o, nil
}

type Orchestrator struct {
	name         string
	ctx          context.Context
	cancelFunc   context.CancelFunc
	scheduler    *scheduler  // manages tickers for job schedule
	queue        Queue       // jobs to be queued for execution
	dispatcher   *dispatcher // manages job runners
	logger       *slog.Logger
	Catalog      job.Catalog             // contains jobs
	Handlers     *task.HandlerRepository // contains task handlers
	chScheduler  chan schedulerMessage   // channel to send updates to the scheduler
	chDispatcher chan dispatcherMessage  // channel to send jobs to dispatcher
	chResults    chan job.Result         // channel to get results from dispatcher
	chTick       chan schedulerTick      // channel to receive ticks from scheduler
	maxRunners   int
	reg          prometheus.Registerer
	mux          sync.Mutex
}

func (o *Orchestrator) Start(ctx context.Context) error {
	o.mux.Lock()
	defer o.mux.Unlock()

	var oCtx context.Context
	oCtx, o.cancelFunc = context.WithCancel(ctx)

	var err error
	if err = o.dispatcher.Start(oCtx); err != nil {
		return err
	}
	if err = o.scheduler.Start(oCtx); err != nil {
		return err
	}

	if o.scheduler.IsRunning() {
		go o.resultHandler(oCtx)
		go o.listenToQueue(oCtx)
		go o.schedulerMessenger(oCtx)
		go o.listenToTickers(oCtx)
	}
	return nil
}

func (o *Orchestrator) Statistics() Statistics {
	o.mux.Lock()
	defer o.mux.Unlock()

	handlerPoolStats := o.Handlers.Statistics()
	queueLength := o.queue.Length()

	return Statistics{
		HandlerPoolStatistics: handlerPoolStats,
		QueueLength:           queueLength,
	}
}

func (o *Orchestrator) Stop() {
	o.mux.Lock()
	defer o.mux.Unlock()
	if o.cancelFunc != nil {
		o.cancelFunc()
	}
}

func (o *Orchestrator) listenToQueue(ctx context.Context) {
	o.logger.LogAttrs(ctx, slog.LevelDebug, "starting queue listener")
	defer o.logger.LogAttrs(ctx, slog.LevelDebug, "stopping queue listener")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var t schedulerTick
			var err error
			if t, err = o.queue.Pop(); err != nil {
				// TODO add custom error type to handle different events?
				// o.logger.LogAttrs(o.ctx, slog.LevelDebug, "no jobs in queue")
				// time.Sleep(100 * time.Millisecond)
				break
			}

			go o.dispatchJob(ctx, t)
		}
	}
}

func (o *Orchestrator) listenToTickers(ctx context.Context) {
	o.logger.LogAttrs(ctx, slog.LevelDebug, "starting ticker listener")
	defer o.logger.LogAttrs(ctx, slog.LevelDebug, "stopping ticker listener")

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-o.chTick:
			go o.queue.Push(t)
		}
	}
}

func (o *Orchestrator) dispatchJob(ctx context.Context, tick schedulerTick) {
	o.logger.LogAttrs(ctx, slog.LevelDebug, "dispatching job", slog.Group("job", slog.String("id", tick.uuid.String()), slog.String("time", tick.time.String())))
	var exit bool
	var err error
	var retries int
	maxRetries := 5

	for {
		if exit {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			if retries < maxRetries {
				var j job.Job
				j, err = o.Catalog.Get(tick.uuid)
				if err != nil {
					retries++
					o.logger.LogAttrs(o.ctx, slog.LevelError, "failed to get job for dispatcher", slog.String("job", tick.uuid.String()), slog.String("error", err.Error()))
					time.Sleep(1 * time.Second) // back off from catalog before retrying
					break
				}

				o.chDispatcher <- dispatcherMessage{
					job:               j,
					handlerRepository: o.Handlers,
					trigger:           tick.time,
				}
			}
			exit = true
		}
	}
}

func (o *Orchestrator) schedulerMessenger(ctx context.Context) {
	o.logger.LogAttrs(ctx, slog.LevelDebug, "starting scheduler messenger")
	defer o.logger.LogAttrs(ctx, slog.LevelDebug, "stopping scheduler messenger")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if o.Catalog.Count() == 0 {
				o.logger.LogAttrs(o.ctx, slog.LevelDebug, "no jobs to send to scheduler")
				// time.Sleep(1 * time.Second) // set to the smallest interval the cron package can handle
				break
			}

			jobs := o.Catalog.All()
			for _, j := range jobs {
				if !j.LimitRuns || (j.LimitRuns && o.Catalog.CountResults(j.Uuid) < j.MaxRuns) {
					go func() {
						o.chScheduler <- schedulerMessage{
							uuid:     j.Uuid,
							enabled:  j.Enabled,
							schedule: j.Schedule,
						}
					}()
				} else if j.LimitRuns && o.Catalog.CountResults(j.Uuid) == j.MaxRuns {
					go func() {
						o.chScheduler <- schedulerMessage{
							uuid:     j.Uuid,
							enabled:  false,
							schedule: j.Schedule,
						}
					}()
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (o *Orchestrator) resultHandler(ctx context.Context) {
	o.logger.LogAttrs(ctx, slog.LevelDebug, "starting result handler")
	defer o.logger.LogAttrs(ctx, slog.LevelDebug, "stopping result handler")
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-o.chResults:
			go o.Catalog.AddResult(r) // make sure results are read from the channel as fast as possible
		}
	}
}
