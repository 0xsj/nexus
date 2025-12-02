// pkg/grpc/interceptors/tracing.go

package interceptors

import (
	"context"
	"path"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/0xsj/nexus-go/pkg/observability/tracing"
)

// ============================================================================
// Tracing Interceptor
// ============================================================================

// TracingInterceptor provides distributed tracing for gRPC handlers.
type TracingInterceptor struct {
	tracer tracing.Tracer
	opts   tracingOptions
}

// tracingOptions holds configuration for the tracing interceptor.
type tracingOptions struct {
	// skipMethods is a list of methods to skip tracing (e.g., health checks).
	skipMethods map[string]bool
}

// TracingOption configures the tracing interceptor.
type TracingOption func(*tracingOptions)

// WithSkipTracing sets methods to skip tracing.
func WithSkipTracing(methods ...string) TracingOption {
	return func(o *tracingOptions) {
		if o.skipMethods == nil {
			o.skipMethods = make(map[string]bool)
		}
		for _, m := range methods {
			o.skipMethods[m] = true
		}
	}
}

// NewTracingInterceptor creates a new tracing interceptor.
func NewTracingInterceptor(tracer tracing.Tracer, opts ...TracingOption) *TracingInterceptor {
	options := tracingOptions{
		skipMethods: make(map[string]bool),
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &TracingInterceptor{
		tracer: tracer,
		opts:   options,
	}
}

// Unary returns a unary server interceptor that traces requests.
//
// Usage:
//
//	tracing := interceptors.NewTracingInterceptor(tracer)
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(tracing.Unary()),
//	)
func (t *TracingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip tracing for excluded methods
		if t.opts.skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract method info for span name
		service := path.Dir(info.FullMethod)[1:] // Remove leading "/"
		method := path.Base(info.FullMethod)
		spanName := info.FullMethod

		// Start span
		ctx, span := t.tracer.Start(ctx, spanName,
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(map[string]any{
				"rpc.system":         "grpc",
				"rpc.service":        service,
				"rpc.method":         method,
				"rpc.grpc.full_method": info.FullMethod,
			}),
		)
		defer span.End()

		// Add context values as attributes
		t.addContextAttributes(ctx, span)

		// Add trace ID to outgoing metadata for debugging
		ctx = t.injectTraceMetadata(ctx, span)

		// Execute handler
		resp, err := handler(ctx, req)

		// Record result
		t.recordResult(span, err)

		return resp, err
	}
}

// Stream returns a stream server interceptor that traces requests.
//
// Usage:
//
//	tracing := interceptors.NewTracingInterceptor(tracer)
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(tracing.Stream()),
//	)
func (t *TracingInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip tracing for excluded methods
		if t.opts.skipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// Extract method info
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)
		spanName := info.FullMethod

		// Start span
		ctx, span := t.tracer.Start(ss.Context(), spanName,
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(map[string]any{
				"rpc.system":           "grpc",
				"rpc.service":          service,
				"rpc.method":           method,
				"rpc.grpc.full_method": info.FullMethod,
				"rpc.grpc.client_stream": info.IsClientStream,
				"rpc.grpc.server_stream": info.IsServerStream,
			}),
		)
		defer span.End()

		// Add context values as attributes
		t.addContextAttributes(ctx, span)

		// Wrap stream with traced context
		wrapped := &tracedServerStream{
			ServerStream: ss,
			ctx:          ctx,
			span:         span,
		}

		// Execute handler
		err := handler(srv, wrapped)

		// Record result
		t.recordResult(span, err)

		return err
	}
}

// addContextAttributes adds context values as span attributes.
func (t *TracingInterceptor) addContextAttributes(ctx context.Context, span tracing.Span) {
	if requestID := GetRequestID(ctx); requestID != "" {
		span.SetAttribute("request.id", requestID)
	}

	if tenantID := GetTenantID(ctx); tenantID != "" {
		span.SetAttribute("tenant.id", tenantID)
	}

	if userID := GetUserID(ctx); userID != "" {
		span.SetAttribute("user.id", userID)
	}

	// Extract peer info from metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get("user-agent"); len(userAgents) > 0 {
			span.SetAttribute("rpc.user_agent", userAgents[0])
		}
	}
}

// injectTraceMetadata injects trace ID into outgoing metadata.
func (t *TracingInterceptor) injectTraceMetadata(ctx context.Context, span tracing.Span) context.Context {
	sc := span.SpanContext()
	if sc.IsValid() {
		// Create or append to outgoing metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		md = md.Copy()
		md.Set("x-trace-id", sc.TraceID)
		md.Set("x-span-id", sc.SpanID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// recordResult records the result of the RPC call on the span.
func (t *TracingInterceptor) recordResult(span tracing.Span, err error) {
	if err != nil {
		// Get gRPC status code
		st, _ := status.FromError(err)
		code := st.Code()

		span.SetAttribute("rpc.grpc.status_code", int(code))
		span.SetAttribute("rpc.grpc.status", code.String())

		// Record error
		span.RecordError(err)

		// Set span status based on error type
		if isServerError(code) {
			span.SetStatus(tracing.StatusError, st.Message())
		} else {
			// Client errors are not span errors
			span.SetStatus(tracing.StatusOK, "")
		}
	} else {
		span.SetAttribute("rpc.grpc.status_code", int(codes.OK))
		span.SetAttribute("rpc.grpc.status", codes.OK.String())
		span.SetStatus(tracing.StatusOK, "")
	}
}

// isServerError returns true if the code represents a server-side error.
func isServerError(code codes.Code) bool {
	switch code {
	case codes.Unknown,
		codes.DeadlineExceeded,
		codes.ResourceExhausted,
		codes.Aborted,
		codes.Internal,
		codes.Unavailable,
		codes.DataLoss:
		return true
	default:
		return false
	}
}

// ============================================================================
// Traced Server Stream
// ============================================================================

// tracedServerStream wraps a grpc.ServerStream with tracing.
type tracedServerStream struct {
	grpc.ServerStream
	ctx  context.Context
	span tracing.Span
}

// Context returns the traced context.
func (s *tracedServerStream) Context() context.Context {
	return s.ctx
}

// SendMsg wraps SendMsg with tracing events.
func (s *tracedServerStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err != nil {
		s.span.AddEvent("message.sent.error", map[string]any{
			"error": err.Error(),
		})
	} else {
		s.span.AddEvent("message.sent", nil)
	}
	return err
}

// RecvMsg wraps RecvMsg with tracing events.
func (s *tracedServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err != nil {
		s.span.AddEvent("message.received.error", map[string]any{
			"error": err.Error(),
		})
	} else {
		s.span.AddEvent("message.received", nil)
	}
	return err
}

// ============================================================================
// Functional API (Alternative)
// ============================================================================

// UnaryTracingInterceptor returns a unary interceptor with functional options.
//
// Usage:
//
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(interceptors.UnaryTracingInterceptor(
//	        tracer,
//	        interceptors.WithSkipTracing("/grpc.health.v1.Health/Check"),
//	    )),
//	)
func UnaryTracingInterceptor(tracer tracing.Tracer, opts ...TracingOption) grpc.UnaryServerInterceptor {
	interceptor := NewTracingInterceptor(tracer, opts...)
	return interceptor.Unary()
}

// StreamTracingInterceptor returns a stream interceptor with functional options.
//
// Usage:
//
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(interceptors.StreamTracingInterceptor(
//	        tracer,
//	        interceptors.WithSkipTracing("/grpc.health.v1.Health/Watch"),
//	    )),
//	)
func StreamTracingInterceptor(tracer tracing.Tracer, opts ...TracingOption) grpc.StreamServerInterceptor {
	interceptor := NewTracingInterceptor(tracer, opts...)
	return interceptor.Stream()
}