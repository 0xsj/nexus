// pkg/grpc/interceptors/recovery.go

package interceptors

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xsj/nexus-go/pkg/observability/logger"
)

// ============================================================================
// Recovery Interceptor
// ============================================================================

// RecoveryInterceptor provides panic recovery for gRPC handlers.
// It catches panics and converts them to gRPC Internal errors.
type RecoveryInterceptor struct {
	logger logger.Logger
}

// NewRecoveryInterceptor creates a new recovery interceptor.
func NewRecoveryInterceptor(log logger.Logger) *RecoveryInterceptor {
	return &RecoveryInterceptor{
		logger: log,
	}
}

// Unary returns a unary server interceptor that recovers from panics.
//
// Usage:
//
//	recovery := interceptors.NewRecoveryInterceptor(log)
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(recovery.Unary()),
//	)
func (r *RecoveryInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				err = r.handlePanic(ctx, info.FullMethod, rec)
			}
		}()

		return handler(ctx, req)
	}
}

// Stream returns a stream server interceptor that recovers from panics.
//
// Usage:
//
//	recovery := interceptors.NewRecoveryInterceptor(log)
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(recovery.Stream()),
//	)
func (r *RecoveryInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				err = r.handlePanic(ss.Context(), info.FullMethod, rec)
			}
		}()

		return handler(srv, ss)
	}
}

// handlePanic logs the panic and returns a gRPC Internal error.
func (r *RecoveryInterceptor) handlePanic(ctx context.Context, method string, recovered interface{}) error {
	// Capture stack trace
	stack := debug.Stack()

	// Build log fields
	fields := []logger.Field{
		logger.String("error", fmt.Sprintf("%v", recovered)),
		logger.String("method", method),
		logger.String("stack_trace", string(stack)),
	}

	// Extract request ID if available
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, logger.String("request_id", requestID))
	}

	// Log the panic
	r.logger.Error("grpc panic recovered", fields...)

	// Return safe error to client (don't expose panic details)
	return status.Error(codes.Internal, "an internal error occurred")
}

// ============================================================================
// Functional Options (Alternative API)
// ============================================================================

// RecoveryOption configures the recovery interceptor.
type RecoveryOption func(*recoveryOptions)

type recoveryOptions struct {
	logger         logger.Logger
	recoveryHandler RecoveryHandlerFunc
}

// RecoveryHandlerFunc is a function that handles panics.
// It receives the panic value and returns an error to send to the client.
type RecoveryHandlerFunc func(ctx context.Context, method string, recovered interface{}) error

// WithRecoveryLogger sets the logger for the recovery interceptor.
func WithRecoveryLogger(log logger.Logger) RecoveryOption {
	return func(o *recoveryOptions) {
		o.logger = log
	}
}

// WithRecoveryHandler sets a custom recovery handler.
// Use this to customize error responses or add additional logging.
func WithRecoveryHandler(handler RecoveryHandlerFunc) RecoveryOption {
	return func(o *recoveryOptions) {
		o.recoveryHandler = handler
	}
}

// UnaryRecoveryInterceptor returns a unary interceptor with functional options.
//
// Usage:
//
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(interceptors.UnaryRecoveryInterceptor(
//	        interceptors.WithRecoveryLogger(log),
//	    )),
//	)
func UnaryRecoveryInterceptor(opts ...RecoveryOption) grpc.UnaryServerInterceptor {
	options := applyRecoveryOptions(opts)

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				err = handleRecovery(ctx, info.FullMethod, rec, options)
			}
		}()

		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor returns a stream interceptor with functional options.
//
// Usage:
//
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(interceptors.StreamRecoveryInterceptor(
//	        interceptors.WithRecoveryLogger(log),
//	    )),
//	)
func StreamRecoveryInterceptor(opts ...RecoveryOption) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		options := applyRecoveryOptions(opts)

		defer func() {
			if rec := recover(); rec != nil {
				err = handleRecovery(ss.Context(), info.FullMethod, rec, options)
			}
		}()

		return handler(srv, ss)
	}
}

// applyRecoveryOptions applies functional options and returns the config.
func applyRecoveryOptions(opts []RecoveryOption) *recoveryOptions {
	options := &recoveryOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// handleRecovery processes a panic with the configured options.
func handleRecovery(ctx context.Context, method string, recovered interface{}, options *recoveryOptions) error {
	// Use custom handler if provided
	if options.recoveryHandler != nil {
		return options.recoveryHandler(ctx, method, recovered)
	}

	// Capture stack trace
	stack := debug.Stack()

	// Log if logger is available
	if options.logger != nil {
		fields := []logger.Field{
			logger.String("error", fmt.Sprintf("%v", recovered)),
			logger.String("method", method),
			logger.String("stack_trace", string(stack)),
		}

		if requestID := getRequestIDFromContext(ctx); requestID != "" {
			fields = append(fields, logger.String("request_id", requestID))
		}

		options.logger.Error("grpc panic recovered", fields...)
	}

	// Return safe error
	return status.Error(codes.Internal, "an internal error occurred")
}

// getRequestIDFromContext extracts the request ID from context.
func getRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}