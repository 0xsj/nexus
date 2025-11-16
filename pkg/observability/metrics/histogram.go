package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// histogram wraps a Prometheus histogram metric.
type histogram struct {
	histogram prometheus.Observer
}

// newHistogram creates a new histogram metric from a Prometheus observer.
func newHistogram(h prometheus.Observer) Histogram {
	return &histogram{histogram: h}
}

// Observe records an observation (e.g., request duration in seconds).
func (h *histogram) Observe(value float64) {
	h.histogram.Observe(value)
}

// ObserveWithContext records an observation with context for trace correlation.
// Currently ignores context, but allows future integration with tracing.
func (h *histogram) ObserveWithContext(ctx context.Context, value float64) {
	// TODO: Extract trace/span info from context for exemplars when we add tracing
	h.histogram.Observe(value)
}
