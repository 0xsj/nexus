// pkg/graphql/scalars/datetime.go

package scalars

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// ============================================================================
// DateTime Scalar
// ============================================================================

// DateTime is a custom GraphQL scalar for time.Time.
// Serializes to ISO 8601 format (RFC3339).
type DateTime struct{}

// MarshalDateTime marshals a time.Time to a GraphQL scalar.
func MarshalDateTime(t time.Time) graphql.Marshaler {
	if t.IsZero() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(t.UTC().Format(time.RFC3339)))
	})
}

// UnmarshalDateTime unmarshals a GraphQL scalar to time.Time.
func UnmarshalDateTime(v interface{}) (time.Time, error) {
	switch v := v.(type) {
	case string:
		return parseDateTime(v)
	case []byte:
		return parseDateTime(string(v))
	case int64:
		return time.Unix(v, 0).UTC(), nil
	case int:
		return time.Unix(int64(v), 0).UTC(), nil
	case float64:
		return time.Unix(int64(v), 0).UTC(), nil
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("datetime must be a string or unix timestamp, got %T", v)
	}
}

// parseDateTime attempts to parse a string as a datetime using multiple formats.
func parseDateTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Try common formats in order of preference
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid datetime format: %s", s)
}

// ============================================================================
// Date Scalar (date only, no time)
// ============================================================================

// Date is a custom GraphQL scalar for date-only values.
// Serializes to YYYY-MM-DD format.
type Date struct{}

const dateFormat = "2006-01-02"

// MarshalDate marshals a time.Time to a date string.
func MarshalDate(t time.Time) graphql.Marshaler {
	if t.IsZero() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(t.Format(dateFormat)))
	})
}

// UnmarshalDate unmarshals a GraphQL scalar to time.Time (date only).
func UnmarshalDate(v interface{}) (time.Time, error) {
	switch v := v.(type) {
	case string:
		if v == "" {
			return time.Time{}, nil
		}
		return time.Parse(dateFormat, v)
	case []byte:
		if len(v) == 0 {
			return time.Time{}, nil
		}
		return time.Parse(dateFormat, string(v))
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("date must be a string in YYYY-MM-DD format, got %T", v)
	}
}

// ============================================================================
// Time Scalar (time only, no date)
// ============================================================================

// Time is a custom GraphQL scalar for time-only values.
// Serializes to HH:MM:SS format.
type Time struct{}

const timeFormat = "15:04:05"

// MarshalTime marshals a time.Time to a time string.
func MarshalTime(t time.Time) graphql.Marshaler {
	if t.IsZero() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(t.Format(timeFormat)))
	})
}

// UnmarshalTime unmarshals a GraphQL scalar to time.Time (time only).
func UnmarshalTime(v interface{}) (time.Time, error) {
	switch v := v.(type) {
	case string:
		if v == "" {
			return time.Time{}, nil
		}
		return time.Parse(timeFormat, v)
	case []byte:
		if len(v) == 0 {
			return time.Time{}, nil
		}
		return time.Parse(timeFormat, string(v))
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("time must be a string in HH:MM:SS format, got %T", v)
	}
}

// ============================================================================
// Duration Scalar
// ============================================================================

// MarshalDuration marshals a time.Duration to a GraphQL scalar.
// Serializes to a string like "1h30m" or milliseconds as int.
func MarshalDuration(d time.Duration) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(d.String()))
	})
}

// UnmarshalDuration unmarshals a GraphQL scalar to time.Duration.
func UnmarshalDuration(v interface{}) (time.Duration, error) {
	switch v := v.(type) {
	case string:
		return time.ParseDuration(v)
	case int64:
		return time.Duration(v) * time.Millisecond, nil
	case int:
		return time.Duration(v) * time.Millisecond, nil
	case float64:
		return time.Duration(v) * time.Millisecond, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("duration must be a string or milliseconds, got %T", v)
	}
}

// ============================================================================
// Timestamp Scalar (Unix timestamp)
// ============================================================================

// MarshalTimestamp marshals a time.Time to a Unix timestamp (seconds).
func MarshalTimestamp(t time.Time) graphql.Marshaler {
	if t.IsZero() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatInt(t.Unix(), 10))
	})
}

// UnmarshalTimestamp unmarshals a Unix timestamp to time.Time.
func UnmarshalTimestamp(v interface{}) (time.Time, error) {
	switch v := v.(type) {
	case int64:
		return time.Unix(v, 0).UTC(), nil
	case int:
		return time.Unix(int64(v), 0).UTC(), nil
	case float64:
		return time.Unix(int64(v), 0).UTC(), nil
	case string:
		ts, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid timestamp: %s", v)
		}
		return time.Unix(ts, 0).UTC(), nil
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("timestamp must be a unix timestamp, got %T", v)
	}
}
