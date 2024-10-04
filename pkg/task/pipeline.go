package task

import (
	"log/slog"
	"sync"
)

func NewPipeline(l *slog.Logger) *Pipeline {
	return &Pipeline{
		logger: l,
		data:   make(map[string]interface{}),
		errors: make([]error, 0),
	}
}

type Pipeline struct {
	logger *slog.Logger
	data   map[string]interface{}
	errors []error
	mux    sync.RWMutex
}

func (p *Pipeline) Data() map[string]interface{} {
	p.mux.RLock()
	defer p.mux.RUnlock()

	output := make(map[string]interface{}, len(p.data))
	for k, v := range p.data {
		output[k] = v
	}
	return output
}

func (p *Pipeline) Errors() []error {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.errors
}

func (p *Pipeline) Get(key string) interface{} {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.data[key]
}

func (p *Pipeline) Keys() []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

func (p *Pipeline) Logger(t Task) *slog.Logger {
	return p.logger.With(LogTaskAttr(t))
}

func (p *Pipeline) Set(key string, value interface{}) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.data[key] = value
}
