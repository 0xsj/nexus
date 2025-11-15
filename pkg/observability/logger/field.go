package logger

import (
	"fmt"
	"runtime"
	"time"
)

// FieldType represents the type of a field value.
type FieldType uint8

const (
	UnknownType FieldType = iota
	BoolType
	IntType
	Int8Type
	Int16Type
	Int32Type
	Int64Type
	UintType
	Uint8Type
	Uint16Type
	Uint32Type
	Uint64Type
	Float32Type
	Float64Type
	StringType
	TimeType
	DurationType
	ErrorType
	StringsType
	IntsType
	AnyType
	SkipType
	// Stack trace types
	StackTraceType
)

// Field represents a key-value pair for structured logging.
type Field struct {
	Key       string
	Type      FieldType
	Integer   int64
	Float     float64
	Str       string // Changed from String to Str
	Interface interface{}
}

// String returns a string representation of the field.
func (f Field) String() string {
	switch f.Type {
	case BoolType:
		return fmt.Sprintf("%s=%t", f.Key, f.Integer == 1)
	case IntType, Int8Type, Int16Type, Int32Type, Int64Type:
		return fmt.Sprintf("%s=%d", f.Key, f.Integer)
	case UintType, Uint8Type, Uint16Type, Uint32Type, Uint64Type:
		return fmt.Sprintf("%s=%d", f.Key, uint64(f.Integer))
	case Float32Type, Float64Type:
		return fmt.Sprintf("%s=%f", f.Key, f.Float)
	case StringType:
		return fmt.Sprintf("%s=%s", f.Key, f.Str) // Changed to f.Str
	case ErrorType:
		if f.Interface != nil {
			return fmt.Sprintf("%s=%v", f.Key, f.Interface)
		}
		return fmt.Sprintf("%s=<nil>", f.Key)
	case TimeType:
		if t, ok := f.Interface.(time.Time); ok {
			return fmt.Sprintf("%s=%s", f.Key, t.Format(time.RFC3339))
		}
		return fmt.Sprintf("%s=%v", f.Key, f.Interface)
	case DurationType:
		return fmt.Sprintf("%s=%s", f.Key, time.Duration(f.Integer).String())
	case StackTraceType:
		return fmt.Sprintf("%s=%s", f.Key, f.Str) // Changed to f.Str
	default:
		return fmt.Sprintf("%s=%v", f.Key, f.Interface)
	}
}

// Bool constructs a field that carries a bool.
func Bool(key string, val bool) Field {
	var i int64
	if val {
		i = 1
	}
	return Field{Key: key, Type: BoolType, Integer: i}
}

// Int constructs a field that carries an int.
func Int(key string, val int) Field {
	return Field{Key: key, Type: IntType, Integer: int64(val)}
}

// Int8 constructs a field that carries an int8.
func Int8(key string, val int8) Field {
	return Field{Key: key, Type: Int8Type, Integer: int64(val)}
}

// Int16 constructs a field that carries an int16.
func Int16(key string, val int16) Field {
	return Field{Key: key, Type: Int16Type, Integer: int64(val)}
}

// Int32 constructs a field that carries an int32.
func Int32(key string, val int32) Field {
	return Field{Key: key, Type: Int32Type, Integer: int64(val)}
}

// Int64 constructs a field that carries an int64.
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: Int64Type, Integer: val}
}

// Uint constructs a field that carries a uint.
func Uint(key string, val uint) Field {
	return Field{Key: key, Type: UintType, Integer: int64(val)}
}

// Uint8 constructs a field that carries a uint8.
func Uint8(key string, val uint8) Field {
	return Field{Key: key, Type: Uint8Type, Integer: int64(val)}
}

// Uint16 constructs a field that carries a uint16.
func Uint16(key string, val uint16) Field {
	return Field{Key: key, Type: Uint16Type, Integer: int64(val)}
}

// Uint32 constructs a field that carries a uint32.
func Uint32(key string, val uint32) Field {
	return Field{Key: key, Type: Uint32Type, Integer: int64(val)}
}

// Uint64 constructs a field that carries a uint64.
func Uint64(key string, val uint64) Field {
	return Field{Key: key, Type: Uint64Type, Integer: int64(val)}
}

// Float32 constructs a field that carries a float32.
func Float32(key string, val float32) Field {
	return Field{Key: key, Type: Float32Type, Float: float64(val)}
}

// Float64 constructs a field that carries a float64.
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: Float64Type, Float: val}
}

// String constructs a field that carries a string.
func String(key string, val string) Field {
	return Field{Key: key, Type: StringType, Str: val} // Changed to Str
}

// Strings constructs a field that carries a slice of strings.
func Strings(key string, val []string) Field {
	return Field{Key: key, Type: StringsType, Interface: val}
}

// Ints constructs a field that carries a slice of ints.
func Ints(key string, val []int) Field {
	return Field{Key: key, Type: IntsType, Interface: val}
}

// Time constructs a field that carries a time.Time.
func Time(key string, val time.Time) Field {
	return Field{Key: key, Type: TimeType, Interface: val}
}

// Duration constructs a field that carries a time.Duration.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: DurationType, Integer: int64(val)}
}

// Err constructs a field that carries an error.
func Err(err error) Field {
	return Field{Key: "error", Type: ErrorType, Interface: err}
}

// NamedErr constructs a field that carries an error with a custom key.
func NamedErr(key string, err error) Field {
	return Field{Key: key, Type: ErrorType, Interface: err}
}

// Any constructs a field that carries an arbitrary value.
// Use this for types that don't have a dedicated constructor.
func Any(key string, val interface{}) Field {
	return Field{Key: key, Type: AnyType, Interface: val}
}

// Skip constructs a no-op field that is ignored.
// Useful for conditional field inclusion.
func Skip() Field {
	return Field{Type: SkipType}
}

// StackTrace captures the current stack trace.
func StackTrace(key string) Field {
	return Field{
		Key:  key,
		Type: StackTraceType,
		Str:  captureStackTrace(3, 10), // Changed to Str
	}
}

// StackTraceWithDepth captures a stack trace with custom skip and depth.
func StackTraceWithDepth(key string, skip int, depth int) Field {
	return Field{
		Key:  key,
		Type: StackTraceType,
		Str:  captureStackTrace(skip, depth), // Changed to Str
	}
}

// captureStackTrace captures the current stack trace.
func captureStackTrace(skip int, depth int) string {
	var frames []string

	pcs := make([]uintptr, depth)
	n := runtime.Callers(skip, pcs)

	if n == 0 {
		return "no stack trace available"
	}

	frames = make([]string, 0, n)
	callersFrames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := callersFrames.Next()
		frames = append(frames, fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line))

		if !more {
			break
		}
	}

	result := ""
	for i, frame := range frames {
		if i > 0 {
			result += "\n"
		}
		result += frame
	}

	return result
}

// StackField is a convenience function that combines error logging with stack trace.
func StackField(err error) []Field {
	if err == nil {
		return []Field{Skip()}
	}
	return []Field{
		Err(err),
		StackTrace("stacktrace"),
	}
}
