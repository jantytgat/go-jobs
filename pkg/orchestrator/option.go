package orchestrator

import (
	"log/slog"

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

func WithLogger(l *slog.Logger) Option {
	return func(o *Orchestrator) {
		o.logger = l
	}
}

func WithQueue(q Queue) Option {
	return func(o *Orchestrator) {
		o.queue = q
	}
}
