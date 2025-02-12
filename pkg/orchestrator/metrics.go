package orchestrator

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

var metrics = NewMetrics()

func NewMetrics() *Metrics {
	m := &Metrics{
		queueLength: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "",
			Subsystem: "",
			Name:      "queue_length",
			Help:      "",
		},
			[]string{"name"}),

		jobsProcessed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "",
			Subsystem: "",
			Name:      "jobs_processed_total",
			Help:      "",
		},
			[]string{"name", "status"}),
	}

	return m
}

type Metrics struct {
	queueLength   *prometheus.GaugeVec
	jobsProcessed *prometheus.GaugeVec
}

func (m *Metrics) Register(reg prometheus.Registerer) error {
	var err error
	if err = reg.Register(metrics.queueLength); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	if err = reg.Register(metrics.jobsProcessed); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return err
	}
	return nil
}
