// pkg/config/utils.go
package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseBool parses a string into a boolean value.
// Accepts: "true", "false", "1", "0", "yes", "no", "on", "off" (case-insensitive)
func ParseBool(value string) (bool, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	switch value {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

// ParseInt parses a string into an integer.
func ParseInt(value string) (int, error) {
	i, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %s", value)
	}
	return i, nil
}

// ParseInt64 parses a string into an int64.
func ParseInt64(value string) (int64, error) {
	i, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid int64 value: %s", value)
	}
	return i, nil
}

// ParseFloat parses a string into a float64.
func ParseFloat(value string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float value: %s", value)
	}
	return f, nil
}

// ParseDuration parses a string into a time.Duration.
// Accepts formats like "5s", "10m", "1h", "500ms"
func ParseDuration(value string) (time.Duration, error) {
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", value)
	}
	return d, nil
}

// ParseStringSlice parses a comma-separated string into a slice of strings.
// Empty strings are filtered out.
func ParseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// ParseIntSlice parses a comma-separated string into a slice of integers.
func ParseIntSlice(value string) ([]int, error) {
	if value == "" {
		return []int{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		i, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid integer in slice: %s", trimmed)
		}
		result = append(result, i)
	}

	return result, nil
}

// MustParseBool is like ParseBool but panics on error.
// Use only when the value is guaranteed to be valid.
func MustParseBool(value string) bool {
	b, err := ParseBool(value)
	if err != nil {
		panic(err)
	}
	return b
}

// MustParseInt is like ParseInt but panics on error.
func MustParseInt(value string) int {
	i, err := ParseInt(value)
	if err != nil {
		panic(err)
	}
	return i
}

// MustParseDuration is like ParseDuration but panics on error.
func MustParseDuration(value string) time.Duration {
	d, err := ParseDuration(value)
	if err != nil {
		panic(err)
	}
	return d
}

// ValidateRange checks if a value is within a specified range.
func ValidateRange(field string, value, min, max int) error {
	if value < min || value > max {
		return ErrOutOfRange(field, min, max)
	}
	return nil
}

// ValidateMinMax checks if min is less than or equal to max.
func ValidateMinMax(minField, maxField string, min, max int) error {
	if min > max {
		return ErrConflict(minField, maxField, "min cannot be greater than max")
	}
	return nil
}

// ValidateStringLength checks if a string length is within bounds.
func ValidateStringLength(field, value string, min, max int) error {
	length := len(value)
	if length < min {
		return ErrTooShort(field, min)
	}
	if max > 0 && length > max {
		return ErrTooLong(field, max)
	}
	return nil
}

// ValidateNotEmpty checks if a string is not empty.
func ValidateNotEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrRequired(field)
	}
	return nil
}

// ValidateOneOf checks if a value is one of the allowed values.
func ValidateOneOf(field, value string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return ErrInvalidValue(field, "must be one of: "+strings.Join(allowed, ", "))
}

// ValidatePositive checks if a value is positive (> 0).
func ValidatePositive(field string, value int) error {
	if value <= 0 {
		return ErrInvalidValue(field, "must be positive")
	}
	return nil
}

// ValidateNonNegative checks if a value is non-negative (>= 0).
func ValidateNonNegative(field string, value int) error {
	if value < 0 {
		return ErrInvalidValue(field, "must be non-negative")
	}
	return nil
}
