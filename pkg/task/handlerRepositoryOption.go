package task

import "github.com/prometheus/client_golang/prometheus"

type HandlerRepositoryOption func(r *HandlerRepository)

func WithHandlerRepositoryPrometheusRegister(reg prometheus.Registerer) HandlerRepositoryOption {
	return func(r *HandlerRepository) {
		r.reg = prometheus.WrapRegistererWith(map[string]string{"handler_repository": r.name}, reg)
		_ = handlerRepositoryMetrics.Register(r.reg)
	}
}
