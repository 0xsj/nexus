package log

import (
	"context"
	"io"
	"os"
)

// Logger is the interface for structured logging.
type Logger interface {
	// Debug logs a debug message.
	Debug(msg string, fields ...Field)

	// Info logs an info message.
	Info(msg string, fields ...Field)

	// Warn logs a warning message.
	Warn(msg string, fields ...Field)

	// Error logs an error message.
	Error(msg string, fields ...Field)

	// Fatal logs a fatal message and exits.
	Fatal(msg string, fields ...Field)

	// Log logs at the specified level.
	Log(level Level, msg string, fields ...Field)

	// With returns a logger with the given fields attached.
	With(fields ...Field) Logger

	// WithContext returns a logger that extracts fields from context.
	WithContext(ctx context.Context) Logger

	// Level returns the current log level.
	Level() Level

	// SetLevel sets the log level.
	SetLevel(level Level)

	// Enabled returns true if the given level is enabled.
	Enabled(level Level) bool
}

// ============================================================================
// Logger Options
// ============================================================================

// Options configures the logger.
type Options struct {
	// Level is the minimum log level.
	Level Level

	// Output is the writer for log output.
	Output io.Writer

	// ErrorOutput is the writer for error-level logs.
	// If nil, Output is used for all levels.
	ErrorOutput io.Writer

	// Colorize enables colorized output.
	Colorize bool

	// IncludeTimestamp includes timestamp in output.
	IncludeTimestamp bool

	// TimestampFormat is the format for timestamps.
	TimestampFormat string

	// IncludeCaller includes caller information.
	IncludeCaller bool

	// CallerSkip is the number of stack frames to skip for caller info.
	CallerSkip int

	// Fields are default fields included in every log.
	Fields Fields
}

// DefaultOptions returns default logger options.
func DefaultOptions() Options {
	return Options{
		Level:            LevelInfo,
		Output:           os.Stdout,
		ErrorOutput:      os.Stderr,
		Colorize:         true,
		IncludeTimestamp: true,
		TimestampFormat:  "15:04:05.000",
		IncludeCaller:    false,
		CallerSkip:       3,
		Fields:           nil,
	}
}

// WithLevel sets the log level.
func (o Options) WithLevel(level Level) Options {
	o.Level = level
	return o
}

// WithOutput sets the output writer.
func (o Options) WithOutput(w io.Writer) Options {
	o.Output = w
	return o
}

// WithErrorOutput sets the error output writer.
func (o Options) WithErrorOutput(w io.Writer) Options {
	o.ErrorOutput = w
	return o
}

// WithColorize enables/disables colorized output.
func (o Options) WithColorize(colorize bool) Options {
	o.Colorize = colorize
	return o
}

// WithTimestamp enables/disables timestamp.
func (o Options) WithTimestamp(include bool) Options {
	o.IncludeTimestamp = include
	return o
}

// WithTimestampFormat sets the timestamp format.
func (o Options) WithTimestampFormat(format string) Options {
	o.TimestampFormat = format
	return o
}

// WithCaller enables/disables caller information.
func (o Options) WithCaller(include bool) Options {
	o.IncludeCaller = include
	return o
}

// WithCallerSkip sets the caller skip frames.
func (o Options) WithCallerSkip(skip int) Options {
	o.CallerSkip = skip
	return o
}

// WithFields sets default fields.
func (o Options) WithFields(fields ...Field) Options {
	o.Fields = fields
	return o
}

// ============================================================================
// Global Logger
// ============================================================================

var globalLogger Logger = NewConsoleLogger(DefaultOptions())

// SetGlobal sets the global logger.
func SetGlobal(logger Logger) {
	globalLogger = logger
}

// Global returns the global logger.
func Global() Logger {
	return globalLogger
}

// Debug logs a debug message using the global logger.
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message using the global logger.
func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message using the global logger.
func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message using the global logger.
func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

// Fatal logs a fatal message using the global logger and exits.
func Fatal(msg string, fields ...Field) {
	globalLogger.Fatal(msg, fields...)
}

// With returns a global logger with fields attached.
func With(fields ...Field) Logger {
	return globalLogger.With(fields...)
}

// WithContext returns a global logger with context.
func WithContext(ctx context.Context) Logger {
	return globalLogger.WithContext(ctx)
}

// ============================================================================
// Helper Functions
// ============================================================================

// New creates a new console logger with default options.
func New() Logger {
	return NewConsoleLogger(DefaultOptions())
}

// NewWithLevel creates a new console logger with the specified level.
func NewWithLevel(level Level) Logger {
	return NewConsoleLogger(DefaultOptions().WithLevel(level))
}

// NewWithOptions creates a new console logger with custom options.
func NewWithOptions(opts Options) Logger {
	return NewConsoleLogger(opts)
}

// NewDevelopment creates a logger configured for development.
func NewDevelopment() Logger {
	return NewConsoleLogger(Options{
		Level:            LevelDebug,
		Output:           os.Stdout,
		ErrorOutput:      os.Stderr,
		Colorize:         true,
		IncludeTimestamp: true,
		TimestampFormat:  "15:04:05.000",
		IncludeCaller:    true,
		CallerSkip:       3,
	})
}

// NewProduction creates a logger configured for production.
func NewProduction() Logger {
	return NewConsoleLogger(Options{
		Level:            LevelInfo,
		Output:           os.Stdout,
		ErrorOutput:      os.Stderr,
		Colorize:         false,
		IncludeTimestamp: true,
		TimestampFormat:  "2006-01-02T15:04:05.000Z07:00",
		IncludeCaller:    false,
		CallerSkip:       3,
	})
}

// NewTest creates a logger for testing that discards output.
func NewTest() Logger {
	return NewNoop()
}
