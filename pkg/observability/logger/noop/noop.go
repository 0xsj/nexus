package noop

import (
	"context"

	"github.com/0xsj/nexus/pkg/observability/logger"
)

// nopLogger is a no-op implementation of Logger.
// It discards all log messages and does nothing.
// Useful for testing or when logging should be disabled.
type nopLogger struct{}

// NewNopLogger returns a logger that does nothing.
func NewNopLogger() logger.Logger {
	return &nopLogger{}
}

// Debug does nothing.
func (n *nopLogger) Debug(msg string, fields ...logger.Field) {}

// Info does nothing.
func (n *nopLogger) Info(msg string, fields ...logger.Field) {}

// Warn does nothing.
func (n *nopLogger) Warn(msg string, fields ...logger.Field) {}

// Error does nothing.
func (n *nopLogger) Error(msg string, fields ...logger.Field) {}

// Fatal does nothing (does NOT call os.Exit).
func (n *nopLogger) Fatal(msg string, fields ...logger.Field) {}

// Panic does nothing (does NOT panic).
func (n *nopLogger) Panic(msg string, fields ...logger.Field) {}

// DebugContext does nothing.
func (n *nopLogger) DebugContext(ctx context.Context, msg string, fields ...logger.Field) {}

// InfoContext does nothing.
func (n *nopLogger) InfoContext(ctx context.Context, msg string, fields ...logger.Field) {}

// WarnContext does nothing.
func (n *nopLogger) WarnContext(ctx context.Context, msg string, fields ...logger.Field) {}

// ErrorContext does nothing.
func (n *nopLogger) ErrorContext(ctx context.Context, msg string, fields ...logger.Field) {}

// With returns the same no-op logger.
func (n *nopLogger) With(fields ...logger.Field) logger.Logger {
	return n
}

// WithError returns the same no-op logger.
func (n *nopLogger) WithError(err error) logger.Logger {
	return n
}

// WithContext returns the same no-op logger.
func (n *nopLogger) WithContext(ctx context.Context) logger.Logger {
	return n
}

// Named returns the same no-op logger.
func (n *nopLogger) Named(name string) logger.Logger {
	return n
}

// Level returns PanicLevel (highest level, effectively disables all logging).
func (n *nopLogger) Level() logger.Level {
	return logger.PanicLevel
}

// Sync does nothing and returns nil.
func (n *nopLogger) Sync() error {
	return nil
}

// Clone returns the same no-op logger.
func (n *nopLogger) Clone() logger.Logger {
	return n
}
