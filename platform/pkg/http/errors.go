package http

import (
	"errors"
	"fmt"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	// ErrRequestBodyTooLarge is returned when the request body exceeds the limit.
	ErrRequestBodyTooLarge = errors.New("request body too large")

	// ErrInvalidJSON is returned when JSON decoding fails.
	ErrInvalidJSON = errors.New("invalid JSON")

	// ErrMissingPathParam is returned when a required path parameter is missing.
	ErrMissingPathParam = errors.New("missing path parameter")

	// ErrMissingQueryParam is returned when a required query parameter is missing.
	ErrMissingQueryParam = errors.New("missing query parameter")

	// ErrInvalidQueryParam is returned when a query parameter has an invalid value.
	ErrInvalidQueryParam = errors.New("invalid query parameter")

	// ErrServerClosed is returned when the server is closed.
	ErrServerClosed = errors.New("server closed")

	// ErrServerTimeout is returned when the server times out.
	ErrServerTimeout = errors.New("server timeout")
)

// ============================================================================
// Error Types
// ============================================================================

// DecodeError represents a request decoding error.
type DecodeError struct {
	Field   string
	Message string
	Err     error
}

// Error implements the error interface.
func (e *DecodeError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("decode error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("decode error: %s", e.Message)
}

// Unwrap returns the underlying error.
func (e *DecodeError) Unwrap() error {
	return e.Err
}

// NewDecodeError creates a new DecodeError.
func NewDecodeError(field, message string, err error) *DecodeError {
	return &DecodeError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// ValidationError represents a request validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors struct {
	Errors []*ValidationError
}

// Error implements the error interface.
func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("validation failed: %d errors", len(e.Errors))
}

// Add adds a validation error.
func (e *ValidationErrors) Add(field, message string) {
	e.Errors = append(e.Errors, NewValidationError(field, message))
}

// HasErrors returns true if there are validation errors.
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// NewValidationErrors creates a new ValidationErrors.
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]*ValidationError, 0),
	}
}

// ============================================================================
// Error Helpers
// ============================================================================

// IsDecodeError checks if the error is a DecodeError.
func IsDecodeError(err error) bool {
	var decodeErr *DecodeError
	return errors.As(err, &decodeErr)
}

// IsValidationError checks if the error is a ValidationError.
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsValidationErrors checks if the error is a ValidationErrors.
func IsValidationErrors(err error) bool {
	var validationErrs *ValidationErrors
	return errors.As(err, &validationErrs)
}

// AsValidationErrors attempts to convert an error to ValidationErrors.
func AsValidationErrors(err error) (*ValidationErrors, bool) {
	var validationErrs *ValidationErrors
	if errors.As(err, &validationErrs) {
		return validationErrs, true
	}
	return nil, false
}
