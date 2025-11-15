package logger

import (
	"fmt"
	"strings"
)

type Level int8

const (
	DebugLevel Level = iota - 1

	InfoLevel

	WarnLevel

	ErrorLevel

	FatalLevel

	PanicLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return fmt.Sprintf("LEVEL(%d)", l)
	}
}

// Color returns ANSI color codes for terminal output
func (l Level) Color() string {
	switch l {
	case DebugLevel:
		return "\033[36m" // Cyan
	case InfoLevel:
		return "\033[32m" // Green
	case WarnLevel:
		return "\033[33m" // Yellow
	case ErrorLevel:
		return "\033[31m" // Red
	case FatalLevel:
		return "\033[35m" // Magenta
	case PanicLevel:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Reset
	}
}

// returns ANSI reset code
func ResetColor() string {
	return "\033[0m"
}

// MarshalText implements encoding.TextMarshaler for JSON serialization.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON deserialization.
func (l *Level) UnmarshalText(text []byte) error {
	str := strings.ToUpper(string(text))
	switch str {
	case "DEBUG":
		*l = DebugLevel
	case "INFO":
		*l = InfoLevel
	case "WARN", "WARNING":
		*l = WarnLevel
	case "ERROR":
		*l = ErrorLevel
	case "FATAL":
		*l = FatalLevel
	case "PANIC":
		*l = PanicLevel
	default:
		return fmt.Errorf("unrecognized log level: %s", str)
	}
	return nil
}

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(target Level) bool {
	return target >= l
}

// ParseLevel converts a string to a Level.
// Returns InfoLevel and an error if the string is not recognized.
func ParseLevel(s string) (Level, error) {
	var level Level
	err := level.UnmarshalText([]byte(s))
	if err != nil {
		return InfoLevel, err
	}
	return level, nil
}
