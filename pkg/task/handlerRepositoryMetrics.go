package task

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

var handlerRepositoryMetrics = NewHandlerRepositoryMetrics()

func NewHandlerRepositoryMetrics() *HandlerRepositoryMetrics {
	m := &HandlerRepositoryMetrics{
		handlerPools: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "handlerpools_total",
			Help:        "",
			ConstLabels: nil,
		},
			[]string{"handler_repository"}),
	}
	return m
}

type HandlerRepositoryMetrics struct {
	handlerPools *prometheus.GaugeVec
}

func (m *HandlerRepositoryMetrics) Register(reg prometheus.Registerer) error {
	var err error
	if err = reg.Register(handlerRepositoryMetrics.handlerPools); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	return nil
}
