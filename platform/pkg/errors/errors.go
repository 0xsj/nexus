package errors

import (
	"errors"
	"fmt"
)

// Error is a structured error with rich context for production applications.
//
// It provides:
//   - Kind: Category for protocol mapping (HTTP status, gRPC code, etc.)
//   - Code: Machine-readable identifier for client error handling
//   - Operation: Context about where the error occurred
//   - Message: Human-readable description
//   - Metadata: Additional structured context (user IDs, resource IDs, etc.)
//   - Severity: Error severity level for logging and alerting
//   - Retryable: Whether the operation can be safely retried
//   - Err: Wrapped underlying error for error chain support
//
// Error implements the error interface and supports errors.Is/As/Unwrap.
type Error struct {
	// Kind categorizes the error for protocol mapping.
	Kind Kind

	// Code is a machine-readable error identifier.
	// Examples: CREDENTIAL_EXPIRED, DID_NOT_FOUND
	Code Code

	// Operation identifies where the error occurred.
	// Format: "Package.Type.Method" or "Service.Method"
	// Examples: "credential.Service.Issue", "did.Resolver.Resolve"
	Operation string

	// Message is a human-readable error description.
	Message string

	// Metadata contains additional structured context.
	// Use for: user IDs, resource IDs, field names, etc.
	Metadata map[string]any

	// Severity indicates the error severity level.
	Severity Severity

	// Retryable indicates if the operation can be safely retried.
	Retryable bool

	// Err is the wrapped underlying error.
	Err error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("%s: %s", e.Operation, e.Message)
	}
	return e.Message
}

// Unwrap returns the wrapped error for errors.Is/As support.
func (e *Error) Unwrap() error {
	return e.Err
}

// ============================================================================
// Constructors
// ============================================================================

// New creates a new Error with the specified kind, code, and message.
func New(kind Kind, code Code, message string) *Error {
	return &Error{
		Kind:      kind,
		Code:      code,
		Message:   message,
		Severity:  SeverityError,
		Retryable: kind.IsRetryableByDefault(),
		Err:       sentinelForKind(kind),
	}
}

// NotFound creates a not found error.
func NotFound(operation, resource string) *Error {
	return &Error{
		Kind:      KindNotFound,
		Code:      CodeNotFound,
		Operation: operation,
		Message:   fmt.Sprintf("%s not found", resource),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrNotFound,
	}
}

// NotFoundf creates a not found error with formatted message.
func NotFoundf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindNotFound,
		Code:      CodeNotFound,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrNotFound,
	}
}

// Validation creates a validation error.
func Validation(operation, message string) *Error {
	return &Error{
		Kind:      KindValidation,
		Code:      CodeInvalidInput,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrValidation,
	}
}

// Validationf creates a validation error with formatted message.
func Validationf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindValidation,
		Code:      CodeInvalidInput,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrValidation,
	}
}

// Conflict creates a conflict error.
func Conflict(operation, resource string) *Error {
	return &Error{
		Kind:      KindConflict,
		Code:      CodeAlreadyExists,
		Operation: operation,
		Message:   fmt.Sprintf("%s already exists", resource),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrConflict,
	}
}

// Conflictf creates a conflict error with formatted message.
func Conflictf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindConflict,
		Code:      CodeConflict,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrConflict,
	}
}

// Unauthorized creates an unauthorized error.
func Unauthorized(operation, message string) *Error {
	return &Error{
		Kind:      KindUnauthorized,
		Code:      CodeUnauthenticated,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrUnauthorized,
	}
}

// Unauthorizedf creates an unauthorized error with formatted message.
func Unauthorizedf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindUnauthorized,
		Code:      CodeUnauthenticated,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrUnauthorized,
	}
}

// Forbidden creates a forbidden error.
func Forbidden(operation, message string) *Error {
	return &Error{
		Kind:      KindForbidden,
		Code:      CodeUnauthorized,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrForbidden,
	}
}

// Forbiddenf creates a forbidden error with formatted message.
func Forbiddenf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindForbidden,
		Code:      CodeUnauthorized,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrForbidden,
	}
}

// Domain creates a domain/business rule error.
func Domain(operation, message string) *Error {
	return &Error{
		Kind:      KindDomain,
		Code:      CodeConflict,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrDomain,
	}
}

// Domainf creates a domain error with formatted message.
func Domainf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindDomain,
		Code:      CodeConflict,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: false,
		Err:       ErrDomain,
	}
}

// Infrastructure creates an infrastructure error.
func Infrastructure(operation string, err error) *Error {
	message := "infrastructure error"
	if err != nil {
		message = err.Error()
	}

	return &Error{
		Kind:      KindInfrastructure,
		Code:      CodeUnavailable,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: true,
		Err:       fmt.Errorf("%w: %v", ErrInfrastructure, err),
	}
}

// Timeout creates a timeout error.
func Timeout(operation, message string) *Error {
	return &Error{
		Kind:      KindTimeout,
		Code:      CodeTimeout,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: true,
		Err:       ErrTimeout,
	}
}

// Timeoutf creates a timeout error with formatted message.
func Timeoutf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindTimeout,
		Code:      CodeTimeout,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: true,
		Err:       ErrTimeout,
	}
}

// RateLimit creates a rate limit error.
func RateLimit(operation, message string) *Error {
	return &Error{
		Kind:      KindRateLimit,
		Code:      CodeRateLimitExceeded,
		Operation: operation,
		Message:   message,
		Severity:  SeverityError,
		Retryable: true,
		Err:       ErrRateLimit,
	}
}

// RateLimitf creates a rate limit error with formatted message.
func RateLimitf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindRateLimit,
		Code:      CodeRateLimitExceeded,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityError,
		Retryable: true,
		Err:       ErrRateLimit,
	}
}

// Internal creates an internal error.
func Internal(operation string, err error) *Error {
	message := "internal error"
	if err != nil {
		message = err.Error()
	}

	return &Error{
		Kind:      KindInternal,
		Code:      CodeInternal,
		Operation: operation,
		Message:   message,
		Severity:  SeverityCritical,
		Retryable: false,
		Err:       fmt.Errorf("%w: %v", ErrInternal, err),
	}
}

// Internalf creates an internal error with formatted message.
func Internalf(operation, format string, args ...any) *Error {
	return &Error{
		Kind:      KindInternal,
		Code:      CodeInternal,
		Operation: operation,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityCritical,
		Retryable: false,
		Err:       ErrInternal,
	}
}

// ============================================================================
// Wrapping
// ============================================================================

// Wrap wraps an error with additional operation context.
// If err is already an *Error, preserves Kind, Code, and Metadata.
// If err is a standard error, wraps it as KindInternal.
func Wrap(err error, operation string) error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if As(err, &appErr) {
		return &Error{
			Kind:      appErr.Kind,
			Code:      appErr.Code,
			Operation: operation,
			Message:   appErr.Message,
			Metadata:  appErr.Metadata,
			Severity:  appErr.Severity,
			Retryable: appErr.Retryable,
			Err:       appErr.Err,
		}
	}

	return &Error{
		Kind:      KindInternal,
		Code:      CodeInternal,
		Operation: operation,
		Message:   err.Error(),
		Severity:  SeverityError,
		Retryable: false,
		Err:       fmt.Errorf("%w: %v", ErrInternal, err),
	}
}

// Wrapf wraps an error with formatted message.
func Wrapf(err error, operation, format string, args ...any) error {
	if err == nil {
		return nil
	}

	wrapped := Wrap(err, operation).(*Error)
	wrapped.Message = fmt.Sprintf(format, args...)
	return wrapped
}

// ============================================================================
// Builder Methods
// ============================================================================

// WithCode sets a custom error code.
func (e *Error) WithCode(code Code) *Error {
	e.Code = code
	return e
}

// WithOperation sets the operation context.
func (e *Error) WithOperation(operation string) *Error {
	e.Operation = operation
	return e
}

// WithMessage sets a custom message.
func (e *Error) WithMessage(message string) *Error {
	e.Message = message
	return e
}

// WithMessagef sets a formatted message.
func (e *Error) WithMessagef(format string, args ...any) *Error {
	e.Message = fmt.Sprintf(format, args...)
	return e
}

// WithMeta adds a metadata key-value pair.
func (e *Error) WithMeta(key string, value any) *Error {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
	return e
}

// WithMetadata replaces all metadata.
func (e *Error) WithMetadata(metadata map[string]any) *Error {
	e.Metadata = metadata
	return e
}

// WithSeverity sets the error severity level.
func (e *Error) WithSeverity(severity Severity) *Error {
	e.Severity = severity
	return e
}

// WithRetryable sets whether the error is retryable.
func (e *Error) WithRetryable(retryable bool) *Error {
	e.Retryable = retryable
	return e
}

// WithKind sets a custom error kind.
func (e *Error) WithKind(kind Kind) *Error {
	e.Kind = kind
	e.Err = sentinelForKind(kind)
	return e
}

// WithEntity adds entity type and ID to metadata.
func (e *Error) WithEntity(entityType string, entityID any) *Error {
	return e.
		WithMeta("entity_type", entityType).
		WithMeta("entity_id", entityID)
}

// ============================================================================
// Type Assertions
// ============================================================================

// As checks if err is or wraps an *Error and assigns it to target.
func As(err error, target **Error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*Error); ok {
		*target = e
		return true
	}

	return errors.As(err, target)
}

// AsError extracts an *Error from err if it exists.
func AsError(err error) *Error {
	var e *Error
	if As(err, &e) {
		return e
	}
	return nil
}

// Is checks if an error matches a target error.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// ============================================================================
// Helpers
// ============================================================================

// GetKind extracts the Kind from an error.
// Returns KindInternal if err is not an *Error.
func GetKind(err error) Kind {
	if e := AsError(err); e != nil {
		return e.Kind
	}
	return KindInternal
}

// GetCode extracts the Code from an error.
// Returns CodeInternal if err is not an *Error.
func GetCode(err error) Code {
	if e := AsError(err); e != nil {
		return e.Code
	}
	return CodeInternal
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	if e := AsError(err); e != nil {
		return e.Retryable
	}
	return false
}
