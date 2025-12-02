// pkg/grpc/server.go

package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/0xsj/nexus-go/pkg/grpc/interceptors"
	"github.com/0xsj/nexus-go/pkg/observability/logger"
	"github.com/0xsj/nexus-go/pkg/observability/metrics"
	"github.com/0xsj/nexus-go/pkg/observability/tracing"
	"github.com/0xsj/nexus-go/pkg/security/jwt"
)

// ============================================================================
// Server Configuration
// ============================================================================

// Config holds gRPC server configuration.
type Config struct {
	// Host is the address to bind to.
	Host string `env:"HOST"`

	// Port is the port to listen on.
	Port int `env:"PORT"`

	// EnableReflection enables gRPC reflection for debugging.
	EnableReflection bool `env:"ENABLE_REFLECTION"`

	// EnableHealthCheck enables the gRPC health check service.
	EnableHealthCheck bool `env:"ENABLE_HEALTH_CHECK"`

	// MaxRecvMsgSize is the maximum message size in bytes the server can receive.
	MaxRecvMsgSize int `env:"MAX_RECV_MSG_SIZE"`

	// MaxSendMsgSize is the maximum message size in bytes the server can send.
	MaxSendMsgSize int `env:"MAX_SEND_MSG_SIZE"`

	// ConnectionTimeout is the timeout for connection establishment.
	ConnectionTimeout time.Duration `env:"CONNECTION_TIMEOUT"`

	// ShutdownTimeout is the timeout for graceful shutdown.
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT"`
}

// DefaultConfig returns default server configuration.
func DefaultConfig() Config {
	return Config{
		Host:              "0.0.0.0",
		Port:              50051,
		EnableReflection:  true,
		EnableHealthCheck: true,
		MaxRecvMsgSize:    4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:    4 * 1024 * 1024, // 4MB
		ConnectionTimeout: 10 * time.Second,
		ShutdownTimeout:   30 * time.Second,
	}
}

// Address returns the full address string.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ============================================================================
// Server
// ============================================================================

// Server wraps a gRPC server with lifecycle management.
type Server struct {
	config       Config
	server       *grpc.Server
	healthServer *health.Server
	logger       logger.Logger
	listener     net.Listener
}

// ServerOption configures the server.
type ServerOption func(*serverOptions)

type serverOptions struct {
	// Interceptors
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor

	// Dependencies for built-in interceptors
	logger          logger.Logger
	jwtService      jwt.Service
	tracer          tracing.Tracer
	metricsProvider metrics.Provider

	// Interceptor options
	skipAuthMethods    []string
	skipLoggingMethods []string
	skipTracingMethods []string
	skipMetricsMethods []string

	// Additional server options
	grpcOptions []grpc.ServerOption
}

// WithLogger sets the logger for interceptors.
func WithLogger(log logger.Logger) ServerOption {
	return func(o *serverOptions) {
		o.logger = log
	}
}

// WithJWTService sets the JWT service for auth interceptor.
func WithJWTService(jwtService jwt.Service) ServerOption {
	return func(o *serverOptions) {
		o.jwtService = jwtService
	}
}

// WithTracer sets the tracer for tracing interceptor.
func WithTracer(tracer tracing.Tracer) ServerOption {
	return func(o *serverOptions) {
		o.tracer = tracer
	}
}

// WithMetricsProvider sets the metrics provider for metrics interceptor.
func WithMetricsProvider(provider metrics.Provider) ServerOption {
	return func(o *serverOptions) {
		o.metricsProvider = provider
	}
}

// WithSkipAuthMethods sets methods to skip authentication.
func WithSkipAuthMethods(methods ...string) ServerOption {
	return func(o *serverOptions) {
		o.skipAuthMethods = append(o.skipAuthMethods, methods...)
	}
}

// WithSkipLoggingMethods sets methods to skip logging.
func WithSkipLoggingMethods(methods ...string) ServerOption {
	return func(o *serverOptions) {
		o.skipLoggingMethods = append(o.skipLoggingMethods, methods...)
	}
}

// WithSkipTracingMethods sets methods to skip tracing.
func WithSkipTracingMethods(methods ...string) ServerOption {
	return func(o *serverOptions) {
		o.skipTracingMethods = append(o.skipTracingMethods, methods...)
	}
}

// WithSkipMetricsMethods sets methods to skip metrics.
func WithSkipMetricsMethods(methods ...string) ServerOption {
	return func(o *serverOptions) {
		o.skipMetricsMethods = append(o.skipMetricsMethods, methods...)
	}
}

// WithUnaryInterceptor adds a custom unary interceptor.
func WithUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) ServerOption {
	return func(o *serverOptions) {
		o.unaryInterceptors = append(o.unaryInterceptors, interceptor)
	}
}

// WithStreamInterceptor adds a custom stream interceptor.
func WithStreamInterceptor(interceptor grpc.StreamServerInterceptor) ServerOption {
	return func(o *serverOptions) {
		o.streamInterceptors = append(o.streamInterceptors, interceptor)
	}
}

// WithGRPCOption adds a raw gRPC server option.
func WithGRPCOption(opt grpc.ServerOption) ServerOption {
	return func(o *serverOptions) {
		o.grpcOptions = append(o.grpcOptions, opt)
	}
}

// NewServer creates a new gRPC server.
func NewServer(config Config, opts ...ServerOption) *Server {
	options := &serverOptions{}

	for _, opt := range opts {
		opt(options)
	}

	// Build interceptor chains
	unaryInterceptors := buildUnaryInterceptors(options)
	streamInterceptors := buildStreamInterceptors(options)

	// Build gRPC server options
	grpcOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.MaxSendMsgSize),
		grpc.ConnectionTimeout(config.ConnectionTimeout),
	}

	// Add interceptors
	if len(unaryInterceptors) > 0 {
		grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		grpcOpts = append(grpcOpts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	// Add custom gRPC options
	grpcOpts = append(grpcOpts, options.grpcOptions...)

	// Create gRPC server
	grpcServer := grpc.NewServer(grpcOpts...)

	// Create health server
	var healthServer *health.Server
	if config.EnableHealthCheck {
		healthServer = health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	}

	// Enable reflection
	if config.EnableReflection {
		reflection.Register(grpcServer)
	}

	return &Server{
		config:       config,
		server:       grpcServer,
		healthServer: healthServer,
		logger:       options.logger,
	}
}

// buildUnaryInterceptors builds the unary interceptor chain.
func buildUnaryInterceptors(opts *serverOptions) []grpc.UnaryServerInterceptor {
	var chain []grpc.UnaryServerInterceptor

	// 1. Request ID (first - generates ID for all subsequent interceptors)
	chain = append(chain, interceptors.UnaryRequestIDInterceptor())

	// 2. Recovery (early - catches panics from all handlers)
	if opts.logger != nil {
		chain = append(chain, interceptors.UnaryRecoveryInterceptor(
			interceptors.WithRecoveryLogger(opts.logger),
		))
	}

	// 3. Metrics (before auth - counts all requests including auth failures)
	if opts.metricsProvider != nil {
		grpcMetrics := interceptors.NewGRPCMetrics(opts.metricsProvider)
		chain = append(chain, interceptors.UnaryMetricsInterceptor(
			grpcMetrics,
			interceptors.WithSkipMetrics(opts.skipMetricsMethods...),
		))
	}

	// 4. Tracing (before auth - traces full request lifecycle)
	if opts.tracer != nil {
		chain = append(chain, interceptors.UnaryTracingInterceptor(
			opts.tracer,
			interceptors.WithSkipTracing(opts.skipTracingMethods...),
		))
	}

	// 5. Logging
	if opts.logger != nil {
		chain = append(chain, interceptors.UnaryLoggingInterceptor(
			opts.logger,
			interceptors.WithSkipMethods(opts.skipLoggingMethods...),
		))
	}

	// 6. Auth (after observability - we want to log/trace auth failures)
	if opts.jwtService != nil {
		chain = append(chain, interceptors.UnaryAuthInterceptor(
			opts.jwtService,
			opts.logger,
			interceptors.WithSkipAuth(opts.skipAuthMethods...),
		))
	}

	// 7. Custom interceptors (last - after all built-in interceptors)
	chain = append(chain, opts.unaryInterceptors...)

	return chain
}

// buildStreamInterceptors builds the stream interceptor chain.
func buildStreamInterceptors(opts *serverOptions) []grpc.StreamServerInterceptor {
	var chain []grpc.StreamServerInterceptor

	// 1. Request ID
	chain = append(chain, interceptors.StreamRequestIDInterceptor())

	// 2. Recovery
	if opts.logger != nil {
		chain = append(chain, interceptors.StreamRecoveryInterceptor(
			interceptors.WithRecoveryLogger(opts.logger),
		))
	}

	// 3. Metrics
	if opts.metricsProvider != nil {
		grpcMetrics := interceptors.NewGRPCMetrics(opts.metricsProvider)
		chain = append(chain, interceptors.StreamMetricsInterceptor(
			grpcMetrics,
			interceptors.WithSkipMetrics(opts.skipMetricsMethods...),
		))
	}

	// 4. Tracing
	if opts.tracer != nil {
		chain = append(chain, interceptors.StreamTracingInterceptor(
			opts.tracer,
			interceptors.WithSkipTracing(opts.skipTracingMethods...),
		))
	}

	// 5. Logging
	if opts.logger != nil {
		chain = append(chain, interceptors.StreamLoggingInterceptor(
			opts.logger,
			interceptors.WithSkipMethods(opts.skipLoggingMethods...),
		))
	}

	// 6. Auth
	if opts.jwtService != nil {
		chain = append(chain, interceptors.StreamAuthInterceptor(
			opts.jwtService,
			opts.logger,
			interceptors.WithSkipAuth(opts.skipAuthMethods...),
		))
	}

	// 7. Custom interceptors
	chain = append(chain, opts.streamInterceptors...)

	return chain
}

// ============================================================================
// Server Lifecycle
// ============================================================================

// Server returns the underlying gRPC server for service registration.
func (s *Server) Server() *grpc.Server {
	return s.server
}

// SetServingStatus sets the health status for a service.
func (s *Server) SetServingStatus(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	if s.healthServer != nil {
		s.healthServer.SetServingStatus(service, status)
	}
}

// Start starts the gRPC server.
func (s *Server) Start(ctx context.Context) error {
	addr := s.config.Address()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener

	if s.logger != nil {
		s.logger.Info("grpc server starting",
			logger.String("address", addr),
			logger.Bool("reflection", s.config.EnableReflection),
			logger.Bool("health_check", s.config.EnableHealthCheck),
		)
	}

	// Set overall health status to serving
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	}

	// Start serving in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := s.server.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	// Check for immediate startup errors
	select {
	case err := <-errCh:
		return fmt.Errorf("grpc server failed to start: %w", err)
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop(ctx context.Context) error {
	if s.logger != nil {
		s.logger.Info("grpc server stopping")
	}

	// Set health status to not serving
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	// Graceful stop with timeout
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		if s.logger != nil {
			s.logger.Info("grpc server stopped gracefully")
		}
		return nil
	case <-shutdownCtx.Done():
		if s.logger != nil {
			s.logger.Warn("grpc server graceful stop timed out, forcing stop")
		}
		s.server.Stop()
		return shutdownCtx.Err()
	}
}

// ============================================================================
// Helper for common skip patterns
// ============================================================================

// HealthCheckMethods returns the health check method names for skipping.
func HealthCheckMethods() []string {
	return []string{
		"/grpc.health.v1.Health/Check",
		"/grpc.health.v1.Health/Watch",
	}
}
