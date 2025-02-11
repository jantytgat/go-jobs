package task

import "github.com/prometheus/client_golang/prometheus"

type HandlerPoolOption func(*HandlerPool)

func WithHandlerPoolPrometheusRegister(r prometheus.Registerer) HandlerPoolOption {
	return func(p *HandlerPool) {
		_ = handlerPoolMetrics.Register(prometheus.WrapRegistererWith(map[string]string{"handler": p.Name()}, r))
	}
}
