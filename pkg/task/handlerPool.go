package task

import (
	"context"
	"runtime"
	"sync"
)

func NewHandlerPool(ctx context.Context, h Handler, maxWorkers int, opts ...HandlerPoolOption) *HandlerPool {
	workerCtx, workerCancel := context.WithCancel(ctx)

	p := &HandlerPool{
		ctx:           ctx,
		workerCtx:     workerCtx,
		workerCancel:  workerCancel,
		handler:       h,
		maxWorkers:    setMaxWorkers(maxWorkers),
		chWorkerInput: make(chan HandlerTask),
	}

	for _, opt := range opts {
		opt(p)
	}
	p.ChPoolInput = make(chan HandlerTask, p.maxWorkers)

	go p.listen()
	return p
}

type HandlerPool struct {
	ctx          context.Context
	workerCtx    context.Context
	workerCancel context.CancelFunc

	maxWorkers int
	workers    int
	handler    Handler

	ChPoolInput   chan HandlerTask
	chWorkerInput chan HandlerTask

	mux sync.RWMutex
}

func (p *HandlerPool) Name() string {
	return p.handler.Name
}

func (p *HandlerPool) Statistics() HandlerPoolStatistics {
	workers := GetMetricValue(handlerPoolMetrics.workers.WithLabelValues(p.handler.Name))
	activeWorkers := GetMetricValue(handlerPoolMetrics.activeWorkers.WithLabelValues(p.handler.Name))
	idleWorkers := workers - activeWorkers
	maxWorkers := GetMetricValue(handlerPoolMetrics.maxWorkers.WithLabelValues(p.handler.Name))
	tasksIngested := GetMetricValue(handlerPoolMetrics.tasksIngested.WithLabelValues(p.handler.Name))
	tasksProcessedNone := GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusNone.String()))
	tasksProcessedPending := GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusPending.String()))
	tasksProcessedSuccess := GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusSuccess.String()))
	tasksProcessedCanceled := GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusCanceled.String()))
	tasksProcessedError := GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusError.String()))
	tasksWaiting := GetMetricValue(handlerPoolMetrics.tasksWaiting.WithLabelValues(p.handler.Name))
	return HandlerPoolStatistics{
		ActiveWorkers:                activeWorkers,
		IdleWorkers:                  idleWorkers,
		Workers:                      workers,
		MaxWorkers:                   maxWorkers,
		TasksIngested:                tasksIngested,
		TasksProcessedStatusNone:     tasksProcessedNone,
		TasksProcessedStatusPending:  tasksProcessedPending,
		TasksProcessedStatusSuccess:  tasksProcessedSuccess,
		TasksProcessedStatusCanceled: tasksProcessedCanceled,
		TasksProcessedStatusError:    tasksProcessedError,
		TasksWaiting:                 tasksWaiting,
	}
}
func (p *HandlerPool) decreaseActiveWorkerCount() {
	handlerPoolMetrics.activeWorkers.WithLabelValues(p.handler.Name).Dec()
}

func (p *HandlerPool) decreaseWorkerCount() {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.workers--
	handlerPoolMetrics.workers.WithLabelValues(p.handler.Name).Dec()
}

func (p *HandlerPool) decreaseTasksWaiting() {
	handlerPoolMetrics.tasksWaiting.WithLabelValues(p.handler.Name).Dec()
}

func (p *HandlerPool) increaseActiveWorkerCount() {
	handlerPoolMetrics.activeWorkers.WithLabelValues(p.handler.Name).Inc()
}

func (p *HandlerPool) increaseTasksIngestedCount() {
	handlerPoolMetrics.tasksIngested.WithLabelValues(p.handler.Name).Inc()
}

func (p *HandlerPool) increaseTasksProcessedCount(status Status) {
	handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, status.String()).Inc()
}

func (p *HandlerPool) increaseTasksWaiting() {
	handlerPoolMetrics.tasksWaiting.WithLabelValues(p.handler.Name).Inc()
}

func (p *HandlerPool) listen() {
	// Make sure all workers have started before accepting tasks
	p.launchWorkers()

	var exit bool
	for {
		if exit {
			break
		}

		select {
		case <-p.ctx.Done():
			exit = true
		case ht, ok := <-p.ChPoolInput:
			if !ok {
				exit = true
			}
			p.increaseTasksIngestedCount()
			p.increaseTasksWaiting()
			p.chWorkerInput <- ht
			p.decreaseTasksWaiting()
		default:
			p.launchWorkers()
		}
	}
	p.workerCancel()
}

func (p *HandlerPool) launchWorkers() {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.workers < p.maxWorkers {
		for i := 0; i < p.maxWorkers-p.workers; i++ {
			p.workers++
			handlerPoolMetrics.workers.WithLabelValues(p.handler.Name).Inc()
			go p.runWorker(p.workerCtx)
		}
	}
}

func (p *HandlerPool) runWorker(ctx context.Context) {
	defer p.decreaseWorkerCount()
	var exit bool
	for {
		if exit {
			break
		}

		select {
		case <-ctx.Done():
			exit = true
		case t := <-p.chWorkerInput:
			p.increaseActiveWorkerCount()

			status, err := p.handler.Execute(ctx, t.Task, t.Pipeline)
			if t.ChResult != nil {
				t.ChResult <- HandlerResult{
					Task:   t.Task,
					Status: status,
					Error:  err,
				}
			}

			p.increaseTasksProcessedCount(status)
			p.decreaseActiveWorkerCount()
		}
	}
}

func setMaxWorkers(maxWorkers int) int {
	if maxWorkers > 0 {
		return maxWorkers
	}
	return runtime.NumCPU()
}
