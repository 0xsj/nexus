package log

import (
	"strings"
)

// Level represents a logging level.
type Level int

const (
	// LevelDebug is for debug messages.
	LevelDebug Level = iota - 1

	// LevelInfo is for informational messages.
	LevelInfo

	// LevelWarn is for warning messages.
	LevelWarn

	// LevelError is for error messages.
	LevelError

	// LevelFatal is for fatal messages (application will exit).
	LevelFatal
)

// String returns the string representation of the level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ShortString returns a short (3-4 char) representation.
func (l Level) ShortString() string {
	switch l {
	case LevelDebug:
		return "DBG"
	case LevelInfo:
		return "INF"
	case LevelWarn:
		return "WRN"
	case LevelError:
		return "ERR"
	case LevelFatal:
		return "FTL"
	default:
		return "???"
	}
}

// ParseLevel parses a string into a Level.
func ParseLevel(s string) (Level, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG", "DBG", "D", "-1":
		return LevelDebug, nil
	case "INFO", "INF", "I", "0":
		return LevelInfo, nil
	case "WARN", "WARNING", "WRN", "W", "1":
		return LevelWarn, nil
	case "ERROR", "ERR", "E", "2":
		return LevelError, nil
	case "FATAL", "FTL", "F", "3":
		return LevelFatal, nil
	default:
		return LevelInfo, ErrLogParseError("ParseLevel", s)
	}
}

// MustParseLevel parses a level or panics.
func MustParseLevel(s string) Level {
	level, err := ParseLevel(s)
	if err != nil {
		panic(err)
	}
	return level
}

// Enabled returns true if this level is enabled for the given minimum level.
func (l Level) Enabled(minLevel Level) bool {
	return l >= minLevel
}

// ============================================================================
// ANSI Colors
// ============================================================================

// Color represents an ANSI color code.
type Color string

const (
	// Reset resets all attributes.
	ColorReset Color = "\033[0m"

	// Regular colors
	ColorBlack   Color = "\033[30m"
	ColorRed     Color = "\033[31m"
	ColorGreen   Color = "\033[32m"
	ColorYellow  Color = "\033[33m"
	ColorBlue    Color = "\033[34m"
	ColorMagenta Color = "\033[35m"
	ColorCyan    Color = "\033[36m"
	ColorWhite   Color = "\033[37m"

	// Bright colors
	ColorBrightBlack   Color = "\033[90m"
	ColorBrightRed     Color = "\033[91m"
	ColorBrightGreen   Color = "\033[92m"
	ColorBrightYellow  Color = "\033[93m"
	ColorBrightBlue    Color = "\033[94m"
	ColorBrightMagenta Color = "\033[95m"
	ColorBrightCyan    Color = "\033[96m"
	ColorBrightWhite   Color = "\033[97m"

	// Styles
	ColorBold      Color = "\033[1m"
	ColorDim       Color = "\033[2m"
	ColorItalic    Color = "\033[3m"
	ColorUnderline Color = "\033[4m"
)

// String returns the color code as a string.
func (c Color) String() string {
	return string(c)
}

// Wrap wraps text with the color and reset.
func (c Color) Wrap(text string) string {
	return string(c) + text + string(ColorReset)
}

// LevelColor returns the color for a log level.
func LevelColor(level Level) Color {
	switch level {
	case LevelDebug:
		return ColorBrightBlack
	case LevelInfo:
		return ColorBrightCyan
	case LevelWarn:
		return ColorBrightYellow
	case LevelError:
		return ColorBrightRed
	case LevelFatal:
		return ColorRed
	default:
		return ColorWhite
	}
}

// LevelColorBold returns the bold color for a log level.
func LevelColorBold(level Level) string {
	return string(ColorBold) + string(LevelColor(level))
}

// ============================================================================
// Colorize Helpers
// ============================================================================

// Colorize wraps text with a color.
func Colorize(color Color, text string) string {
	return color.Wrap(text)
}

// ColorizeLevel returns the level string with appropriate color.
func ColorizeLevel(level Level) string {
	return LevelColor(level).Wrap(level.ShortString())
}

// ColorizeLevelPadded returns a padded, colored level string.
func ColorizeLevelPadded(level Level) string {
	return LevelColor(level).Wrap(padRight(level.ShortString(), 5))
}

// ColorizeKey colors a field key.
func ColorizeKey(key string) string {
	return ColorCyan.Wrap(key)
}

// ColorizeString colors a string value.
func ColorizeString(value string) string {
	return ColorGreen.Wrap(`"` + value + `"`)
}

// ColorizeNumber colors a numeric value.
func ColorizeNumber(value string) string {
	return ColorYellow.Wrap(value)
}

// ColorizeBool colors a boolean value.
func ColorizeBool(value string) string {
	return ColorMagenta.Wrap(value)
}

// ColorizeError colors an error value.
func ColorizeError(value string) string {
	return ColorBrightRed.Wrap(value)
}

// ColorizeTimestamp colors a timestamp.
func ColorizeTimestamp(value string) string {
	return ColorBrightBlack.Wrap(value)
}

// ColorizeCaller colors caller information.
func ColorizeCaller(value string) string {
	return ColorBrightBlack.Wrap(value)
}

// ColorizeMessage colors the log message based on level.
func ColorizeMessage(level Level, msg string) string {
	switch level {
	case LevelError, LevelFatal:
		return ColorBrightRed.Wrap(msg)
	case LevelWarn:
		return ColorBrightYellow.Wrap(msg)
	default:
		return msg
	}
}

// ============================================================================
// Utility Functions
// ============================================================================

// padRight pads a string to the right with spaces.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// StripColors removes ANSI color codes from a string.
func StripColors(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			// Find the end of the escape sequence
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				i = j + 1
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}

	return result.String()
}
