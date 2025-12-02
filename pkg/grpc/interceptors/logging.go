// pkg/grpc/interceptors/logging.go

package interceptors

import (
	"context"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xsj/nexus-go/pkg/observability/logger"
)

// ============================================================================
// Logging Interceptor
// ============================================================================

// LoggingInterceptor provides request/response logging for gRPC handlers.
type LoggingInterceptor struct {
	logger logger.Logger
	opts   loggingOptions
}

// loggingOptions holds configuration for the logging interceptor.
type loggingOptions struct {
	// logPayloads enables logging of request/response payloads.
	// Use with caution in production (may log sensitive data).
	logPayloads bool

	// skipMethods is a list of methods to skip logging (e.g., health checks).
	skipMethods map[string]bool

	// slowThreshold is the duration above which requests are logged as slow.
	slowThreshold time.Duration
}

// LoggingOption configures the logging interceptor.
type LoggingOption func(*loggingOptions)

// WithPayloadLogging enables logging of request/response payloads.
// Use with caution — may log sensitive data.
func WithPayloadLogging(enabled bool) LoggingOption {
	return func(o *loggingOptions) {
		o.logPayloads = enabled
	}
}

// WithSkipMethods sets methods to skip logging (e.g., health checks).
func WithSkipMethods(methods ...string) LoggingOption {
	return func(o *loggingOptions) {
		if o.skipMethods == nil {
			o.skipMethods = make(map[string]bool)
		}
		for _, m := range methods {
			o.skipMethods[m] = true
		}
	}
}

// WithSlowThreshold sets the duration threshold for slow request logging.
// Requests exceeding this duration are logged with a warning.
func WithSlowThreshold(d time.Duration) LoggingOption {
	return func(o *loggingOptions) {
		o.slowThreshold = d
	}
}

// NewLoggingInterceptor creates a new logging interceptor.
func NewLoggingInterceptor(log logger.Logger, opts ...LoggingOption) *LoggingInterceptor {
	options := loggingOptions{
		slowThreshold: 3 * time.Second, // Default slow threshold
		skipMethods:   make(map[string]bool),
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &LoggingInterceptor{
		logger: log,
		opts:   options,
	}
}

// Unary returns a unary server interceptor that logs requests.
//
// Usage:
//
//	logging := interceptors.NewLoggingInterceptor(log)
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(logging.Unary()),
//	)
func (l *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip logging for excluded methods
		if l.opts.skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract method info
		service := path.Dir(info.FullMethod)[1:] // Remove leading "/"
		method := path.Base(info.FullMethod)

		// Build common fields
		fields := l.buildFields(ctx, service, method)

		// Log request start
		l.logger.Info("grpc request started", fields...)

		// Log payload if enabled
		if l.opts.logPayloads && req != nil {
			l.logger.Debug("grpc request payload",
				logger.String("method", info.FullMethod),
				logger.Any("payload", req),
			)
		}

		// Execute handler
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Log response
		l.logResponse(info.FullMethod, service, method, duration, err, fields)

		// Log response payload if enabled
		if l.opts.logPayloads && resp != nil && err == nil {
			l.logger.Debug("grpc response payload",
				logger.String("method", info.FullMethod),
				logger.Any("payload", resp),
			)
		}

		return resp, err
	}
}

// Stream returns a stream server interceptor that logs requests.
//
// Usage:
//
//	logging := interceptors.NewLoggingInterceptor(log)
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(logging.Stream()),
//	)
func (l *LoggingInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip logging for excluded methods
		if l.opts.skipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// Extract method info
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		// Build common fields
		ctx := ss.Context()
		fields := l.buildFields(ctx, service, method)

		// Add stream-specific fields
		fields = append(fields,
			logger.Bool("client_stream", info.IsClientStream),
			logger.Bool("server_stream", info.IsServerStream),
		)

		// Log stream start
		l.logger.Info("grpc stream started", fields...)

		// Execute handler
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		// Log stream completion
		l.logResponse(info.FullMethod, service, method, duration, err, fields)

		return err
	}
}

// buildFields creates common log fields from context.
func (l *LoggingInterceptor) buildFields(ctx context.Context, service, method string) []logger.Field {
	fields := []logger.Field{
		logger.String("service", service),
		logger.String("method", method),
	}

	// Extract request ID
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, logger.String("request_id", requestID))
	}

	// Extract tenant ID
	if tenantID := getTenantIDFromContext(ctx); tenantID != "" {
		fields = append(fields, logger.String("tenant_id", tenantID))
	}

	// Extract user ID
	if userID := getUserIDFromContext(ctx); userID != "" {
		fields = append(fields, logger.String("user_id", userID))
	}

	return fields
}

// logResponse logs the response with appropriate level based on status.
func (l *LoggingInterceptor) logResponse(
	fullMethod string,
	service string,
	method string,
	duration time.Duration,
	err error,
	baseFields []logger.Field,
) {
	// Get gRPC status code
	code := codes.OK
	if err != nil {
		code = status.Code(err)
	}

	// Build response fields
	fields := append(baseFields,
		logger.String("grpc_code", code.String()),
		logger.Duration("duration", duration),
	)

	// Add error message if present
	if err != nil {
		fields = append(fields, logger.String("error", err.Error()))
	}

	// Determine log level based on code and duration
	switch {
	case code == codes.OK && duration > l.opts.slowThreshold:
		l.logger.Warn("grpc request completed (slow)", fields...)

	case code == codes.OK:
		l.logger.Info("grpc request completed", fields...)

	case isClientError(code):
		l.logger.Warn("grpc request failed (client error)", fields...)

	default:
		l.logger.Error("grpc request failed (server error)", fields...)
	}
}

// isClientError returns true if the code represents a client-side error.
func isClientError(code codes.Code) bool {
	switch code {
	case codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated,
		codes.FailedPrecondition,
		codes.OutOfRange:
		return true
	default:
		return false
	}
}

// ============================================================================
// Functional API (Alternative)
// ============================================================================

// UnaryLoggingInterceptor returns a unary interceptor with functional options.
//
// Usage:
//
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(interceptors.UnaryLoggingInterceptor(
//	        log,
//	        interceptors.WithSlowThreshold(5*time.Second),
//	    )),
//	)
func UnaryLoggingInterceptor(log logger.Logger, opts ...LoggingOption) grpc.UnaryServerInterceptor {
	interceptor := NewLoggingInterceptor(log, opts...)
	return interceptor.Unary()
}

// StreamLoggingInterceptor returns a stream interceptor with functional options.
//
// Usage:
//
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(interceptors.StreamLoggingInterceptor(
//	        log,
//	        interceptors.WithSlowThreshold(5*time.Second),
//	    )),
//	)
func StreamLoggingInterceptor(log logger.Logger, opts ...LoggingOption) grpc.StreamServerInterceptor {
	interceptor := NewLoggingInterceptor(log, opts...)
	return interceptor.Stream()
}

// ============================================================================
// Context Helpers
// ============================================================================

// getTenantIDFromContext extracts the tenant ID from context.
func getTenantIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(tenantIDKey{}).(string); ok {
		return id
	}
	return ""
}

// getUserIDFromContext extracts the user ID from context.
func getUserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey{}).(string); ok {
		return id
	}
	return ""
}
