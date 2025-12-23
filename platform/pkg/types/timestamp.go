package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp wraps time.Time with consistent JSON/SQL serialization.
// Uses UTC internally and serializes to ISO8601 format.
//
// Zero value represents an unset timestamp.
type Timestamp struct {
	value time.Time
}

// Now returns the current timestamp in UTC.
func Now() Timestamp {
	return Timestamp{value: time.Now().UTC()}
}

// FromTime creates a Timestamp from a time.Time.
// Converts to UTC internally.
func FromTime(t time.Time) Timestamp {
	return Timestamp{value: t.UTC()}
}

// FromUnix creates a Timestamp from Unix seconds.
func FromUnix(sec int64) Timestamp {
	return Timestamp{value: time.Unix(sec, 0).UTC()}
}

// FromUnixMilli creates a Timestamp from Unix milliseconds.
func FromUnixMilli(msec int64) Timestamp {
	return Timestamp{value: time.UnixMilli(msec).UTC()}
}

// ParseTimestamp parses an ISO8601/RFC3339 string into a Timestamp.
func ParseTimestamp(s string) (Timestamp, error) {
	if s == "" {
		return Timestamp{}, nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try RFC3339Nano as fallback
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return Timestamp{}, fmt.Errorf("invalid timestamp format: %w", err)
		}
	}

	return Timestamp{value: t.UTC()}, nil
}

// MustParseTimestamp parses a timestamp and panics if invalid.
// Only use for constants or tests.
func MustParseTimestamp(s string) Timestamp {
	ts, err := ParseTimestamp(s)
	if err != nil {
		panic(fmt.Sprintf("invalid timestamp: %v", err))
	}
	return ts
}

// Time returns the underlying time.Time value.
func (t Timestamp) Time() time.Time {
	return t.value
}

// IsZero returns true if the timestamp is unset.
func (t Timestamp) IsZero() bool {
	return t.value.IsZero()
}

// IsValid returns true if the timestamp is set (non-zero).
func (t Timestamp) IsValid() bool {
	return !t.value.IsZero()
}

// Unix returns the Unix timestamp in seconds.
func (t Timestamp) Unix() int64 {
	return t.value.Unix()
}

// UnixMilli returns the Unix timestamp in milliseconds.
func (t Timestamp) UnixMilli() int64 {
	return t.value.UnixMilli()
}

// String returns the ISO8601/RFC3339 string representation.
func (t Timestamp) String() string {
	if t.IsZero() {
		return ""
	}
	return t.value.Format(time.RFC3339)
}

// Format returns a custom formatted string.
func (t Timestamp) Format(layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.value.Format(layout)
}

// Before reports whether t is before other.
func (t Timestamp) Before(other Timestamp) bool {
	return t.value.Before(other.value)
}

// After reports whether t is after other.
func (t Timestamp) After(other Timestamp) bool {
	return t.value.After(other.value)
}

// Equal reports whether t and other represent the same instant.
func (t Timestamp) Equal(other Timestamp) bool {
	return t.value.Equal(other.value)
}

// Add returns the timestamp t + duration.
func (t Timestamp) Add(d time.Duration) Timestamp {
	return Timestamp{value: t.value.Add(d)}
}

// Sub returns the duration t - other.
func (t Timestamp) Sub(other Timestamp) time.Duration {
	return t.value.Sub(other.value)
}

// Truncate returns the timestamp truncated to the given duration.
func (t Timestamp) Truncate(d time.Duration) Timestamp {
	return Timestamp{value: t.value.Truncate(d)}
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
// Serializes to ISO8601/RFC3339 format.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.value.Format(time.RFC3339))
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Try unmarshaling as time.Time directly
		var tm time.Time
		if err := json.Unmarshal(data, &tm); err != nil {
			return err
		}
		*t = Timestamp{value: tm.UTC()}
		return nil
	}

	if s == "" || s == "null" {
		*t = Timestamp{}
		return nil
	}

	parsed, err := ParseTimestamp(s)
	if err != nil {
		return err
	}

	*t = parsed
	return nil
}

// ============================================================================
// SQL Scanning
// ============================================================================

// Scan implements sql.Scanner.
func (t *Timestamp) Scan(value interface{}) error {
	if value == nil {
		*t = Timestamp{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*t = Timestamp{value: v.UTC()}
		return nil

	case string:
		if v == "" {
			*t = Timestamp{}
			return nil
		}
		parsed, err := ParseTimestamp(v)
		if err != nil {
			return err
		}
		*t = parsed
		return nil

	case []byte:
		if len(v) == 0 {
			*t = Timestamp{}
			return nil
		}
		parsed, err := ParseTimestamp(string(v))
		if err != nil {
			return err
		}
		*t = parsed
		return nil

	case int64:
		*t = FromUnix(v)
		return nil

	default:
		return fmt.Errorf("cannot scan %T into Timestamp", value)
	}
}

// Value implements driver.Valuer.
func (t Timestamp) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.value, nil
}

// ============================================================================
// Text Marshaling
// ============================================================================

// MarshalText implements encoding.TextMarshaler.
func (t Timestamp) MarshalText() ([]byte, error) {
	if t.IsZero() {
		return []byte{}, nil
	}
	return []byte(t.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Timestamp) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*t = Timestamp{}
		return nil
	}

	parsed, err := ParseTimestamp(string(text))
	if err != nil {
		return err
	}

	*t = parsed
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// GoString implements fmt.GoStringer for debugging.
func (t Timestamp) GoString() string {
	if t.IsZero() {
		return "Timestamp{zero}"
	}
	return fmt.Sprintf("Timestamp{%s}", t.String())
}
