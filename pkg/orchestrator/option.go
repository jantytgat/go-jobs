package orchestrator

import (
	"github.com/prometheus/client_golang/prometheus"

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

func WithPrometheusRegistry(reg prometheus.Registerer) Option {
	return func(o *Orchestrator) {
		o.reg = prometheus.WrapRegistererWith(map[string]string{"orchestrator": o.name}, reg)
		_ = metrics.Register(o.reg)

	}
}

func WithQueue(q Queue) Option {
	return func(o *Orchestrator) {
		o.queue = q
	}
}
