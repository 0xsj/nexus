package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ConsoleLogger is a colorized console logger.
type ConsoleLogger struct {
	mu          sync.Mutex
	opts        Options
	level       Level
	output      io.Writer
	errorOutput io.Writer
	fields      Fields
}

// NewConsoleLogger creates a new console logger.
func NewConsoleLogger(opts Options) *ConsoleLogger {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.ErrorOutput == nil {
		opts.ErrorOutput = opts.Output
	}

	return &ConsoleLogger{
		opts:        opts,
		level:       opts.Level,
		output:      opts.Output,
		errorOutput: opts.ErrorOutput,
		fields:      opts.Fields,
	}
}

// Debug logs a debug message.
func (l *ConsoleLogger) Debug(msg string, fields ...Field) {
	l.Log(LevelDebug, msg, fields...)
}

// Info logs an info message.
func (l *ConsoleLogger) Info(msg string, fields ...Field) {
	l.Log(LevelInfo, msg, fields...)
}

// Warn logs a warning message.
func (l *ConsoleLogger) Warn(msg string, fields ...Field) {
	l.Log(LevelWarn, msg, fields...)
}

// Error logs an error message.
func (l *ConsoleLogger) Error(msg string, fields ...Field) {
	l.Log(LevelError, msg, fields...)
}

// Fatal logs a fatal message and exits.
func (l *ConsoleLogger) Fatal(msg string, fields ...Field) {
	l.Log(LevelFatal, msg, fields...)
	os.Exit(1)
}

// Log logs at the specified level.
func (l *ConsoleLogger) Log(level Level, msg string, fields ...Field) {
	if !l.Enabled(level) {
		return
	}

	// Build the log line
	line := l.formatLine(level, msg, fields)

	// Select output writer
	w := l.output
	if level >= LevelError {
		w = l.errorOutput
	}

	// Write with mutex protection
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(w, line)
}

// With returns a logger with the given fields attached.
func (l *ConsoleLogger) With(fields ...Field) Logger {
	return &ConsoleLogger{
		opts:        l.opts,
		level:       l.level,
		output:      l.output,
		errorOutput: l.errorOutput,
		fields:      l.fields.Merge(fields),
	}
}

// WithContext returns a logger that extracts fields from context.
func (l *ConsoleLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	contextFields := ExtractContextFields(ctx)
	if len(contextFields) == 0 {
		return l
	}

	return l.With(contextFields...)
}

// Level returns the current log level.
func (l *ConsoleLogger) Level() Level {
	return l.level
}

// SetLevel sets the log level.
func (l *ConsoleLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Enabled returns true if the given level is enabled.
func (l *ConsoleLogger) Enabled(level Level) bool {
	return level.Enabled(l.level)
}

// ============================================================================
// Formatting
// ============================================================================

// formatLine formats a complete log line.
func (l *ConsoleLogger) formatLine(level Level, msg string, fields Fields) string {
	var b strings.Builder
	b.Grow(256)

	// Timestamp
	if l.opts.IncludeTimestamp {
		ts := time.Now().Format(l.opts.TimestampFormat)
		if l.opts.Colorize {
			b.WriteString(ColorizeTimestamp(ts))
		} else {
			b.WriteString(ts)
		}
		b.WriteString(" ")
	}

	// Level
	if l.opts.Colorize {
		b.WriteString(ColorizeLevelPadded(level))
	} else {
		b.WriteString(padRight(level.ShortString(), 5))
	}
	b.WriteString(" ")

	// Caller
	if l.opts.IncludeCaller {
		caller := l.getCaller()
		if l.opts.Colorize {
			b.WriteString(ColorizeCaller(caller))
		} else {
			b.WriteString(caller)
		}
		b.WriteString(" ")
	}

	// Message
	if l.opts.Colorize {
		b.WriteString(ColorizeMessage(level, msg))
	} else {
		b.WriteString(msg)
	}

	// Merge default fields with provided fields
	allFields := l.fields.Merge(fields)

	// Fields
	if len(allFields) > 0 {
		b.WriteString(" ")
		l.formatFields(&b, allFields)
	}

	return b.String()
}

// formatFields formats fields for output.
func (l *ConsoleLogger) formatFields(b *strings.Builder, fields Fields) {
	for i, field := range fields {
		if i > 0 {
			b.WriteString(" ")
		}

		// Handle nested groups
		if field.Type == FieldTypeGroup {
			if groupFields, ok := field.Value.([]Field); ok {
				if l.opts.Colorize {
					b.WriteString(ColorizeKey(field.Key))
				} else {
					b.WriteString(field.Key)
				}
				b.WriteString("={")
				l.formatFields(b, groupFields)
				b.WriteString("}")
				continue
			}
		}

		// Key
		if l.opts.Colorize {
			b.WriteString(ColorizeKey(field.Key))
		} else {
			b.WriteString(field.Key)
		}
		b.WriteString("=")

		// Value
		if l.opts.Colorize {
			b.WriteString(field.ColoredValue())
		} else {
			b.WriteString(l.formatValue(field))
		}
	}
}

// formatValue formats a field value without color.
func (l *ConsoleLogger) formatValue(field Field) string {
	switch field.Type {
	case FieldTypeString:
		return fmt.Sprintf("%q", field.Value)
	case FieldTypeError:
		if err, ok := field.Value.(error); ok && err != nil {
			return fmt.Sprintf("%q", err.Error())
		}
		return "<nil>"
	default:
		return field.StringValue()
	}
}

// getCaller returns the caller information.
func (l *ConsoleLogger) getCaller() string {
	_, file, line, ok := runtime.Caller(l.opts.CallerSkip)
	if !ok {
		return "???"
	}

	// Get just the filename, not the full path
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}

	return fmt.Sprintf("%s:%d", short, line)
}

// ============================================================================
// Pretty Print Helpers
// ============================================================================

// PrettyLogger creates a highly formatted logger for development.
func PrettyLogger() Logger {
	return NewConsoleLogger(Options{
		Level:            LevelDebug,
		Output:           os.Stdout,
		ErrorOutput:      os.Stderr,
		Colorize:         true,
		IncludeTimestamp: true,
		TimestampFormat:  "15:04:05",
		IncludeCaller:    true,
		CallerSkip:       3,
	})
}

// BoxMessage creates a boxed message for important logs.
func BoxMessage(msg string) string {
	width := len(msg) + 4
	border := strings.Repeat("─", width)

	return fmt.Sprintf("┌%s┐\n│  %s  │\n└%s┘", border, msg, border)
}

// Banner logs a banner message.
func Banner(logger Logger, msg string) {
	logger.Info(BoxMessage(msg))
}

// Section logs a section header.
func Section(logger Logger, msg string) {
	separator := strings.Repeat("─", 40)
	logger.Info(separator)
	logger.Info(msg)
	logger.Info(separator)
}

// ============================================================================
// Specialized Loggers
// ============================================================================

// ComponentLogger creates a logger with a component field.
func ComponentLogger(component string) Logger {
	return globalLogger.With(Component(component))
}

// RequestLogger creates a logger for HTTP requests.
func RequestLogger(requestID string) Logger {
	return globalLogger.With(RequestID(requestID))
}

// OperationLogger creates a logger for a specific operation.
func OperationLogger(operation string) Logger {
	return globalLogger.With(Operation(operation))
}
