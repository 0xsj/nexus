package log

import (
	"fmt"
	"time"
)

// FieldType represents the type of a log field.
type FieldType uint8

const (
	FieldTypeString FieldType = iota
	FieldTypeInt
	FieldTypeInt64
	FieldTypeUint
	FieldTypeUint64
	FieldTypeFloat64
	FieldTypeBool
	FieldTypeTime
	FieldTypeDuration
	FieldTypeError
	FieldTypeAny
	FieldTypeGroup
)

// Field represents a structured log field.
type Field struct {
	Key   string
	Type  FieldType
	Value any
}

// ============================================================================
// Field Constructors
// ============================================================================

// String creates a string field.
func String(key string, value string) Field {
	return Field{Key: key, Type: FieldTypeString, Value: value}
}

// Strings creates a string slice field.
func Strings(key string, values []string) Field {
	return Field{Key: key, Type: FieldTypeAny, Value: values}
}

// Int creates an int field.
func Int(key string, value int) Field {
	return Field{Key: key, Type: FieldTypeInt, Value: value}
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return Field{Key: key, Type: FieldTypeInt64, Value: value}
}

// Uint creates a uint field.
func Uint(key string, value uint) Field {
	return Field{Key: key, Type: FieldTypeUint, Value: value}
}

// Uint64 creates a uint64 field.
func Uint64(key string, value uint64) Field {
	return Field{Key: key, Type: FieldTypeUint64, Value: value}
}

// Float64 creates a float64 field.
func Float64(key string, value float64) Field {
	return Field{Key: key, Type: FieldTypeFloat64, Value: value}
}

// Bool creates a bool field.
func Bool(key string, value bool) Field {
	return Field{Key: key, Type: FieldTypeBool, Value: value}
}

// Time creates a time field.
func Time(key string, value time.Time) Field {
	return Field{Key: key, Type: FieldTypeTime, Value: value}
}

// Duration creates a duration field.
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Type: FieldTypeDuration, Value: value}
}

// Err creates an error field with key "error".
func Err(err error) Field {
	return Field{Key: "error", Type: FieldTypeError, Value: err}
}

// NamedErr creates an error field with a custom key.
func NamedErr(key string, err error) Field {
	return Field{Key: key, Type: FieldTypeError, Value: err}
}

// Any creates a field with any value.
func Any(key string, value any) Field {
	return Field{Key: key, Type: FieldTypeAny, Value: value}
}

// Group creates a group of fields under a key.
func Group(key string, fields ...Field) Field {
	return Field{Key: key, Type: FieldTypeGroup, Value: fields}
}

// ============================================================================
// Common Fields
// ============================================================================

// RequestID creates a request ID field.
func RequestID(id string) Field {
	return String("request_id", id)
}

// TraceID creates a trace ID field.
func TraceID(id string) Field {
	return String("trace_id", id)
}

// SpanID creates a span ID field.
func SpanID(id string) Field {
	return String("span_id", id)
}

// UserID creates a user ID field.
func UserID(id string) Field {
	return String("user_id", id)
}

// Component creates a component field.
func Component(name string) Field {
	return String("component", name)
}

// Operation creates an operation field.
func Operation(name string) Field {
	return String("operation", name)
}

// Method creates an HTTP method field.
func Method(method string) Field {
	return String("method", method)
}

// Path creates an HTTP path field.
func Path(path string) Field {
	return String("path", path)
}

// Status creates an HTTP status code field.
func Status(code int) Field {
	return Int("status", code)
}

// Latency creates a latency duration field.
func Latency(d time.Duration) Field {
	return Duration("latency", d)
}

// BytesIn creates a bytes received field.
func BytesIn(n int64) Field {
	return Int64("bytes_in", n)
}

// BytesOut creates a bytes sent field.
func BytesOut(n int64) Field {
	return Int64("bytes_out", n)
}

// ============================================================================
// Field Formatting
// ============================================================================

// StringValue returns the field value as a string.
func (f Field) StringValue() string {
	switch f.Type {
	case FieldTypeString:
		return f.Value.(string)
	case FieldTypeInt:
		return fmt.Sprintf("%d", f.Value.(int))
	case FieldTypeInt64:
		return fmt.Sprintf("%d", f.Value.(int64))
	case FieldTypeUint:
		return fmt.Sprintf("%d", f.Value.(uint))
	case FieldTypeUint64:
		return fmt.Sprintf("%d", f.Value.(uint64))
	case FieldTypeFloat64:
		return fmt.Sprintf("%.3f", f.Value.(float64))
	case FieldTypeBool:
		return fmt.Sprintf("%t", f.Value.(bool))
	case FieldTypeTime:
		return f.Value.(time.Time).Format(time.RFC3339)
	case FieldTypeDuration:
		return f.Value.(time.Duration).String()
	case FieldTypeError:
		if err, ok := f.Value.(error); ok && err != nil {
			return err.Error()
		}
		return "<nil>"
	case FieldTypeAny:
		return fmt.Sprintf("%v", f.Value)
	case FieldTypeGroup:
		return fmt.Sprintf("%v", f.Value)
	default:
		return fmt.Sprintf("%v", f.Value)
	}
}

// ColoredValue returns the field value with appropriate coloring.
func (f Field) ColoredValue() string {
	switch f.Type {
	case FieldTypeString:
		return ColorizeString(f.Value.(string))
	case FieldTypeInt, FieldTypeInt64, FieldTypeUint, FieldTypeUint64, FieldTypeFloat64:
		return ColorizeNumber(f.StringValue())
	case FieldTypeBool:
		return ColorizeBool(f.StringValue())
	case FieldTypeDuration:
		return ColorizeNumber(f.StringValue())
	case FieldTypeTime:
		return ColorizeTimestamp(f.StringValue())
	case FieldTypeError:
		if err, ok := f.Value.(error); ok && err != nil {
			return ColorizeError(err.Error())
		}
		return ColorizeBool("<nil>")
	default:
		return f.StringValue()
	}
}

// IsEmpty returns true if the field has an empty/nil value.
func (f Field) IsEmpty() bool {
	switch f.Type {
	case FieldTypeString:
		return f.Value.(string) == ""
	case FieldTypeError:
		err, ok := f.Value.(error)
		return !ok || err == nil
	default:
		return f.Value == nil
	}
}

// ============================================================================
// Fields Collection
// ============================================================================

// Fields is a slice of Field.
type Fields []Field

// Add adds fields to the collection.
func (f Fields) Add(fields ...Field) Fields {
	return append(f, fields...)
}

// Get returns the field with the given key, or nil if not found.
func (f Fields) Get(key string) *Field {
	for i := range f {
		if f[i].Key == key {
			return &f[i]
		}
	}
	return nil
}

// Has returns true if a field with the given key exists.
func (f Fields) Has(key string) bool {
	return f.Get(key) != nil
}

// Keys returns all field keys.
func (f Fields) Keys() []string {
	keys := make([]string, len(f))
	for i, field := range f {
		keys[i] = field.Key
	}
	return keys
}

// ToMap converts fields to a map.
func (f Fields) ToMap() map[string]any {
	m := make(map[string]any, len(f))
	for _, field := range f {
		m[field.Key] = field.Value
	}
	return m
}

// Merge merges another Fields into this one.
// Later fields override earlier ones with the same key.
func (f Fields) Merge(other Fields) Fields {
	result := make(Fields, 0, len(f)+len(other))
	seen := make(map[string]int, len(f)+len(other))

	// Add original fields
	for _, field := range f {
		seen[field.Key] = len(result)
		result = append(result, field)
	}

	// Add/override with other fields
	for _, field := range other {
		if idx, exists := seen[field.Key]; exists {
			result[idx] = field
		} else {
			seen[field.Key] = len(result)
			result = append(result, field)
		}
	}

	return result
}

// Clone creates a copy of the fields.
func (f Fields) Clone() Fields {
	if f == nil {
		return nil
	}
	clone := make(Fields, len(f))
	copy(clone, f)
	return clone
}
