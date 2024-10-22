package task

import (
	"fmt"
	"log/slog"
	"sync"
)

var pipelineDataFields = make([]string, 0)

func RegisterPipelineDataField(s string) error {
	for _, v := range pipelineDataFields {
		if v == s {
			return fmt.Errorf("pipeline data field %s already registered", s)
		}
	}
	pipelineDataFields = append(pipelineDataFields, s)
	return nil
}

func RegisterPipelineDataFields(s []string) error {
	for _, v := range s {
		if err := RegisterPipelineDataField(v); err != nil {
			return err
		}
	}
	return nil
}

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

func (p *Pipeline) Get(key string) (interface{}, error) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	var v interface{}
	var found bool
	if v, found = p.data[key]; !found {
		return nil, fmt.Errorf("pipeline key %s not found", key)
	}
	return v, nil
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
