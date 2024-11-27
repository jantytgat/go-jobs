package task

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

var handlerPoolMetrics = NewHandlerPoolMetrics()

func NewHandlerPoolMetrics() *HandlerPoolMetrics {
	m := &HandlerPoolMetrics{
		maxWorkers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "workers_max",
			Help:        "",
			ConstLabels: nil,
		},
			[]string{"handler"}),

		workers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "workers_total",
			Help:        "",
			ConstLabels: nil,
		},
			[]string{"handler"}),

		activeWorkers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "workers_active",
			Help:        "",
			ConstLabels: nil,
		},
			[]string{"handler"}),

		idleWorkers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "workers_idle",
			Help:        "",
			ConstLabels: nil,
		},
			[]string{"handler"}),

		tasksIngested: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "tasks_ingested",
			Help:        "",
			ConstLabels: nil,
		},
			[]string{"handler"}),

		tasksProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "tasks_processed",
			Help:        "",
			ConstLabels: nil,
		},
			[]string{"handler", "status"}),

		tasksWaiting: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "tasks_waiting",
			Help:        "",
			ConstLabels: nil,
		},
			[]string{"handler"}),
	}

	return m
}

type HandlerPoolMetrics struct {
	maxWorkers     *prometheus.GaugeVec
	workers        *prometheus.GaugeVec
	activeWorkers  *prometheus.GaugeVec
	idleWorkers    *prometheus.GaugeVec
	tasksIngested  *prometheus.CounterVec
	tasksProcessed *prometheus.CounterVec
	tasksWaiting   *prometheus.GaugeVec
}

func (m *HandlerPoolMetrics) Register(reg prometheus.Registerer) error {
	var err error
	if err = reg.Register(handlerPoolMetrics.workers); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	if err = reg.Register(handlerPoolMetrics.maxWorkers); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	if err = reg.Register(handlerPoolMetrics.activeWorkers); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	if err = reg.Register(handlerPoolMetrics.idleWorkers); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	if err = reg.Register(handlerPoolMetrics.tasksIngested); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	if err = reg.Register(handlerPoolMetrics.tasksProcessed); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	if err = reg.Register(handlerPoolMetrics.tasksWaiting); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	return nil
}
