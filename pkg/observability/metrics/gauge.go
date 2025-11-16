package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// gauge wraps a Prometheus gauge metric.
type gauge struct {
	gauge prometheus.Gauge
}

// newGauge creates a new gauge metric from a Prometheus gauge.
func newGauge(g prometheus.Gauge) Gauge {
	return &gauge{gauge: g}
}

// Set sets the gauge to the given value.
func (g *gauge) Set(value float64) {
	g.gauge.Set(value)
}

// Inc increments the gauge by 1.
func (g *gauge) Inc() {
	g.gauge.Inc()
}

// Dec decrements the gauge by 1.
func (g *gauge) Dec() {
	g.gauge.Dec()
}

// Add adds the given value to the gauge (can be negative).
func (g *gauge) Add(value float64) {
	g.gauge.Add(value)
}
