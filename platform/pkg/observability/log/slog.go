package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// SlogLogger wraps slog.Logger to implement our Logger interface.
type SlogLogger struct {
	logger *slog.Logger
	level  Level
	fields Fields
}

// NewSlogLogger creates a new logger using slog.
func NewSlogLogger(opts Options) *SlogLogger {
	// Convert our level to slog level
	slogLevel := levelToSlog(opts.Level)

	// Create handler options
	handlerOpts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: opts.IncludeCaller,
	}

	// Create handler based on options
	var handler slog.Handler
	output := opts.Output
	if output == nil {
		output = os.Stdout
	}

	if opts.Colorize {
		// Use our custom colored handler
		handler = NewColoredSlogHandler(output, handlerOpts)
	} else {
		// Use default JSON handler for production
		handler = slog.NewJSONHandler(output, handlerOpts)
	}

	return &SlogLogger{
		logger: slog.New(handler),
		level:  opts.Level,
		fields: opts.Fields,
	}
}

// Debug logs a debug message.
func (l *SlogLogger) Debug(msg string, fields ...Field) {
	l.Log(LevelDebug, msg, fields...)
}

// Info logs an info message.
func (l *SlogLogger) Info(msg string, fields ...Field) {
	l.Log(LevelInfo, msg, fields...)
}

// Warn logs a warning message.
func (l *SlogLogger) Warn(msg string, fields ...Field) {
	l.Log(LevelWarn, msg, fields...)
}

// Error logs an error message.
func (l *SlogLogger) Error(msg string, fields ...Field) {
	l.Log(LevelError, msg, fields...)
}

// Fatal logs a fatal message and exits.
func (l *SlogLogger) Fatal(msg string, fields ...Field) {
	l.Log(LevelFatal, msg, fields...)
	os.Exit(1)
}

// Log logs at the specified level.
func (l *SlogLogger) Log(level Level, msg string, fields ...Field) {
	if !l.Enabled(level) {
		return
	}

	// Merge default fields with provided fields
	allFields := l.fields.Merge(fields)

	// Convert to slog attrs
	attrs := fieldsToSlogAttrs(allFields)

	// Log at appropriate level
	l.logger.LogAttrs(context.Background(), levelToSlog(level), msg, attrs...)
}

// With returns a logger with the given fields attached.
func (l *SlogLogger) With(fields ...Field) Logger {
	return &SlogLogger{
		logger: l.logger,
		level:  l.level,
		fields: l.fields.Merge(fields),
	}
}

// WithContext returns a logger that extracts fields from context.
func (l *SlogLogger) WithContext(ctx context.Context) Logger {
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
func (l *SlogLogger) Level() Level {
	return l.level
}

// SetLevel sets the log level.
func (l *SlogLogger) SetLevel(level Level) {
	l.level = level
}

// Enabled returns true if the given level is enabled.
func (l *SlogLogger) Enabled(level Level) bool {
	return level.Enabled(l.level)
}

// ============================================================================
// Level Conversion
// ============================================================================

// levelToSlog converts our Level to slog.Level.
func levelToSlog(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelFatal:
		return slog.LevelError + 4 // Custom level above error
	default:
		return slog.LevelInfo
	}
}

// slogToLevel converts slog.Level to our Level.
func slogToLevel(level slog.Level) Level {
	switch {
	case level < slog.LevelInfo:
		return LevelDebug
	case level < slog.LevelWarn:
		return LevelInfo
	case level < slog.LevelError:
		return LevelWarn
	case level < slog.LevelError+4:
		return LevelError
	default:
		return LevelFatal
	}
}

// ============================================================================
// Field Conversion
// ============================================================================

// fieldsToSlogAttrs converts our Fields to slog.Attr slice.
func fieldsToSlogAttrs(fields Fields) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(fields))

	for _, field := range fields {
		attrs = append(attrs, fieldToSlogAttr(field))
	}

	return attrs
}

// fieldToSlogAttr converts a single Field to slog.Attr.
func fieldToSlogAttr(field Field) slog.Attr {
	switch field.Type {
	case FieldTypeString:
		return slog.String(field.Key, field.Value.(string))
	case FieldTypeInt:
		return slog.Int(field.Key, field.Value.(int))
	case FieldTypeInt64:
		return slog.Int64(field.Key, field.Value.(int64))
	case FieldTypeUint64:
		return slog.Uint64(field.Key, field.Value.(uint64))
	case FieldTypeFloat64:
		return slog.Float64(field.Key, field.Value.(float64))
	case FieldTypeBool:
		return slog.Bool(field.Key, field.Value.(bool))
	case FieldTypeTime:
		return slog.Time(field.Key, field.Value.(time.Time))
	case FieldTypeDuration:
		return slog.Duration(field.Key, field.Value.(time.Duration))
	case FieldTypeError:
		if err, ok := field.Value.(error); ok && err != nil {
			return slog.String(field.Key, err.Error())
		}
		return slog.String(field.Key, "<nil>")
	case FieldTypeGroup:
		if groupFields, ok := field.Value.([]Field); ok {
			groupAttrs := fieldsToSlogAttrs(groupFields)
			return slog.Group(field.Key, attrsToAny(groupAttrs)...)
		}
		return slog.Any(field.Key, field.Value)
	default:
		return slog.Any(field.Key, field.Value)
	}
}

// attrsToAny converts []slog.Attr to []any for slog.Group.
func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}
	return result
}

// ============================================================================
// Colored Slog Handler
// ============================================================================

// ColoredSlogHandler is a slog handler that outputs colored text.
type ColoredSlogHandler struct {
	opts   *slog.HandlerOptions
	output io.Writer
	attrs  []slog.Attr
	group  string
}

// NewColoredSlogHandler creates a new colored slog handler.
func NewColoredSlogHandler(output io.Writer, opts *slog.HandlerOptions) *ColoredSlogHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ColoredSlogHandler{
		opts:   opts,
		output: output,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *ColoredSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle handles the Record.
func (h *ColoredSlogHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.Grow(256)

	// Timestamp
	ts := r.Time.Format("15:04:05.000")
	b.WriteString(ColorizeTimestamp(ts))
	b.WriteString(" ")

	// Level
	level := slogToLevel(r.Level)
	b.WriteString(ColorizeLevelPadded(level))
	b.WriteString(" ")

	// Message
	b.WriteString(ColorizeMessage(level, r.Message))

	// Pre-defined attrs
	for _, attr := range h.attrs {
		b.WriteString(" ")
		h.appendAttr(&b, attr)
	}

	// Record attrs
	r.Attrs(func(attr slog.Attr) bool {
		b.WriteString(" ")
		h.appendAttr(&b, attr)
		return true
	})

	b.WriteString("\n")

	_, err := h.output.Write([]byte(b.String()))
	return err
}

// WithAttrs returns a new handler with the given attributes.
func (h *ColoredSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &ColoredSlogHandler{
		opts:   h.opts,
		output: h.output,
		attrs:  newAttrs,
		group:  h.group,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *ColoredSlogHandler) WithGroup(name string) slog.Handler {
	return &ColoredSlogHandler{
		opts:   h.opts,
		output: h.output,
		attrs:  h.attrs,
		group:  name,
	}
}

// appendAttr appends a colored attribute to the builder.
func (h *ColoredSlogHandler) appendAttr(b *strings.Builder, attr slog.Attr) {
	// Handle groups
	if attr.Value.Kind() == slog.KindGroup {
		attrs := attr.Value.Group()
		if len(attrs) == 0 {
			return
		}
		b.WriteString(ColorizeKey(attr.Key))
		b.WriteString("={")
		for i, a := range attrs {
			if i > 0 {
				b.WriteString(" ")
			}
			h.appendAttr(b, a)
		}
		b.WriteString("}")
		return
	}

	// Key
	b.WriteString(ColorizeKey(attr.Key))
	b.WriteString("=")

	// Value
	b.WriteString(h.colorizeValue(attr.Value))
}

// colorizeValue returns a colored string representation of the value.
func (h *ColoredSlogHandler) colorizeValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return ColorizeString(v.String())
	case slog.KindInt64:
		return ColorizeNumber(fmt.Sprintf("%d", v.Int64()))
	case slog.KindUint64:
		return ColorizeNumber(fmt.Sprintf("%d", v.Uint64()))
	case slog.KindFloat64:
		return ColorizeNumber(fmt.Sprintf("%.3f", v.Float64()))
	case slog.KindBool:
		return ColorizeBool(fmt.Sprintf("%t", v.Bool()))
	case slog.KindDuration:
		return ColorizeNumber(v.Duration().String())
	case slog.KindTime:
		return ColorizeTimestamp(v.Time().Format(time.RFC3339))
	default:
		return fmt.Sprintf("%v", v.Any())
	}
}
