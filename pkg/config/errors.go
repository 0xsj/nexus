package config

import "fmt"

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return fmt.Sprintf("config validation error: %s: %s", e.Field, e.Message)
}

// ErrInvalidValue creates a validation error for invalid values.
func ErrInvalidValue(field, reason string) error {
	return ValidationError{
		Field:   field,
		Message: reason,
	}
}

// ErrRequired creates a validation error for required fields.
func ErrRequired(field string) error {
	return ValidationError{
		Field:   field,
		Message: "is required",
	}
}

// ErrOutOfRange creates a validation error for values out of range.
func ErrOutOfRange(field string, min, max interface{}) error {
	return ValidationError{
		Field:   field,
		Message: fmt.Sprintf("must be between %v and %v", min, max),
	}
}

// ErrTooShort creates a validation error for values that are too short.
func ErrTooShort(field string, minLength int) error {
	return ValidationError{
		Field:   field,
		Message: fmt.Sprintf("must be at least %d characters", minLength),
	}
}

// ErrTooLong creates a validation error for values that are too long.
func ErrTooLong(field string, maxLength int) error {
	return ValidationError{
		Field:   field,
		Message: fmt.Sprintf("must be at most %d characters", maxLength),
	}
}

// ErrInvalidFormat creates a validation error for invalid formats.
func ErrInvalidFormat(field, expectedFormat string) error {
	return ValidationError{
		Field:   field,
		Message: fmt.Sprintf("invalid format, expected %s", expectedFormat),
	}
}

// ErrConflict creates a validation error for conflicting configurations.
func ErrConflict(field1, field2, reason string) error {
	return ValidationError{
		Field:   fmt.Sprintf("%s and %s", field1, field2),
		Message: reason,
	}
}
