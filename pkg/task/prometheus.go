package task

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// https://stackoverflow.com/a/58875389/17906878
// GetMetricValue returns the sum of the Counter metrics associated with the Collector
// e.g. the metric for a non-vector, or the sum of the metrics for vector labels.
// If the metric is a Histogram then number of samples is used.
func GetMetricValue(col prometheus.Collector) float64 {
	var total float64
	collect(col, func(m dto.Metric) {
		if h := m.GetHistogram(); h != nil {
			total += float64(h.GetSampleCount())
		} else if g := m.GetGauge(); g != nil {
			total += g.GetValue()
		} else {
			total += m.GetCounter().GetValue()
		}
	})
	return total
}

// collect calls the function for each metric associated with the Collector
func collect(col prometheus.Collector, do func(dto.Metric)) {
	c := make(chan prometheus.Metric)
	go func(c chan prometheus.Metric) {
		col.Collect(c)
		close(c)
	}(c)
	for x := range c { // eg range across distinct label vector values
		m := dto.Metric{}
		_ = x.Write(&m)
		do(m)
	}
}
