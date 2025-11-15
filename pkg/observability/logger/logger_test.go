package logger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/nexus/pkg/observability/logger/noop"
	"github.com/0xsj/nexus/pkg/observability/logger/zap"
)

// TestNopLogger tests the no-op logger implementation.
func TestNopLogger(t *testing.T) {
	log := logger.NopLogger()

	// All these should not panic or cause issues
	log.Debug("debug message")
	log.Info("info message", logger.String("key", "value"))
	log.Warn("warn message")
	log.Error("error message", logger.Err(errors.New("test error")))

	// Fatal and Panic should NOT exit or panic in nop logger
	log.Fatal("fatal message")
	log.Panic("panic message")

	// Test context methods
	ctx := context.Background()
	log.DebugContext(ctx, "debug with context")
	log.InfoContext(ctx, "info with context")

	// Test immutability
	log2 := log.With(logger.String("field", "value"))
	if log2 == nil {
		t.Fatal("With() returned nil")
	}

	log3 := log.WithError(errors.New("test"))
	if log3 == nil {
		t.Fatal("WithError() returned nil")
	}

	log4 := log.Named("test")
	if log4 == nil {
		t.Fatal("Named() returned nil")
	}

	// Test level
	if log.Level() != logger.PanicLevel {
		t.Errorf("Expected PanicLevel, got %v", log.Level())
	}

	// Test sync
	if err := log.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}

// TestNoopPackage tests the standalone noop package.
func TestNoopPackage(t *testing.T) {
	log := noop.NewNopLogger()

	if log == nil {
		t.Fatal("NewNopLogger() returned nil")
	}

	// Should work identically to logger.NopLogger()
	log.Info("test message")

	if err := log.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}

// TestLevelString tests log level string representations.
func TestLevelString(t *testing.T) {
	tests := []struct {
		level    logger.Level
		expected string
	}{
		{logger.DebugLevel, "DEBUG"},
		{logger.InfoLevel, "INFO"},
		{logger.WarnLevel, "WARN"},
		{logger.ErrorLevel, "ERROR"},
		{logger.FatalLevel, "FATAL"},
		{logger.PanicLevel, "PANIC"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLevelParsing tests parsing log levels from strings.
func TestLevelParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected logger.Level
		wantErr  bool
	}{
		{"debug", logger.DebugLevel, false},
		{"DEBUG", logger.DebugLevel, false},
		{"info", logger.InfoLevel, false},
		{"INFO", logger.InfoLevel, false},
		{"warn", logger.WarnLevel, false},
		{"WARN", logger.WarnLevel, false},
		{"warning", logger.WarnLevel, false},
		{"error", logger.ErrorLevel, false},
		{"ERROR", logger.ErrorLevel, false},
		{"fatal", logger.FatalLevel, false},
		{"panic", logger.PanicLevel, false},
		{"invalid", logger.InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := logger.ParseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestLevelEnabled tests level filtering.
func TestLevelEnabled(t *testing.T) {
	tests := []struct {
		base     logger.Level
		target   logger.Level
		expected bool
	}{
		{logger.InfoLevel, logger.DebugLevel, false},
		{logger.InfoLevel, logger.InfoLevel, true},
		{logger.InfoLevel, logger.WarnLevel, true},
		{logger.InfoLevel, logger.ErrorLevel, true},
		{logger.ErrorLevel, logger.InfoLevel, false},
		{logger.ErrorLevel, logger.ErrorLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.base.String()+"_"+tt.target.String(), func(t *testing.T) {
			if got := tt.base.Enabled(tt.target); got != tt.expected {
				t.Errorf("Level.Enabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestFieldConstructors tests field constructor functions.
func TestFieldConstructors(t *testing.T) {
	// Test basic types
	boolField := logger.Bool("bool", true)
	if boolField.Type != logger.BoolType || boolField.Key != "bool" {
		t.Error("Bool() field incorrect")
	}

	intField := logger.Int("int", 42)
	if intField.Type != logger.IntType || intField.Key != "int" {
		t.Error("Int() field incorrect")
	}

	strField := logger.String("str", "test")
	if strField.Type != logger.StringType || strField.Key != "str" || strField.Str != "test" {
		t.Error("String() field incorrect")
	}

	durField := logger.Duration("dur", 5*time.Second)
	if durField.Type != logger.DurationType || durField.Key != "dur" {
		t.Error("Duration() field incorrect")
	}

	errField := logger.Err(errors.New("test error"))
	if errField.Type != logger.ErrorType || errField.Key != "error" {
		t.Error("Err() field incorrect")
	}

	// Test Skip field
	skipField := logger.Skip()
	if skipField.Type != logger.SkipType {
		t.Error("Skip() field incorrect")
	}
}

// TestContextExtraction tests context field extraction.
func TestContextExtraction(t *testing.T) {
	ctx := context.Background()

	// Empty context
	fields := logger.ExtractContextFields(ctx)
	if len(fields) != 0 {
		t.Errorf("Expected 0 fields from empty context, got %d", len(fields))
	}

	// Add trace ID
	ctx = logger.WithTraceID(ctx, "trace-123")
	fields = logger.ExtractContextFields(ctx)
	if len(fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(fields))
	}

	// Add more context
	ctx = logger.WithSpanID(ctx, "span-456")
	ctx = logger.WithRequestID(ctx, "req-789")
	ctx = logger.WithUserID(ctx, "user-abc")
	fields = logger.ExtractContextFields(ctx)
	if len(fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(fields))
	}

	// Test extraction functions
	if traceID, ok := logger.ExtractTraceID(ctx); !ok || traceID != "trace-123" {
		t.Error("Failed to extract trace ID")
	}
	if spanID, ok := logger.ExtractSpanID(ctx); !ok || spanID != "span-456" {
		t.Error("Failed to extract span ID")
	}
	if reqID, ok := logger.ExtractRequestID(ctx); !ok || reqID != "req-789" {
		t.Error("Failed to extract request ID")
	}
	if userID, ok := logger.ExtractUserID(ctx); !ok || userID != "user-abc" {
		t.Error("Failed to extract user ID")
	}
}

// TestConfig tests configuration creation.
func TestConfig(t *testing.T) {
	// Production config
	prodCfg := logger.NewConfig()
	if prodCfg.Level != logger.InfoLevel {
		t.Error("Production config should default to InfoLevel")
	}
	if prodCfg.Development {
		t.Error("Production config should not be in development mode")
	}
	if prodCfg.Encoding != "json" {
		t.Error("Production config should use json encoding")
	}

	// Development config
	devCfg := logger.NewDevelopmentConfig()
	if devCfg.Level != logger.DebugLevel {
		t.Error("Development config should default to DebugLevel")
	}
	if !devCfg.Development {
		t.Error("Development config should be in development mode")
	}
	if devCfg.Encoding != "console" {
		t.Error("Development config should use console encoding")
	}
	if !devCfg.ColorEnabled {
		t.Error("Development config should have colors enabled")
	}
}

// TestZapLogger tests the Zap logger implementation (basic).
func TestZapLogger(t *testing.T) {
	// Test production logger creation
	log, err := zap.NewProductionLogger()
	if err != nil {
		t.Fatalf("Failed to create production logger: %v", err)
	}
	defer log.Sync()

	// Basic logging (should not panic)
	log.Info("test message", logger.String("key", "value"))

	// Test development logger creation
	devLog, err := zap.NewDevelopmentLogger()
	if err != nil {
		t.Fatalf("Failed to create development logger: %v", err)
	}
	defer devLog.Sync()

	devLog.Debug("debug message")

	// Test With
	withLog := log.With(logger.String("component", "test"))
	if withLog == nil {
		t.Fatal("With() returned nil")
	}
	withLog.Info("with field")

	// Test WithError
	errLog := log.WithError(errors.New("test error"))
	if errLog == nil {
		t.Fatal("WithError() returned nil")
	}
	errLog.Error("error occurred")

	// Test Named
	namedLog := log.Named("database")
	if namedLog == nil {
		t.Fatal("Named() returned nil")
	}
	namedLog.Info("named logger")

	// Test context logging
	ctx := logger.WithTraceID(context.Background(), "trace-xyz")
	log.InfoContext(ctx, "with trace")
}

// TestZapTestLogger tests the test logger.
func TestZapTestLogger(t *testing.T) {
	log := zap.NewTestLogger()
	if log == nil {
		t.Fatal("NewTestLogger() returned nil")
	}

	// Should not produce any output
	log.Info("test message")
	log.Error("error message")

	if err := log.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}
