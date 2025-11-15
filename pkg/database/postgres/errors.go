// Package postgres provides PostgreSQL-specific database implementations.
package postgres

import (
	"errors"

	"github.com/0xsj/result"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes
// Reference: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	// Class 23 - Integrity Constraint Violation
	ErrCodeUniqueViolation     = "23505"
	ErrCodeForeignKeyViolation = "23503"
	ErrCodeCheckViolation      = "23514"
	ErrCodeNotNullViolation    = "23502"

	// Class 40 - Transaction Rollback
	ErrCodeDeadlockDetected     = "40P01"
	ErrCodeSerializationFailure = "40001"

	// Class 57 - Operator Intervention
	ErrCodeQueryCanceled = "57014"

	// Class 08 - Connection Exception
	ErrCodeConnectionException = "08000"
	ErrCodeConnectionFailure   = "08006"
)

// ToResultError converts a PostgreSQL error to a Result library error.
// This is the primary function for converting database errors to domain errors.
//
// Error mapping:
//   - pgx.ErrNoRows           → result.NotFound
//   - Unique violation        → result.Conflict
//   - Foreign key violation   → result.Domain
//   - Connection errors       → result.Infrastructure
//   - Deadlocks/timeouts      → result.Infrastructure
//   - Unknown errors          → result.Infrastructure
//
// Example:
//
//	_, err := pool.Exec(ctx, query, args...)
//	if err != nil {
//	    return postgres.ToResultError("UserRepository.Create", err)
//	}
func ToResultError(op string, err error) error {
	if err == nil {
		return nil
	}

	// Handle pgx.ErrNoRows (common case)
	if errors.Is(err, pgx.ErrNoRows) {
		return result.NotFound(op, "resource")
	}

	// Handle PostgreSQL-specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErrorToResult(op, pgErr)
	}

	// Unknown error - wrap as infrastructure error
	return result.Infrastructure(op, err)
}

// pgErrorToResult converts a pgconn.PgError to the appropriate Result error kind
func pgErrorToResult(op string, pgErr *pgconn.PgError) error {
	switch pgErr.Code {
	case ErrCodeUniqueViolation:
		// Duplicate key → Conflict
		// Extract constraint name for better error messages
		resource := "resource"
		if pgErr.ConstraintName != "" {
			resource = pgErr.ConstraintName
		}
		return result.Conflict(op, resource)

	case ErrCodeForeignKeyViolation:
		// Foreign key violation → Domain error (business rule)
		msg := "foreign key constraint violation"
		if pgErr.ConstraintName != "" {
			msg = pgErr.ConstraintName + " constraint violated"
		}
		return result.Domain(op, msg)

	case ErrCodeCheckViolation, ErrCodeNotNullViolation:
		// Check/Not null constraint → Domain error
		msg := pgErr.Message
		if msg == "" {
			msg = "constraint violation"
		}
		return result.Domain(op, msg)

	case ErrCodeDeadlockDetected, ErrCodeSerializationFailure:
		// Deadlock/serialization → Infrastructure (transient, can retry)
		return result.Infrastructure(op, pgErr)

	case ErrCodeQueryCanceled:
		// Query timeout → Infrastructure
		return result.Infrastructure(op, pgErr)

	case ErrCodeConnectionException, ErrCodeConnectionFailure:
		// Connection errors → Infrastructure
		return result.Infrastructure(op, pgErr)

	default:
		// Unknown PostgreSQL error → Infrastructure
		return result.Infrastructure(op, pgErr)
	}
}

// Helper functions for checking specific error types

// IsUniqueViolation checks if the error is a unique constraint violation
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == ErrCodeUniqueViolation
	}
	return false
}

// IsForeignKeyViolation checks if the error is a foreign key constraint violation
func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == ErrCodeForeignKeyViolation
	}
	return false
}

// IsDeadlock checks if the error is a deadlock
func IsDeadlock(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == ErrCodeDeadlockDetected
	}
	return false
}

// IsNotFound checks if the error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || result.Is(err, result.KindNotFound)
}

// IsConflict checks if the error is a conflict error
func IsConflict(err error) bool {
	return result.Is(err, result.KindConflict)
}

// GetConstraintName extracts the constraint name from a PostgreSQL error.
// Returns empty string if not a constraint violation or constraint name is not available.
func GetConstraintName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// GetTableName extracts the table name from a PostgreSQL error.
// Returns empty string if table name is not available.
func GetTableName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.TableName
	}
	return ""
}

// GetColumnName extracts the column name from a PostgreSQL error.
// Returns empty string if column name is not available.
func GetColumnName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ColumnName
	}
	return ""
}
