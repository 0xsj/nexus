package database

import "github.com/0xsj/nexus/platform/pkg/errors"

// Database-specific error codes.
const (
	CodeConnectionFailed    errors.Code = "DB_CONNECTION_FAILED"
	CodeQueryFailed         errors.Code = "DB_QUERY_FAILED"
	CodeTransactionFailed   errors.Code = "DB_TRANSACTION_FAILED"
	CodeNotFound            errors.Code = "DB_NOT_FOUND"
	CodeDuplicateKey        errors.Code = "DB_DUPLICATE_KEY"
	CodeConstraintViolation errors.Code = "DB_CONSTRAINT_VIOLATION"
	CodeTimeout             errors.Code = "DB_TIMEOUT"
	CodeCanceled            errors.Code = "DB_CANCELED"
)

// ErrConnectionFailed creates an error for connection failures.
func ErrConnectionFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeConnectionFailed)
}

// ErrQueryFailed creates an error for query execution failures.
func ErrQueryFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeQueryFailed)
}

// ErrTransactionFailed creates an error for transaction failures.
func ErrTransactionFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeTransactionFailed)
}

// ErrNotFound creates an error when a record is not found.
func ErrNotFound(operation string, entity string) *errors.Error {
	return errors.NotFound(operation, entity).
		WithCode(CodeNotFound)
}

// ErrDuplicateKey creates an error for duplicate key violations.
func ErrDuplicateKey(operation string, key string) *errors.Error {
	return errors.Conflict(operation, "duplicate key: "+key).
		WithCode(CodeDuplicateKey).
		WithMeta("key", key)
}

// ErrConstraintViolation creates an error for constraint violations.
func ErrConstraintViolation(operation string, constraint string) *errors.Error {
	return errors.Validation(operation, "constraint violation: "+constraint).
		WithCode(CodeConstraintViolation).
		WithMeta("constraint", constraint)
}

// ErrTimeout creates an error for query timeouts.
func ErrTimeout(operation string) *errors.Error {
	return errors.Infrastructure(operation, nil).
		WithCode(CodeTimeout).
		WithMessage("database operation timed out")
}

// ErrCanceled creates an error for canceled operations.
func ErrCanceled(operation string) *errors.Error {
	return errors.Infrastructure(operation, nil).
		WithCode(CodeCanceled).
		WithMessage("database operation was canceled")
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeNotFound
	}
	return false
}

// IsDuplicateKey returns true if the error is a duplicate key error.
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeDuplicateKey
	}
	return false
}

// IsConstraintViolation returns true if the error is a constraint violation.
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeConstraintViolation
	}
	return false
}

// IsTimeout returns true if the error is a timeout error.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeTimeout
	}
	return false
}

// IsCanceled returns true if the error is a canceled error.
func IsCanceled(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeCanceled
	}
	return false
}
