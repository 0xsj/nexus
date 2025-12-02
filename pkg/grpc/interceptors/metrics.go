// pkg/grpc/interceptors/metrics.go

package interceptors

import (
	"context"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xsj/nexus-go/pkg/observability/metrics"
)

// ============================================================================
// gRPC Metrics
// ============================================================================

// GRPCMetrics holds the metrics for gRPC requests.
type GRPCMetrics struct {
	requestsTotal      metrics.Counter
	requestDuration    metrics.Histogram
	requestsInFlight   metrics.Gauge
	streamMsgSent      metrics.Counter
	streamMsgReceived  metrics.Counter
}

// NewGRPCMetrics creates gRPC metrics using the provided metrics provider.
func NewGRPCMetrics(provider metrics.Provider) *GRPCMetrics {
	return &GRPCMetrics{
		requestsTotal: provider.Counter(
			"grpc_requests_total",
			"Total number of gRPC requests",
			"service", "method", "grpc_code",
		),
		requestDuration: provider.Histogram(
			"grpc_request_duration_seconds",
			"gRPC request duration in seconds",
			metrics.HTTPLatencyBuckets(), // Same buckets work for gRPC
			"service", "method", "grpc_code",
		),
		requestsInFlight: provider.Gauge(
			"grpc_requests_in_flight",
			"Number of gRPC requests currently being processed",
			"service", "method",
		),
		streamMsgSent: provider.Counter(
			"grpc_stream_messages_sent_total",
			"Total number of gRPC stream messages sent",
			"service", "method",
		),
		streamMsgReceived: provider.Counter(
			"grpc_stream_messages_received_total",
			"Total number of gRPC stream messages received",
			"service", "method",
		),
	}
}

// ============================================================================
// Metrics Interceptor
// ============================================================================

// MetricsInterceptor provides metrics collection for gRPC handlers.
type MetricsInterceptor struct {
	metrics *GRPCMetrics
	opts    metricsOptions
}

// metricsOptions holds configuration for the metrics interceptor.
type metricsOptions struct {
	// skipMethods is a list of methods to skip metrics (e.g., health checks).
	skipMethods map[string]bool
}

// MetricsOption configures the metrics interceptor.
type MetricsOption func(*metricsOptions)

// WithSkipMetrics sets methods to skip metrics collection.
func WithSkipMetrics(methods ...string) MetricsOption {
	return func(o *metricsOptions) {
		if o.skipMethods == nil {
			o.skipMethods = make(map[string]bool)
		}
		for _, m := range methods {
			o.skipMethods[m] = true
		}
	}
}

// NewMetricsInterceptor creates a new metrics interceptor.
func NewMetricsInterceptor(m *GRPCMetrics, opts ...MetricsOption) *MetricsInterceptor {
	options := metricsOptions{
		skipMethods: make(map[string]bool),
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &MetricsInterceptor{
		metrics: m,
		opts:    options,
	}
}

// Unary returns a unary server interceptor that collects metrics.
//
// Usage:
//
//	grpcMetrics := interceptors.NewGRPCMetrics(metricsProvider)
//	metricsInterceptor := interceptors.NewMetricsInterceptor(grpcMetrics)
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(metricsInterceptor.Unary()),
//	)
func (m *MetricsInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip metrics for excluded methods
		if m.opts.skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract service and method
		service := path.Dir(info.FullMethod)[1:] // Remove leading "/"
		method := path.Base(info.FullMethod)

		// Track in-flight requests
		m.metrics.requestsInFlight.Inc(service, method)
		defer m.metrics.requestsInFlight.Dec(service, method)

		// Record start time
		start := time.Now()

		// Execute handler
		resp, err := handler(ctx, req)

		// Record metrics
		duration := time.Since(start).Seconds()
		code := statusCode(err)

		m.metrics.requestsTotal.Inc(service, method, code)
		m.metrics.requestDuration.Observe(duration, service, method, code)

		return resp, err
	}
}

// Stream returns a stream server interceptor that collects metrics.
//
// Usage:
//
//	grpcMetrics := interceptors.NewGRPCMetrics(metricsProvider)
//	metricsInterceptor := interceptors.NewMetricsInterceptor(grpcMetrics)
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(metricsInterceptor.Stream()),
//	)
func (m *MetricsInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip metrics for excluded methods
		if m.opts.skipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// Extract service and method
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		// Track in-flight requests
		m.metrics.requestsInFlight.Inc(service, method)
		defer m.metrics.requestsInFlight.Dec(service, method)

		// Record start time
		start := time.Now()

		// Wrap stream to count messages
		wrapped := &metricsServerStream{
			ServerStream: ss,
			metrics:      m.metrics,
			service:      service,
			method:       method,
		}

		// Execute handler
		err := handler(srv, wrapped)

		// Record metrics
		duration := time.Since(start).Seconds()
		code := statusCode(err)

		m.metrics.requestsTotal.Inc(service, method, code)
		m.metrics.requestDuration.Observe(duration, service, method, code)

		return err
	}
}

// statusCode extracts the gRPC status code from an error.
func statusCode(err error) string {
	if err == nil {
		return codes.OK.String()
	}
	st, _ := status.FromError(err)
	return st.Code().String()
}

// ============================================================================
// Metrics Server Stream
// ============================================================================

// metricsServerStream wraps a grpc.ServerStream to count messages.
type metricsServerStream struct {
	grpc.ServerStream
	metrics *GRPCMetrics
	service string
	method  string
}

// SendMsg wraps SendMsg to count sent messages.
func (s *metricsServerStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err == nil {
		s.metrics.streamMsgSent.Inc(s.service, s.method)
	}
	return err
}

// RecvMsg wraps RecvMsg to count received messages.
func (s *metricsServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.metrics.streamMsgReceived.Inc(s.service, s.method)
	}
	return err
}

// ============================================================================
// Functional API (Alternative)
// ============================================================================

// UnaryMetricsInterceptor returns a unary interceptor with functional options.
//
// Usage:
//
//	grpcMetrics := interceptors.NewGRPCMetrics(metricsProvider)
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(interceptors.UnaryMetricsInterceptor(
//	        grpcMetrics,
//	        interceptors.WithSkipMetrics("/grpc.health.v1.Health/Check"),
//	    )),
//	)
func UnaryMetricsInterceptor(m *GRPCMetrics, opts ...MetricsOption) grpc.UnaryServerInterceptor {
	interceptor := NewMetricsInterceptor(m, opts...)
	return interceptor.Unary()
}

// StreamMetricsInterceptor returns a stream interceptor with functional options.
//
// Usage:
//
//	grpcMetrics := interceptors.NewGRPCMetrics(metricsProvider)
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(interceptors.StreamMetricsInterceptor(
//	        grpcMetrics,
//	        interceptors.WithSkipMetrics("/grpc.health.v1.Health/Watch"),
//	    )),
//	)
func StreamMetricsInterceptor(m *GRPCMetrics, opts ...MetricsOption) grpc.StreamServerInterceptor {
	interceptor := NewMetricsInterceptor(m, opts...)
	return interceptor.Stream()
}