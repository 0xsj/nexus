package zap

import (
	"context"
	"time"

	"github.com/0xsj/nexus/pkg/observability/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger is the Zap implementation of the Logger interface.
type zapLogger struct {
	zap                   *zap.Logger
	level                 logger.Level
	autoStackTraceOnError bool
}

// NewZapLogger creates a new Zap logger with the given configuration.
func NewZapLogger(cfg logger.Config) (logger.Logger, error) {
	// Build the Zap configuration
	zapCfg := buildZapConfig(cfg)

	// Build the logger
	zapLog, err := zapCfg.Build(buildOptions(cfg)...)
	if err != nil {
		return nil, err
	}

	return &zapLogger{
		zap:                   zapLog,
		level:                 cfg.Level,
		autoStackTraceOnError: cfg.AutoStackTraceOnError,
	}, nil
}

// NewZapLoggerFromCore creates a Zap logger from a custom core.
// Useful for testing or advanced configuration.
func NewZapLoggerFromCore(core zapcore.Core, cfg logger.Config) logger.Logger {
	zapLog := zap.New(core, buildOptions(cfg)...)

	return &zapLogger{
		zap:                   zapLog,
		level:                 cfg.Level,
		autoStackTraceOnError: cfg.AutoStackTraceOnError,
	}
}

// Debug logs a debug message with structured fields.
func (z *zapLogger) Debug(msg string, fields ...logger.Field) {
	if z.zap.Core().Enabled(zapcore.DebugLevel) {
		z.zap.Debug(msg, convertFields(fields)...)
	}
}

// Info logs an info message with structured fields.
func (z *zapLogger) Info(msg string, fields ...logger.Field) {
	if z.zap.Core().Enabled(zapcore.InfoLevel) {
		z.zap.Info(msg, convertFields(fields)...)
	}
}

// Warn logs a warning message with structured fields.
func (z *zapLogger) Warn(msg string, fields ...logger.Field) {
	if z.zap.Core().Enabled(zapcore.WarnLevel) {
		z.zap.Warn(msg, convertFields(fields)...)
	}
}

// Error logs an error message with structured fields.
func (z *zapLogger) Error(msg string, fields ...logger.Field) {
	if z.zap.Core().Enabled(zapcore.ErrorLevel) {
		z.zap.Error(msg, convertFields(fields)...)
	}
}

// Fatal logs a fatal message with structured fields and then calls os.Exit(1).
func (z *zapLogger) Fatal(msg string, fields ...logger.Field) {
	z.zap.Fatal(msg, convertFields(fields)...)
}

// Panic logs a panic message with structured fields and then panics.
func (z *zapLogger) Panic(msg string, fields ...logger.Field) {
	z.zap.Panic(msg, convertFields(fields)...)
}

// DebugContext logs a debug message with context and structured fields.
func (z *zapLogger) DebugContext(ctx context.Context, msg string, fields ...logger.Field) {
	if z.zap.Core().Enabled(zapcore.DebugLevel) {
		contextFields := logger.ExtractContextFields(ctx)
		allFields := append(contextFields, fields...)
		z.zap.Debug(msg, convertFields(allFields)...)
	}
}

// InfoContext logs an info message with context and structured fields.
func (z *zapLogger) InfoContext(ctx context.Context, msg string, fields ...logger.Field) {
	if z.zap.Core().Enabled(zapcore.InfoLevel) {
		contextFields := logger.ExtractContextFields(ctx)
		allFields := append(contextFields, fields...)
		z.zap.Info(msg, convertFields(allFields)...)
	}
}

// WarnContext logs a warning message with context and structured fields.
func (z *zapLogger) WarnContext(ctx context.Context, msg string, fields ...logger.Field) {
	if z.zap.Core().Enabled(zapcore.WarnLevel) {
		contextFields := logger.ExtractContextFields(ctx)
		allFields := append(contextFields, fields...)
		z.zap.Warn(msg, convertFields(allFields)...)
	}
}

// ErrorContext logs an error message with context and structured fields.
func (z *zapLogger) ErrorContext(ctx context.Context, msg string, fields ...logger.Field) {
	if z.zap.Core().Enabled(zapcore.ErrorLevel) {
		contextFields := logger.ExtractContextFields(ctx)
		allFields := append(contextFields, fields...)
		z.zap.Error(msg, convertFields(allFields)...)
	}
}

// With returns a new logger with additional fields attached.
func (z *zapLogger) With(fields ...logger.Field) logger.Logger {
	return &zapLogger{
		zap:                   z.zap.With(convertFields(fields)...),
		level:                 z.level,
		autoStackTraceOnError: z.autoStackTraceOnError,
	}
}

// WithError returns a new logger with an error field attached.
// If AutoStackTraceOnError is enabled, it also adds a stack trace.
func (z *zapLogger) WithError(err error) logger.Logger {
	if err == nil {
		return z
	}

	fields := []logger.Field{logger.Err(err)}

	// Add stack trace if configured
	if z.autoStackTraceOnError {
		fields = append(fields, logger.StackTrace("stacktrace"))
	}

	return &zapLogger{
		zap:                   z.zap.With(convertFields(fields)...),
		level:                 z.level,
		autoStackTraceOnError: z.autoStackTraceOnError,
	}
}

// WithContext returns a new logger that will extract tracing information
// from the provided context on each log call.
func (z *zapLogger) WithContext(ctx context.Context) logger.Logger {
	fields := logger.ExtractContextFields(ctx)
	if len(fields) == 0 {
		return z
	}

	return &zapLogger{
		zap:                   z.zap.With(convertFields(fields)...),
		level:                 z.level,
		autoStackTraceOnError: z.autoStackTraceOnError,
	}
}

// Named creates a child logger with the given name.
func (z *zapLogger) Named(name string) logger.Logger {
	return &zapLogger{
		zap:                   z.zap.Named(name),
		level:                 z.level,
		autoStackTraceOnError: z.autoStackTraceOnError,
	}
}

// Level returns the minimum enabled log level.
func (z *zapLogger) Level() logger.Level {
	return z.level
}

// Sync flushes any buffered log entries.
func (z *zapLogger) Sync() error {
	return z.zap.Sync()
}

// Clone creates a copy of the logger.
func (z *zapLogger) Clone() logger.Logger {
	return &zapLogger{
		zap:                   z.zap,
		level:                 z.level,
		autoStackTraceOnError: z.autoStackTraceOnError,
	}
}

// convertFields converts our logger.Field slice to zap.Field slice.
func convertFields(fields []logger.Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields))

	for _, field := range fields {
		// Skip no-op fields
		if field.Type == logger.SkipType {
			continue
		}

		zapFields = append(zapFields, convertField(field))
	}

	return zapFields
}

// convertField converts a single logger.Field to zap.Field.
func convertField(field logger.Field) zap.Field {
	switch field.Type {
	case logger.BoolType:
		return zap.Bool(field.Key, field.Integer == 1)

	case logger.IntType:
		return zap.Int(field.Key, int(field.Integer))

	case logger.Int8Type:
		return zap.Int8(field.Key, int8(field.Integer))

	case logger.Int16Type:
		return zap.Int16(field.Key, int16(field.Integer))

	case logger.Int32Type:
		return zap.Int32(field.Key, int32(field.Integer))

	case logger.Int64Type:
		return zap.Int64(field.Key, field.Integer)

	case logger.UintType:
		return zap.Uint(field.Key, uint(field.Integer))

	case logger.Uint8Type:
		return zap.Uint8(field.Key, uint8(field.Integer))

	case logger.Uint16Type:
		return zap.Uint16(field.Key, uint16(field.Integer))

	case logger.Uint32Type:
		return zap.Uint32(field.Key, uint32(field.Integer))

	case logger.Uint64Type:
		return zap.Uint64(field.Key, uint64(field.Integer))

	case logger.Float32Type:
		return zap.Float32(field.Key, float32(field.Float))

	case logger.Float64Type:
		return zap.Float64(field.Key, field.Float)

	case logger.StringType:
		return zap.String(field.Key, field.Str)

	case logger.TimeType:
		if t, ok := field.Interface.(time.Time); ok {
			return zap.Time(field.Key, t)
		}
		return zap.Any(field.Key, field.Interface)

	case logger.DurationType:
		return zap.Duration(field.Key, time.Duration(field.Integer))

	case logger.ErrorType:
		if err, ok := field.Interface.(error); ok {
			return zap.Error(err)
		}
		return zap.String(field.Key, "<nil>")

	case logger.StringsType:
		if strs, ok := field.Interface.([]string); ok {
			return zap.Strings(field.Key, strs)
		}
		return zap.Any(field.Key, field.Interface)

	case logger.IntsType:
		if ints, ok := field.Interface.([]int); ok {
			return zap.Ints(field.Key, ints)
		}
		return zap.Any(field.Key, field.Interface)

	case logger.StackTraceType:
		return zap.String(field.Key, field.Str)

	case logger.AnyType:
		return zap.Any(field.Key, field.Interface)

	default:
		return zap.Any(field.Key, field.Interface)
	}
}

// MustNewZapLogger creates a new Zap logger and panics on error.
// Useful for application initialization where logger failure is fatal.
func MustNewZapLogger(cfg logger.Config) logger.Logger {
	log, err := NewZapLogger(cfg)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	return log
}

// NewDevelopmentLogger creates a logger suitable for development.
// This is a convenience function that uses NewDevelopmentConfig.
func NewDevelopmentLogger() (logger.Logger, error) {
	return NewZapLogger(logger.NewDevelopmentConfig())
}

// NewProductionLogger creates a logger suitable for production.
// This is a convenience function that uses NewConfig.
func NewProductionLogger() (logger.Logger, error) {
	return NewZapLogger(logger.NewConfig())
}

// NewTestLogger creates a logger that writes to /dev/null.
// Useful for testing when you don't want log output.
func NewTestLogger() logger.Logger {
	core := zapcore.NewNopCore()
	return &zapLogger{
		zap:                   zap.New(core),
		level:                 logger.InfoLevel,
		autoStackTraceOnError: false,
	}
}
