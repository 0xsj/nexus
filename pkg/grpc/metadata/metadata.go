package metadata

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// Metadata header keys
const (
	RequestIDKey = "x-request-id"
	TraceIDKey   = "x-trace-id"
	SpanIDKey    = "x-span-id"
	UserAgentKey = "user-agent"
)

// ExtractRequestID gets request ID from incoming gRPC metadata
func ExtractRequestID(ctx context.Context) string {
	return extractFirst(ctx, RequestIDKey)
}

// ExtractTraceID gets trace ID from incoming gRPC metadata
func ExtractTraceID(ctx context.Context) string {
	return extractFirst(ctx, TraceIDKey)
}

// ExtractSpanID gets span ID from incoming gRPC metadata
func ExtractSpanID(ctx context.Context) string {
	return extractFirst(ctx, SpanIDKey)
}

// ExtractUserAgent gets user agent from incoming gRPC metadata
func ExtractUserAgent(ctx context.Context) string {
	return extractFirst(ctx, UserAgentKey)
}

// WithRequestID adds request ID to outgoing gRPC metadata
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return appendMetadata(ctx, RequestIDKey, requestID)
}

// WithTraceID adds trace ID to outgoing gRPC metadata
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return appendMetadata(ctx, TraceIDKey, traceID)
}

// WithSpanID adds span ID to outgoing gRPC metadata
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return appendMetadata(ctx, SpanIDKey, spanID)
}

// extractFirst gets the first value for a key from incoming metadata
func extractFirst(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

// appendMetadata appends a key-value pair to outgoing metadata
func appendMetadata(ctx context.Context, key, value string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}

	md = md.Copy()
	md.Append(key, value)

	return metadata.NewOutgoingContext(ctx, md)
}

// ExtractAll extracts all common metadata from context
func ExtractAll(ctx context.Context) map[string]string {
	return map[string]string{
		"request_id": ExtractRequestID(ctx),
		"trace_id":   ExtractTraceID(ctx),
		"span_id":    ExtractSpanID(ctx),
		"user_agent": ExtractUserAgent(ctx),
	}
}
