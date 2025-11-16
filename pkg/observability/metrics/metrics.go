// Package metrics provides an abstraction over metrics collection systems.
package metrics

import "context"

// Labels represents metric labels as key-value pairs.
// Labels should be low-cardinality (limited distinct values).
type Labels map[string]string

// Counter represents a monotonically increasing counter metric.
// Counters can only increase (e.g., total requests, total errors).
type Counter interface {
	// Inc increments the counter by 1.
	Inc()

	// Add increments the counter by the given value.
	// The value must be non-negative.
	Add(value float64)
}

// Gauge represents a metric that can go up and down.
// Gauges are used for values that can increase or decrease (e.g., active connections, queue size).
type Gauge interface {
	// Set sets the gauge to the given value.
	Set(value float64)

	// Inc increments the gauge by 1.
	Inc()

	// Dec decrements the gauge by 1.
	Dec()

	// Add adds the given value to the gauge (can be negative).
	Add(value float64)
}

// Histogram represents a metric that samples observations and counts them in configurable buckets.
// Histograms are used for measuring distributions (e.g., request duration, response size).
type Histogram interface {
	// Observe records an observation (e.g., request duration in seconds).
	Observe(value float64)

	// ObserveWithContext records an observation with context for trace correlation.
	ObserveWithContext(ctx context.Context, value float64)
}

// Provider is a factory for creating metrics.
// Each implementation (Prometheus, Noop, etc.) provides its own Provider.
type Provider interface {
	// Counter creates or retrieves a counter metric.
	// Name should follow the convention: spotlight_<subsystem>_<metric>_total
	// Example: spotlight_http_requests_total
	Counter(name, help string, labels Labels) Counter

	// Gauge creates or retrieves a gauge metric.
	// Name should follow the convention: spotlight_<subsystem>_<metric>
	// Example: spotlight_active_connections
	Gauge(name, help string, labels Labels) Gauge

	// Histogram creates or retrieves a histogram metric.
	// Name should follow the convention: spotlight_<subsystem>_<metric>_<unit>
	// Example: spotlight_http_request_duration_seconds
	// Buckets define the histogram buckets. If nil, default buckets are used.
	Histogram(name, help string, labels Labels, buckets []float64) Histogram
}
