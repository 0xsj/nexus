package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultBuckets are the default histogram buckets (in seconds).
// Covers from 1ms to 10s, suitable for most HTTP/RPC latencies.
var DefaultBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// prometheusProvider implements Provider using Prometheus.
type prometheusProvider struct {
	registry  *prometheus.Registry
	namespace string

	// Metric caches to avoid re-creating metrics
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	mu         sync.RWMutex
}

// PrometheusConfig configures the Prometheus provider.
type PrometheusConfig struct {
	// Namespace is the prefix for all metrics (e.g., "spotlight")
	Namespace string

	// Registry is the Prometheus registry to use.
	// If nil, a new registry is created.
	Registry *prometheus.Registry
}

// NewPrometheusProvider creates a new Prometheus metrics provider.
func NewPrometheusProvider(config PrometheusConfig) Provider {
	registry := config.Registry
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	return &prometheusProvider{
		registry:   registry,
		namespace:  config.Namespace,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}
}

// Counter creates or retrieves a counter metric.
func (p *prometheusProvider) Counter(name, help string, labels Labels) Counter {
	labelNames := extractLabelNames(labels)
	key := metricKey(name, labelNames)

	p.mu.RLock()
	vec, exists := p.counters[key]
	p.mu.RUnlock()

	if !exists {
		p.mu.Lock()
		// Double-check after acquiring write lock
		if vec, exists = p.counters[key]; !exists {
			vec = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: p.namespace,
					Name:      name,
					Help:      help,
				},
				labelNames,
			)

			if err := p.registry.Register(vec); err != nil {
				// If already registered, try to get existing
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					vec = are.ExistingCollector.(*prometheus.CounterVec)
				} else {
					// Return noop on registration error
					return NewNoopCounter()
				}
			}

			p.counters[key] = vec
		}
		p.mu.Unlock()
	}

	return newCounter(vec.With(toPrometheusLabels(labels)))
}

// Gauge creates or retrieves a gauge metric.
func (p *prometheusProvider) Gauge(name, help string, labels Labels) Gauge {
	labelNames := extractLabelNames(labels)
	key := metricKey(name, labelNames)

	p.mu.RLock()
	vec, exists := p.gauges[key]
	p.mu.RUnlock()

	if !exists {
		p.mu.Lock()
		// Double-check after acquiring write lock
		if vec, exists = p.gauges[key]; !exists {
			vec = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: p.namespace,
					Name:      name,
					Help:      help,
				},
				labelNames,
			)

			if err := p.registry.Register(vec); err != nil {
				// If already registered, try to get existing
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					vec = are.ExistingCollector.(*prometheus.GaugeVec)
				} else {
					// Return noop on registration error
					return NewNoopGauge()
				}
			}

			p.gauges[key] = vec
		}
		p.mu.Unlock()
	}

	return newGauge(vec.With(toPrometheusLabels(labels)))
}

// Histogram creates or retrieves a histogram metric.
func (p *prometheusProvider) Histogram(name, help string, labels Labels, buckets []float64) Histogram {
	if buckets == nil {
		buckets = DefaultBuckets
	}

	labelNames := extractLabelNames(labels)
	key := metricKey(name, labelNames)

	p.mu.RLock()
	vec, exists := p.histograms[key]
	p.mu.RUnlock()

	if !exists {
		p.mu.Lock()
		// Double-check after acquiring write lock
		if vec, exists = p.histograms[key]; !exists {
			vec = prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: p.namespace,
					Name:      name,
					Help:      help,
					Buckets:   buckets,
				},
				labelNames,
			)

			if err := p.registry.Register(vec); err != nil {
				// If already registered, try to get existing
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					vec = are.ExistingCollector.(*prometheus.HistogramVec)
				} else {
					// Return noop on registration error
					return NewNoopHistogram()
				}
			}

			p.histograms[key] = vec
		}
		p.mu.Unlock()
	}

	return newHistogram(vec.With(toPrometheusLabels(labels)))
}

// Registry returns the underlying Prometheus registry.
// This is useful for exposing metrics via HTTP.
func (p *prometheusProvider) Registry() *prometheus.Registry {
	return p.registry
}

// Helper functions

// extractLabelNames extracts label keys from Labels map.
func extractLabelNames(labels Labels) []string {
	if labels == nil {
		return []string{}
	}

	names := make([]string, 0, len(labels))
	for k := range labels {
		names = append(names, k)
	}
	return names
}

// toPrometheusLabels converts our Labels type to Prometheus labels.
func toPrometheusLabels(labels Labels) prometheus.Labels {
	if labels == nil {
		return prometheus.Labels{}
	}

	promLabels := make(prometheus.Labels, len(labels))
	for k, v := range labels {
		promLabels[k] = v
	}
	return promLabels
}

// metricKey creates a unique key for caching metric vectors.
func metricKey(name string, labelNames []string) string {
	return fmt.Sprintf("%s:%v", name, labelNames)
}
