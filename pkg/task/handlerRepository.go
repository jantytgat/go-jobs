package task

import (
	"context"
	"fmt"
	"sync"
)

func NewHandlerRepository() *HandlerRepository {
	return &HandlerRepository{
		handlerPools: make(map[string]*HandlerPool),
		mux:          sync.RWMutex{},
	}
}

type HandlerRepository struct {
	handlerPools map[string]*HandlerPool
	mux          sync.RWMutex
}

func (r *HandlerRepository) Execute(ctx context.Context, t HandlerTask) error {
	handlerPool, err := r.get(t.Task.Name())
	if err != nil {
		r.registerHandlerPool(t.Task.DefaultHandlerPool(ctx))
		if handlerPool, err = r.get(t.Task.Name()); err != nil {
			return err
		}
	}

	handlerPool.ChTasks <- t
	return nil
}

func (r *HandlerRepository) RegisterHandlerPools(p []*HandlerPool) {
	for _, handlerPool := range p {
		r.registerHandlerPool(handlerPool)
	}
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

func (r *HandlerRepository) registerHandlerPool(p *HandlerPool) {
	r.mux.Lock()
	defer r.mux.Unlock()

	_, found := r.handlerPools[p.Name()]
	if !found {
		r.handlerPools[p.Name()] = p
	}
}
