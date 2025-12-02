// pkg/grpc/interceptors/requestid.go

package interceptors

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ============================================================================
// Request ID Interceptor
// ============================================================================

const (
	// RequestIDHeader is the metadata key for request ID.
	RequestIDHeader = "x-request-id"
)

// RequestIDInterceptor provides request ID generation and propagation.
type RequestIDInterceptor struct {
	generator func() string
}

// RequestIDOption configures the request ID interceptor.
type RequestIDOption func(*RequestIDInterceptor)

// WithIDGenerator sets a custom ID generator function.
func WithIDGenerator(generator func() string) RequestIDOption {
	return func(r *RequestIDInterceptor) {
		r.generator = generator
	}
}

// NewRequestIDInterceptor creates a new request ID interceptor.
func NewRequestIDInterceptor(opts ...RequestIDOption) *RequestIDInterceptor {
	r := &RequestIDInterceptor{
		generator: func() string {
			return uuid.New().String()
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Unary returns a unary server interceptor that handles request IDs.
//
// Usage:
//
//	requestID := interceptors.NewRequestIDInterceptor()
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(requestID.Unary()),
//	)
func (r *RequestIDInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx = r.extractOrGenerateRequestID(ctx)
		return handler(ctx, req)
	}
}

// Stream returns a stream server interceptor that handles request IDs.
//
// Usage:
//
//	requestID := interceptors.NewRequestIDInterceptor()
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(requestID.Stream()),
//	)
func (r *RequestIDInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := r.extractOrGenerateRequestID(ss.Context())

		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}

// extractOrGenerateRequestID extracts request ID from metadata or generates a new one.
func (r *RequestIDInterceptor) extractOrGenerateRequestID(ctx context.Context) context.Context {
	var requestID string

	// Try to extract from incoming metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get(RequestIDHeader); len(values) > 0 && values[0] != "" {
			requestID = values[0]
		}
	}

	// Generate if not present
	if requestID == "" {
		requestID = r.generator()
	}

	// Add to context
	ctx = WithRequestID(ctx, requestID)

	// Add to outgoing metadata for downstream propagation
	ctx = metadata.AppendToOutgoingContext(ctx, RequestIDHeader, requestID)

	return ctx
}

// ============================================================================
// Functional API (Alternative)
// ============================================================================

// UnaryRequestIDInterceptor returns a unary interceptor with functional options.
//
// Usage:
//
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(interceptors.UnaryRequestIDInterceptor()),
//	)
func UnaryRequestIDInterceptor(opts ...RequestIDOption) grpc.UnaryServerInterceptor {
	interceptor := NewRequestIDInterceptor(opts...)
	return interceptor.Unary()
}

// StreamRequestIDInterceptor returns a stream interceptor with functional options.
//
// Usage:
//
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(interceptors.StreamRequestIDInterceptor()),
//	)
func StreamRequestIDInterceptor(opts ...RequestIDOption) grpc.StreamServerInterceptor {
	interceptor := NewRequestIDInterceptor(opts...)
	return interceptor.Stream()
}