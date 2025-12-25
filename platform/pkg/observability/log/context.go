package log

import (
	"context"
)

// Context keys for log fields.
type contextKey string

const (
	// loggerKey is the context key for the logger.
	loggerKey contextKey = "logger"

	// fieldsKey is the context key for log fields.
	fieldsKey contextKey = "log_fields"

	// requestIDKey is the context key for request ID.
	requestIDKey contextKey = "request_id"

	// traceIDKey is the context key for trace ID.
	traceIDKey contextKey = "trace_id"

	// spanIDKey is the context key for span ID.
	spanIDKey contextKey = "span_id"

	// userIDKey is the context key for user ID.
	userIDKey contextKey = "user_id"
)

// ============================================================================
// Logger Context
// ============================================================================

// WithLogger returns a context with the logger attached.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns the logger from context, or the global logger if not found.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return globalLogger
	}
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return globalLogger
}

// ============================================================================
// Fields Context
// ============================================================================

// ContextWithFields returns a context with additional log fields.
func ContextWithFields(ctx context.Context, fields ...Field) context.Context {
	existing := FieldsFromContext(ctx)
	merged := existing.Merge(fields)
	return context.WithValue(ctx, fieldsKey, merged)
}

// FieldsFromContext returns log fields from the context.
func FieldsFromContext(ctx context.Context) Fields {
	if ctx == nil {
		return nil
	}
	if fields, ok := ctx.Value(fieldsKey).(Fields); ok {
		return fields
	}
	return nil
}

// ============================================================================
// Common Context Values
// ============================================================================

// ContextWithRequestID returns a context with a request ID.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	return ContextWithFields(ctx, RequestID(requestID))
}

// RequestIDFromContext returns the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// ContextWithTraceID returns a context with a trace ID.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	ctx = context.WithValue(ctx, traceIDKey, traceID)
	return ContextWithFields(ctx, TraceID(traceID))
}

// TraceIDFromContext returns the trace ID from context.
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// ContextWithSpanID returns a context with a span ID.
func ContextWithSpanID(ctx context.Context, spanID string) context.Context {
	ctx = context.WithValue(ctx, spanIDKey, spanID)
	return ContextWithFields(ctx, SpanID(spanID))
}

// SpanIDFromContext returns the span ID from context.
func SpanIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(spanIDKey).(string); ok {
		return id
	}
	return ""
}

// ContextWithUserID returns a context with a user ID.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	return ContextWithFields(ctx, UserID(userID))
}

// UserIDFromContext returns the user ID from context.
func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// ============================================================================
// Context Logging Shortcuts
// ============================================================================

// Ctx returns a logger from context, ready to use.
// This is a convenience shortcut for FromContext(ctx).
func Ctx(ctx context.Context) Logger {
	return FromContext(ctx)
}

// DebugCtx logs a debug message using the logger from context.
func DebugCtx(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).WithContext(ctx).Debug(msg, fields...)
}

// InfoCtx logs an info message using the logger from context.
func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).WithContext(ctx).Info(msg, fields...)
}

// WarnCtx logs a warning message using the logger from context.
func WarnCtx(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).WithContext(ctx).Warn(msg, fields...)
}

// ErrorCtx logs an error message using the logger from context.
func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).WithContext(ctx).Error(msg, fields...)
}

// ============================================================================
// Extract All Context Fields
// ============================================================================

// ExtractContextFields extracts all known fields from context.
func ExtractContextFields(ctx context.Context) Fields {
	if ctx == nil {
		return nil
	}

	var fields Fields

	// Extract stored fields
	if stored := FieldsFromContext(ctx); stored != nil {
		fields = fields.Merge(stored)
	}

	return fields
}
