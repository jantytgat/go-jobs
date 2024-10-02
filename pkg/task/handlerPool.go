package task

import (
	"context"
	"runtime"
	"sync"
)

type HandlerPoolStatistics struct {
	ActiveWorkers  int
	IdleWorkers    int
	Workers        int
	MaxWorkers     int
	TasksProcessed int
}

func NewHandlerPool(ctx context.Context, h Handler, maxWorkers int) *HandlerPool {
	workerCtx, workerCancel := context.WithCancel(ctx)
	p := &HandlerPool{
		ctx:           ctx,
		workerCtx:     workerCtx,
		workerCancel:  workerCancel,
		handler:       h,
		maxWorkers:    setMaxWorkers(maxWorkers),
		chWorkerInput: make(chan HandlerTask),
	}

	p.ChHandlerTask = make(chan HandlerTask, p.maxWorkers)

	go p.listen()
	return p
}

type HandlerPool struct {
	ctx          context.Context
	workerCtx    context.Context
	workerCancel context.CancelFunc

	workers        int
	maxWorkers     int
	activeWorkers  int
	tasksProcessed int
	handler        Handler

	ChHandlerTask chan HandlerTask
	chWorkerInput chan HandlerTask

	mux sync.RWMutex
}

func (p *HandlerPool) Name() string {
	return p.handler.Name
}

func (p *HandlerPool) Statistics() HandlerPoolStatistics {
	p.mux.Lock()
	defer p.mux.Unlock()
	return HandlerPoolStatistics{
		ActiveWorkers:  p.activeWorkers,
		IdleWorkers:    p.workers - p.activeWorkers,
		Workers:        p.workers,
		MaxWorkers:     p.maxWorkers,
		TasksProcessed: p.tasksProcessed,
	}
}

func (p *HandlerPool) listen() {
	p.mux.Lock()
	for i := 0; i < p.maxWorkers; i++ {
		p.workers++
		go p.runWorker(p.workerCtx)
	}
	p.mux.Unlock()

	var exit bool
	for {
		if exit {
			break
		}

		select {
		case <-p.ctx.Done():
			exit = true
		case ht, ok := <-p.ChHandlerTask:
			if !ok {
				exit = true
			}
			p.chWorkerInput <- ht
		default:
			p.mux.Lock()
			if p.workers < p.maxWorkers {
				p.workers++
				go p.runWorker(p.workerCtx)
			}
			p.mux.Unlock()
		}
	}
	p.workerCancel()
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
			p.mux.Lock()
			p.activeWorkers++
			p.mux.Unlock()

			status, err := p.handler.Execute(ctx, t.Task, t.Pipeline)
			if t.ChResult != nil {
				t.ChResult <- HandlerTaskResult{
					Status: status,
					Error:  err,
				}
			}

			p.mux.Lock()
			p.tasksProcessed++
			p.activeWorkers--
			p.mux.Unlock()
		}
	}
}

func (p *HandlerPool) decreaseWorkerCount() {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.workers--
}

func setMaxWorkers(maxWorkers int) int {
	if maxWorkers > 0 {
		return maxWorkers
	}
	return runtime.NumCPU()
}
