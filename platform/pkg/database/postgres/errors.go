package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/0xsj/nexus/platform/pkg/database"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// PostgreSQL error codes.
// See: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	// Class 23 — Integrity Constraint Violation
	PgCodeUniqueViolation     = "23505"
	PgCodeForeignKeyViolation = "23503"
	PgCodeCheckViolation      = "23514"
	PgCodeNotNullViolation    = "23502"

	// Class 40 — Transaction Rollback
	PgCodeSerializationFailure = "40001"
	PgCodeDeadlockDetected     = "40P01"

	// Class 57 — Operator Intervention
	PgCodeQueryCanceled    = "57014"
	PgCodeAdminShutdown    = "57P01"
	PgCodeCrashShutdown    = "57P02"
	PgCodeCannotConnectNow = "57P03"

	// Class 08 — Connection Exception
	PgCodeConnectionException = "08000"
	PgCodeConnectionFailure   = "08006"

	// Class 53 — Insufficient Resources
	PgCodeDiskFull           = "53100"
	PgCodeOutOfMemory        = "53200"
	PgCodeTooManyConnections = "53300"
)

// Application-level error codes for PostgreSQL-specific errors.
const (
	CodeDeadlock      pkgerrors.Code = "PG_DEADLOCK"
	CodeSerialization pkgerrors.Code = "PG_SERIALIZATION_FAILURE"
)

// MapError converts a PostgreSQL error to a database error.
func MapError(operation string, err error) error {
	if err == nil {
		return nil
	}

	// Check for context errors first
	if errors.Is(err, context.Canceled) {
		return database.ErrCanceled(operation)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return database.ErrTimeout(operation)
	}

	// Check for pgx/pgconn errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return mapPgError(operation, pgErr)
	}

	// Check for connection errors by message
	if isConnectionError(err) {
		return database.ErrConnectionFailed(operation, err)
	}

	// Default to query failed
	return database.ErrQueryFailed(operation, err)
}

// mapPgError maps a pgconn.PgError to a database error.
func mapPgError(operation string, pgErr *pgconn.PgError) *pkgerrors.Error {
	switch pgErr.Code {
	// Unique violation
	case PgCodeUniqueViolation:
		return database.ErrDuplicateKey(operation, pgErr.ConstraintName).
			WithMeta("table", pgErr.TableName).
			WithMeta("detail", pgErr.Detail)

	// Foreign key violation
	case PgCodeForeignKeyViolation:
		return database.ErrConstraintViolation(operation, pgErr.ConstraintName).
			WithMeta("table", pgErr.TableName).
			WithMeta("detail", pgErr.Detail)

	// Check constraint violation
	case PgCodeCheckViolation:
		return database.ErrConstraintViolation(operation, pgErr.ConstraintName).
			WithMeta("table", pgErr.TableName).
			WithMeta("detail", pgErr.Detail)

	// Not null violation
	case PgCodeNotNullViolation:
		return database.ErrConstraintViolation(operation, pgErr.ColumnName+" cannot be null").
			WithMeta("table", pgErr.TableName).
			WithMeta("column", pgErr.ColumnName)

	// Serialization failure (retry possible)
	case PgCodeSerializationFailure:
		return pkgerrors.Infrastructure(operation, pgErr).
			WithCode(CodeSerialization).
			WithMeta("hint", pgErr.Hint)

	// Deadlock
	case PgCodeDeadlockDetected:
		return pkgerrors.Infrastructure(operation, pgErr).
			WithCode(CodeDeadlock).
			WithMeta("detail", pgErr.Detail)

	// Query canceled
	case PgCodeQueryCanceled:
		return database.ErrCanceled(operation)

	// Connection errors
	case PgCodeConnectionException, PgCodeConnectionFailure,
		PgCodeAdminShutdown, PgCodeCrashShutdown, PgCodeCannotConnectNow:
		return database.ErrConnectionFailed(operation, pgErr).
			WithMeta("pg_code", pgErr.Code)

	// Resource errors
	case PgCodeDiskFull, PgCodeOutOfMemory, PgCodeTooManyConnections:
		return database.ErrConnectionFailed(operation, pgErr).
			WithMeta("pg_code", pgErr.Code).
			WithMeta("detail", pgErr.Detail)

	default:
		return database.ErrQueryFailed(operation, pgErr).
			WithMeta("pg_code", pgErr.Code).
			WithMeta("detail", pgErr.Detail).
			WithMeta("hint", pgErr.Hint)
	}
}

// isConnectionError checks if the error is a connection-related error.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	connectionIndicators := []string{
		"connection refused",
		"connection reset",
		"connection closed",
		"no connection",
		"broken pipe",
		"network is unreachable",
		"host is unreachable",
		"timeout",
		"i/o timeout",
		"dial tcp",
		"eof",
	}

	for _, indicator := range connectionIndicators {
		if strings.Contains(msg, indicator) {
			return true
		}
	}

	return false
}

// IsRetryable returns true if the error is retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for our error types
	var pkgErr *pkgerrors.Error
	if errors.As(err, &pkgErr) {
		switch pkgErr.Code {
		case CodeSerialization, CodeDeadlock:
			return true
		}
	}

	// Check for pgx errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PgCodeSerializationFailure, PgCodeDeadlockDetected:
			return true
		}
	}

	// Connection errors may be retryable
	if isConnectionError(err) {
		return true
	}

	return false
}

// IsNotFound checks if the error indicates no rows found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return database.IsNotFound(err)
}

// IsDuplicateKey checks if the error is a duplicate key violation.
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}

	// Check our error type
	if database.IsDuplicateKey(err) {
		return true
	}

	// Check pgx error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == PgCodeUniqueViolation
	}

	return false
}

// IsConstraintViolation checks if the error is a constraint violation.
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	// Check our error type
	if database.IsConstraintViolation(err) {
		return true
	}

	// Check pgx error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PgCodeUniqueViolation, PgCodeForeignKeyViolation,
			PgCodeCheckViolation, PgCodeNotNullViolation:
			return true
		}
	}

	return false
}

// ConstraintName extracts the constraint name from a PostgreSQL error.
func ConstraintName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// TableName extracts the table name from a PostgreSQL error.
func TableName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.TableName
	}
	return ""
}

// ColumnName extracts the column name from a PostgreSQL error.
func ColumnName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ColumnName
	}
	return ""
}

// ErrorCode extracts the PostgreSQL error code.
func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// ErrorDetail extracts the error detail message.
func ErrorDetail(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Detail
	}
	return ""
}

// ErrorHint extracts the error hint message.
func ErrorHint(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Hint
	}
	return ""
}
