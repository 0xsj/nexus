package metrics

import "context"

// noopCounter is a counter that does nothing.
type noopCounter struct{}

// NewNoopCounter creates a new no-op counter.
func NewNoopCounter() Counter {
	return &noopCounter{}
}

func (n *noopCounter) Inc()              {}
func (n *noopCounter) Add(value float64) {}

// noopGauge is a gauge that does nothing.
type noopGauge struct{}

// NewNoopGauge creates a new no-op gauge.
func NewNoopGauge() Gauge {
	return &noopGauge{}
}

func (n *noopGauge) Set(value float64) {}
func (n *noopGauge) Inc()              {}
func (n *noopGauge) Dec()              {}
func (n *noopGauge) Add(value float64) {}

// noopHistogram is a histogram that does nothing.
type noopHistogram struct{}

// NewNoopHistogram creates a new no-op histogram.
func NewNoopHistogram() Histogram {
	return &noopHistogram{}
}

func (n *noopHistogram) Observe(value float64)                                 {}
func (n *noopHistogram) ObserveWithContext(ctx context.Context, value float64) {}

// noopProvider is a provider that returns no-op metrics.
type noopProvider struct{}

// NewNoopProvider creates a new no-op metrics provider.
func NewNoopProvider() Provider {
	return &noopProvider{}
}

func (n *noopProvider) Counter(name, help string, labels Labels) Counter {
	return NewNoopCounter()
}

func (n *noopProvider) Gauge(name, help string, labels Labels) Gauge {
	return NewNoopGauge()
}

func (n *noopProvider) Histogram(name, help string, labels Labels, buckets []float64) Histogram {
	return NewNoopHistogram()
}
