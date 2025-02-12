package task

import (
	"context"
	"runtime"
	"sync"
	"time"
)

func NewHandlerPool(ctx context.Context, h Handler, maxWorkers int, opts ...HandlerPoolOption) *HandlerPool {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	p := &HandlerPool{
		handler:       h,
		maxWorkers:    maxWorkers,
		chWorkerInput: make(chan HandlerTask),
		ChPoolInput:   make(chan HandlerTask, maxWorkers),
	}

	for _, opt := range opts {
		opt(p)
	}

	p.ChPoolInput = make(chan HandlerTask, p.maxWorkers)

	go p.listen(ctx)
	return p
}

type HandlerPool struct {
	maxWorkers     int
	workers        int
	recycleWorkers bool
	recycleAfter   int
	handler        Handler
	ChPoolInput    chan HandlerTask
	chWorkerInput  chan HandlerTask
	mux            sync.RWMutex
}

func (p *HandlerPool) IsRunning() bool {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.workers == 0 {
		return false
	}
	return true
}

func (p *HandlerPool) Name() string {
	return p.handler.Name
}

func (p *HandlerPool) Statistics() HandlerPoolStatistics {
	activeWorkers := GetMetricValue(handlerPoolMetrics.activeWorkers.WithLabelValues(p.handler.Name))
	workers := GetMetricValue(handlerPoolMetrics.workers.WithLabelValues(p.handler.Name))
	idleWorkers := workers - activeWorkers

	return HandlerPoolStatistics{
		ActiveWorkers:                activeWorkers,
		IdleWorkers:                  idleWorkers,
		Workers:                      workers,
		MaxWorkers:                   GetMetricValue(handlerPoolMetrics.maxWorkers.WithLabelValues(p.handler.Name)),
		RecycledWorkers:              GetMetricValue(handlerPoolMetrics.recycledWorkers.WithLabelValues(p.handler.Name)),
		TasksIngested:                GetMetricValue(handlerPoolMetrics.tasksIngested.WithLabelValues(p.handler.Name)),
		TasksProcessedStatusNone:     GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusNone.String())),
		TasksProcessedStatusPending:  GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusPending.String())),
		TasksProcessedStatusSuccess:  GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusSuccess.String())),
		TasksProcessedStatusCanceled: GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusCanceled.String())),
		TasksProcessedStatusError:    GetMetricValue(handlerPoolMetrics.tasksProcessed.WithLabelValues(p.handler.Name, StatusError.String())),
		TasksWaiting:                 GetMetricValue(handlerPoolMetrics.tasksWaiting.WithLabelValues(p.handler.Name)),
	}
}

func (p *HandlerPool) decreaseActiveWorkerCount() {
	handlerPoolMetrics.activeWorkers.WithLabelValues(p.handler.Name).Dec()
}

func (p *HandlerPool) decreaseTasksWaiting() {
	handlerPoolMetrics.tasksWaiting.WithLabelValues(p.handler.Name).Dec()
}

func (p *HandlerPool) decreaseWorkerCount(recycled bool) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.workers--
	handlerPoolMetrics.workers.WithLabelValues(p.handler.Name).Dec()

	if recycled {
		handlerPoolMetrics.recycledWorkers.WithLabelValues(p.handler.Name).Inc()
	}
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

func (p *HandlerPool) launchWorkers(ctx context.Context) {
	handlerPoolMetrics.maxWorkers.WithLabelValues(p.handler.Name).Set(float64(p.maxWorkers))
	for {
		p.mux.Lock()
		if p.workers < p.maxWorkers {
			for i := 0; i < p.maxWorkers-p.workers; i++ {
				p.workers++
				handlerPoolMetrics.workers.WithLabelValues(p.handler.Name).Inc()
				go p.runWorker(ctx)
			}
		}
		p.mux.Unlock()
	}
}

func (p *HandlerPool) listen(ctx context.Context) {
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// Make sure all workers have started before accepting tasks
	go p.launchWorkers(workerCtx)

	// Start listening for handler pool input messages until the listen context is canceled or the input channel is closed.
Listen:
	for {
		select {
		case <-ctx.Done():
			break Listen
		case ht, ok := <-p.ChPoolInput:
			if !ok {
				break Listen
			}
      
			p.sendToWorker(ht)
		}
	}

	// Get the current length of the input channel, so it can be drained properly
	drainQueue := len(p.ChPoolInput)

	// Start draining the input channel
Drain:
	for i := 0; i < drainQueue; i++ {
		ht, ok := <-p.ChPoolInput
		if !ok {
			break Drain
		}
		p.sendToWorker(ht)
	}

	// When all messages from the input channel have been sent to a worker, close the worker input channel.
	// This will stop idle workers.
	close(p.chWorkerInput)

	// After draining the queue, wait for all workers to finish gracefully.
	// This will never take longer than the maximum timeout value of the handler, as the worker function will time out earlier,
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), p.handler.timeout)
	defer cleanupCancel()

Cleanup:
	for {
		select {
		case <-cleanupCtx.Done():
			break Cleanup
		default:
			if !p.IsRunning() {
				break Cleanup
			} else {
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
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
	var recycle bool
	tasksExecuted := 0

Run:
	for {
		if p.recycleWorkers && tasksExecuted == p.recycleAfter {
			recycle = true
			break Run
		}

		select {
		case <-ctx.Done():
			break Run
		case t, ok := <-p.chWorkerInput:
			if !ok {
				break Run
			}

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
			tasksExecuted++
		}
	}

	p.decreaseWorkerCount(recycle)
}

func (p *HandlerPool) sendToWorker(ht HandlerTask) {
	p.increaseTasksIngestedCount()
	p.increaseTasksWaiting()
	p.chWorkerInput <- ht
	p.decreaseTasksWaiting()
}
