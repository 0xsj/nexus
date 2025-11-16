package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// counter wraps a Prometheus counter metric.
type counter struct {
	counter prometheus.Counter
}

// newCounter creates a new counter metric from a Prometheus counter.
func newCounter(c prometheus.Counter) Counter {
	return &counter{counter: c}
}

// Inc increments the counter by 1.
func (c *counter) Inc() {
	c.counter.Inc()
}

// Add increments the counter by the given value.
// Negative values are silently ignored as counters can only increase.
func (c *counter) Add(value float64) {
	if value < 0 {
		return
	}
	c.counter.Add(value)
}
