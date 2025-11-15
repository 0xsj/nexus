package logger

import (
	"context"
	"io"
)

// Logger provides structured logging with context propagation and level-based filtering.
// All methods are safe for concurrent use.
type Logger interface {
	// Core logging methods with structured fields
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field) // Calls os.Exit(1)
	Panic(msg string, fields ...Field) // Calls panic()

	// Context-aware logging methods
	// These methods extract trace/span IDs from context if available
	DebugContext(ctx context.Context, msg string, fields ...Field)
	InfoContext(ctx context.Context, msg string, fields ...Field)
	WarnContext(ctx context.Context, msg string, fields ...Field)
	ErrorContext(ctx context.Context, msg string, fields ...Field)

	// With returns a new logger with additional fields attached.
	// The original logger is unchanged (immutable pattern).
	With(fields ...Field) Logger

	// WithError returns a new logger with an error field attached.
	// This is a convenience method equivalent to With(Err(err)).
	// If includeStackTrace is true, also captures the stack trace.
	WithError(err error) Logger

	// WithContext returns a new logger that will extract tracing information
	// from the provided context on each log call.
	WithContext(ctx context.Context) Logger

	// Named creates a child logger with the given name.
	// Useful for creating subsystem-specific loggers.
	// Example: logger.Named("http"), logger.Named("database")
	Named(name string) Logger

	// Level returns the minimum enabled log level.
	Level() Level

	// Sync flushes any buffered log entries.
	// Applications should call Sync before exiting.
	Sync() error

	// Clone creates a copy of the logger.
	// Useful for creating independent logger instances.
	Clone() Logger
}

// Config holds configuration for logger creation.
type Config struct {
	// Level is the minimum enabled logging level.
	Level Level

	// Development puts the logger in development mode, which changes
	// the behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool

	// Encoding sets the logger's encoding. Valid values are "json" and "console".
	Encoding string

	// OutputPaths is a list of URLs or file paths to write logging output to.
	// See Open for details.
	OutputPaths []string

	// ErrorOutputPaths is a list of URLs to write internal logger errors to.
	ErrorOutputPaths []string

	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool

	// DisableStacktrace completely disables automatic stacktrace capturing.
	DisableStacktrace bool

	// StacktraceLevel is the level at and above which stacktraces are captured.
	// By default, stacktraces are captured for ErrorLevel and above.
	StacktraceLevel Level

	// AutoStackTraceOnError automatically adds stack traces when using WithError().
	// Default: true in development, false in production.
	AutoStackTraceOnError bool

	// TimeEncoder sets how timestamps are encoded. Options: "iso8601", "millis", "nanos", "epoch"
	TimeEncoder string

	// CallerEncoder sets how caller information is encoded. Options: "short", "full"
	CallerEncoder string

	// Sampling configures log sampling to reduce output volume.
	Sampling *SamplingConfig

	// InitialFields are fields to add to the root logger.
	InitialFields map[string]interface{}

	// ColorEnabled enables colored output for console encoding.
	// Automatically disabled when output is not a TTY.
	ColorEnabled bool
}

// SamplingConfig configures log sampling to reduce volume.
type SamplingConfig struct {
	// Initial is the number of messages to log per second before sampling begins.
	Initial int

	// Thereafter, the logger will log every Nth message after the initial count.
	Thereafter int
}

// OutputConfig represents a single output destination.
type OutputConfig struct {
	// Writer is the destination for log output.
	Writer io.Writer

	// Level is the minimum level for this output.
	// If not set, uses the logger's global level.
	Level *Level

	// Encoder determines the output format for this destination.
	Encoder string // "json" or "console"
}

// NewConfig creates a default configuration suitable for production.
func NewConfig() Config {
	return Config{
		Level:                 InfoLevel,
		Development:           false,
		Encoding:              "json",
		OutputPaths:           []string{"stdout"},
		ErrorOutputPaths:      []string{"stderr"},
		DisableCaller:         false,
		DisableStacktrace:     false,
		StacktraceLevel:       ErrorLevel,
		AutoStackTraceOnError: false,
		TimeEncoder:           "iso8601",
		CallerEncoder:         "short",
		ColorEnabled:          false,
	}
}

// NewDevelopmentConfig creates a default configuration suitable for development.
func NewDevelopmentConfig() Config {
	return Config{
		Level:                 DebugLevel,
		Development:           true,
		Encoding:              "console",
		OutputPaths:           []string{"stdout"},
		ErrorOutputPaths:      []string{"stderr"},
		DisableCaller:         false,
		DisableStacktrace:     false,
		StacktraceLevel:       WarnLevel,
		AutoStackTraceOnError: true,
		TimeEncoder:           "iso8601",
		CallerEncoder:         "short",
		ColorEnabled:          true,
	}
}

// ContextKey is the type for context keys used by the logger.
type ContextKey string

const (
	// TraceIDKey is the context key for trace IDs.
	TraceIDKey ContextKey = "trace_id"

	// SpanIDKey is the context key for span IDs.
	SpanIDKey ContextKey = "span_id"

	// RequestIDKey is the context key for request IDs.
	RequestIDKey ContextKey = "request_id"

	// UserIDKey is the context key for user IDs.
	UserIDKey ContextKey = "user_id"
)

// WithTraceID adds a trace ID to the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithSpanID adds a span ID to the context.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID adds a user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// ExtractTraceID extracts the trace ID from the context.
func ExtractTraceID(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	return traceID, ok
}

// ExtractSpanID extracts the span ID from the context.
func ExtractSpanID(ctx context.Context) (string, bool) {
	spanID, ok := ctx.Value(SpanIDKey).(string)
	return spanID, ok
}

// ExtractRequestID extracts the request ID from the context.
func ExtractRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}

// ExtractUserID extracts the user ID from the context.
func ExtractUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// ExtractContextFields extracts all logger-relevant fields from the context.
func ExtractContextFields(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}

	var fields []Field

	if traceID, ok := ExtractTraceID(ctx); ok {
		fields = append(fields, String("trace_id", traceID))
	}

	if spanID, ok := ExtractSpanID(ctx); ok {
		fields = append(fields, String("span_id", spanID))
	}

	if requestID, ok := ExtractRequestID(ctx); ok {
		fields = append(fields, String("request_id", requestID))
	}

	if userID, ok := ExtractUserID(ctx); ok {
		fields = append(fields, String("user_id", userID))
	}

	return fields
}

// NopLogger returns a logger that does nothing.
// Useful for testing or when logging should be disabled.
func NopLogger() Logger {
	return &nopLogger{}
}

// nopLogger is a no-op implementation of Logger.
type nopLogger struct{}

func (n *nopLogger) Debug(msg string, fields ...Field)                             {}
func (n *nopLogger) Info(msg string, fields ...Field)                              {}
func (n *nopLogger) Warn(msg string, fields ...Field)                              {}
func (n *nopLogger) Error(msg string, fields ...Field)                             {}
func (n *nopLogger) Fatal(msg string, fields ...Field)                             {}
func (n *nopLogger) Panic(msg string, fields ...Field)                             {}
func (n *nopLogger) DebugContext(ctx context.Context, msg string, fields ...Field) {}
func (n *nopLogger) InfoContext(ctx context.Context, msg string, fields ...Field)  {}
func (n *nopLogger) WarnContext(ctx context.Context, msg string, fields ...Field)  {}
func (n *nopLogger) ErrorContext(ctx context.Context, msg string, fields ...Field) {}
func (n *nopLogger) With(fields ...Field) Logger                                   { return n }
func (n *nopLogger) WithError(err error) Logger                                    { return n }
func (n *nopLogger) WithContext(ctx context.Context) Logger                        { return n }
func (n *nopLogger) Named(name string) Logger                                      { return n }
func (n *nopLogger) Level() Level                                                  { return PanicLevel }
func (n *nopLogger) Sync() error                                                   { return nil }
func (n *nopLogger) Clone() Logger                                                 { return n }
