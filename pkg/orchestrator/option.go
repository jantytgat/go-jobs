package orchestrator

import (
	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/task"
)

type Option func(*Orchestrator)

func WithCatalog(catalog job.Catalog) Option {
	return func(o *Orchestrator) {
		o.Catalog = catalog
	}
}

func WithHandlerRepository(r *task.HandlerRepository) Option {
	return func(o *Orchestrator) {
		o.Handlers = r
	}
}

func WithQueue(q Queue) Option {
	return func(o *Orchestrator) {
		o.queue = q
	}
}
