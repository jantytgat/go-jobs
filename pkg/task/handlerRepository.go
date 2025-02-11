package task

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

func NewHandlerRepository(name string, opts ...HandlerRepositoryOption) *HandlerRepository {
	r := &HandlerRepository{
		name:         name,
		handlerPools: make(map[string]*HandlerPool),
		mux:          sync.RWMutex{},
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.reg == nil {
		r.reg = prometheus.NewRegistry()
	}
	return r
}

type HandlerRepository struct {
	name         string
	handlerPools map[string]*HandlerPool
	reg          prometheus.Registerer
	mux          sync.RWMutex
}

func (r *HandlerRepository) Execute(ctx context.Context, t HandlerTask) error {
	handlerPool, err := r.get(t.Task.Name())

	if err != nil {
		// Try to register the default handler pool for task, return on error
		if err = r.registerHandlerPool(t.Task.DefaultHandlerPool(ctx)); err != nil {
			return err
		}

		// Second attempt to get the handler pool for the task, return on error
		if handlerPool, err = r.get(t.Task.Name()); err != nil {
			return err
		}
	}

	// Send the task to the handler pool channel
	handlerPool.ChPoolInput <- t
	return nil
}

func (r *HandlerRepository) RegisterHandlerPools(p []*HandlerPool) error {
	for _, handlerPool := range p {
		if err := r.registerHandlerPool(handlerPool); err != nil {
			return err
		}
	}
	return nil
}

func (r *HandlerRepository) Statistics() map[string]HandlerPoolStatistics {
	r.mux.RLock()
	defer r.mux.RUnlock()

	stats := make(map[string]HandlerPoolStatistics)
	for k, v := range r.handlerPools {
		stats[k] = v.Statistics()
	}
	return stats
}

func (r *HandlerRepository) get(name string) (*HandlerPool, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	pool, found := r.handlerPools[name]
	if !found {
		return nil, fmt.Errorf("no handler pool found for %s", name)
	}
	return pool, nil
}

func (r *HandlerRepository) registerHandlerPool(p *HandlerPool) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if p == nil {
		return fmt.Errorf("handler pool is nil")
	}

	_, found := r.handlerPools[p.Name()]
	if !found {
		r.handlerPools[p.Name()] = p
		handlerRepositoryMetrics.handlerPools.WithLabelValues(r.name).Inc()
		if err := handlerPoolMetrics.Register(r.reg); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
			return err
		}
	}

	return nil
}
