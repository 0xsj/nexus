package log

import (
	"context"
)

// NoopLogger is a logger that discards all output.
// Useful for testing and benchmarks.
type NoopLogger struct {
	level  Level
	fields Fields
}

// NewNoop creates a new no-op logger.
func NewNoop() *NoopLogger {
	return &NoopLogger{
		level: LevelInfo,
	}
}

// NewNoopWithLevel creates a new no-op logger with a specific level.
func NewNoopWithLevel(level Level) *NoopLogger {
	return &NoopLogger{
		level: level,
	}
}

// Debug does nothing.
func (l *NoopLogger) Debug(msg string, fields ...Field) {}

// Info does nothing.
func (l *NoopLogger) Info(msg string, fields ...Field) {}

// Warn does nothing.
func (l *NoopLogger) Warn(msg string, fields ...Field) {}

// Error does nothing.
func (l *NoopLogger) Error(msg string, fields ...Field) {}

// Fatal does nothing (does NOT exit).
func (l *NoopLogger) Fatal(msg string, fields ...Field) {}

// Log does nothing.
func (l *NoopLogger) Log(level Level, msg string, fields ...Field) {}

// With returns the same no-op logger.
func (l *NoopLogger) With(fields ...Field) Logger {
	return &NoopLogger{
		level:  l.level,
		fields: l.fields.Merge(fields),
	}
}

// WithContext returns the same no-op logger.
func (l *NoopLogger) WithContext(ctx context.Context) Logger {
	return l
}

// Level returns the current log level.
func (l *NoopLogger) Level() Level {
	return l.level
}

// SetLevel sets the log level.
func (l *NoopLogger) SetLevel(level Level) {
	l.level = level
}

// Enabled returns true if the given level is enabled.
func (l *NoopLogger) Enabled(level Level) bool {
	return level.Enabled(l.level)
}

// ============================================================================
// Recording Logger (for testing)
// ============================================================================

// LogEntry represents a recorded log entry.
type LogEntry struct {
	Level   Level
	Message string
	Fields  Fields
}

// RecordingLogger records all log entries for inspection in tests.
type RecordingLogger struct {
	level   Level
	fields  Fields
	entries []LogEntry
}

// NewRecordingLogger creates a new recording logger.
func NewRecordingLogger() *RecordingLogger {
	return &RecordingLogger{
		level:   LevelDebug, // Capture everything by default
		entries: make([]LogEntry, 0),
	}
}

// Debug records a debug message.
func (l *RecordingLogger) Debug(msg string, fields ...Field) {
	l.Log(LevelDebug, msg, fields...)
}

// Info records an info message.
func (l *RecordingLogger) Info(msg string, fields ...Field) {
	l.Log(LevelInfo, msg, fields...)
}

// Warn records a warning message.
func (l *RecordingLogger) Warn(msg string, fields ...Field) {
	l.Log(LevelWarn, msg, fields...)
}

// Error records an error message.
func (l *RecordingLogger) Error(msg string, fields ...Field) {
	l.Log(LevelError, msg, fields...)
}

// Fatal records a fatal message (does NOT exit).
func (l *RecordingLogger) Fatal(msg string, fields ...Field) {
	l.Log(LevelFatal, msg, fields...)
}

// Log records at the specified level.
func (l *RecordingLogger) Log(level Level, msg string, fields ...Field) {
	if !l.Enabled(level) {
		return
	}

	allFields := l.fields.Merge(fields)

	l.entries = append(l.entries, LogEntry{
		Level:   level,
		Message: msg,
		Fields:  allFields.Clone(),
	})
}

// With returns a logger with the given fields attached.
func (l *RecordingLogger) With(fields ...Field) Logger {
	return &RecordingLogger{
		level:   l.level,
		fields:  l.fields.Merge(fields),
		entries: l.entries, // Share entries slice
	}
}

// WithContext returns a logger that extracts fields from context.
func (l *RecordingLogger) WithContext(ctx context.Context) Logger {
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
func (l *RecordingLogger) Level() Level {
	return l.level
}

// SetLevel sets the log level.
func (l *RecordingLogger) SetLevel(level Level) {
	l.level = level
}

// Enabled returns true if the given level is enabled.
func (l *RecordingLogger) Enabled(level Level) bool {
	return level.Enabled(l.level)
}

// ============================================================================
// Recording Logger - Inspection Methods
// ============================================================================

// Entries returns all recorded entries.
func (l *RecordingLogger) Entries() []LogEntry {
	return l.entries
}

// Len returns the number of recorded entries.
func (l *RecordingLogger) Len() int {
	return len(l.entries)
}

// Clear removes all recorded entries.
func (l *RecordingLogger) Clear() {
	l.entries = l.entries[:0]
}

// Last returns the last recorded entry, or nil if empty.
func (l *RecordingLogger) Last() *LogEntry {
	if len(l.entries) == 0 {
		return nil
	}
	return &l.entries[len(l.entries)-1]
}

// First returns the first recorded entry, or nil if empty.
func (l *RecordingLogger) First() *LogEntry {
	if len(l.entries) == 0 {
		return nil
	}
	return &l.entries[0]
}

// EntriesAt returns all entries at the given level.
func (l *RecordingLogger) EntriesAt(level Level) []LogEntry {
	var result []LogEntry
	for _, entry := range l.entries {
		if entry.Level == level {
			result = append(result, entry)
		}
	}
	return result
}

// HasEntry returns true if an entry with the given message exists.
func (l *RecordingLogger) HasEntry(msg string) bool {
	for _, entry := range l.entries {
		if entry.Message == msg {
			return true
		}
	}
	return false
}

// HasEntryAt returns true if an entry with the given level and message exists.
func (l *RecordingLogger) HasEntryAt(level Level, msg string) bool {
	for _, entry := range l.entries {
		if entry.Level == level && entry.Message == msg {
			return true
		}
	}
	return false
}

// HasField returns true if any entry has a field with the given key.
func (l *RecordingLogger) HasField(key string) bool {
	for _, entry := range l.entries {
		if entry.Fields.Has(key) {
			return true
		}
	}
	return false
}

// FindByMessage returns the first entry with the given message.
func (l *RecordingLogger) FindByMessage(msg string) *LogEntry {
	for i, entry := range l.entries {
		if entry.Message == msg {
			return &l.entries[i]
		}
	}
	return nil
}

// FindByField returns the first entry with the given field key and value.
func (l *RecordingLogger) FindByField(key string, value any) *LogEntry {
	for i, entry := range l.entries {
		if field := entry.Fields.Get(key); field != nil && field.Value == value {
			return &l.entries[i]
		}
	}
	return nil
}

// DebugEntries returns all debug entries.
func (l *RecordingLogger) DebugEntries() []LogEntry {
	return l.EntriesAt(LevelDebug)
}

// InfoEntries returns all info entries.
func (l *RecordingLogger) InfoEntries() []LogEntry {
	return l.EntriesAt(LevelInfo)
}

// WarnEntries returns all warning entries.
func (l *RecordingLogger) WarnEntries() []LogEntry {
	return l.EntriesAt(LevelWarn)
}

// ErrorEntries returns all error entries.
func (l *RecordingLogger) ErrorEntries() []LogEntry {
	return l.EntriesAt(LevelError)
}

// FatalEntries returns all fatal entries.
func (l *RecordingLogger) FatalEntries() []LogEntry {
	return l.EntriesAt(LevelFatal)
}
