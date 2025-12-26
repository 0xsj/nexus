package eventsourcing

import (
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/errors"
)

// Event sourcing specific error codes.
const (
	CodeEventValidation     errors.Code = "EVENT_VALIDATION"
	CodeEventNotFound       errors.Code = "EVENT_NOT_FOUND"
	CodeEventStoreFailed    errors.Code = "EVENT_STORE_FAILED"
	CodeEventLoadFailed     errors.Code = "EVENT_LOAD_FAILED"
	CodeConcurrencyConflict errors.Code = "CONCURRENCY_CONFLICT"

	CodeAggregateNotFound   errors.Code = "AGGREGATE_NOT_FOUND"
	CodeAggregateValidation errors.Code = "AGGREGATE_VALIDATION"

	CodeSnapshotFailed   errors.Code = "SNAPSHOT_FAILED"
	CodeSnapshotNotFound errors.Code = "SNAPSHOT_NOT_FOUND"

	CodeProjectionFailed errors.Code = "PROJECTION_FAILED"
)

// ============================================================================
// Event Errors
// ============================================================================

// ErrEventValidation creates an event validation error.
func ErrEventValidation(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeEventValidation)
}

// ErrEventNotFound creates an error when events are not found.
func ErrEventNotFound(operation string, aggregateID string) *errors.Error {
	return errors.NotFound(operation, "events for aggregate: "+aggregateID).
		WithCode(CodeEventNotFound).
		WithMeta("aggregate_id", aggregateID)
}

// ErrEventStoreFailed creates an error when storing events fails.
func ErrEventStoreFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeEventStoreFailed)
}

// ErrEventLoadFailed creates an error when loading events fails.
func ErrEventLoadFailed(operation string, aggregateID string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeEventLoadFailed).
		WithMeta("aggregate_id", aggregateID)
}

// ============================================================================
// Concurrency Errors
// ============================================================================

// ErrConcurrencyConflict creates an optimistic concurrency error.
func ErrConcurrencyConflict(operation string, aggregateID string, expected, actual int) *errors.Error {
	return errors.Conflict(operation, "aggregate: "+aggregateID).
		WithCode(CodeConcurrencyConflict).
		WithMeta("aggregate_id", aggregateID).
		WithMeta("expected_version", expected).
		WithMeta("actual_version", actual).
		WithMessage(fmt.Sprintf("concurrency conflict: expected version %d, got %d", expected, actual))
}

// ============================================================================
// Aggregate Errors
// ============================================================================

// ErrAggregateNotFound creates an error when an aggregate is not found.
func ErrAggregateNotFound(operation string, aggregateType string, aggregateID string) *errors.Error {
	return errors.NotFound(operation, aggregateType+": "+aggregateID).
		WithCode(CodeAggregateNotFound).
		WithMeta("aggregate_type", aggregateType).
		WithMeta("aggregate_id", aggregateID)
}

// ErrAggregateValidation creates an aggregate validation error.
func ErrAggregateValidation(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeAggregateValidation)
}

// ============================================================================
// Snapshot Errors
// ============================================================================

// ErrSnapshotFailed creates an error when snapshotting fails.
func ErrSnapshotFailed(operation string, aggregateID string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeSnapshotFailed).
		WithMeta("aggregate_id", aggregateID)
}

// ErrSnapshotNotFound creates an error when a snapshot is not found.
func ErrSnapshotNotFound(operation string, aggregateID string) *errors.Error {
	return errors.NotFound(operation, "snapshot for aggregate: "+aggregateID).
		WithCode(CodeSnapshotNotFound).
		WithMeta("aggregate_id", aggregateID)
}

// ============================================================================
// Projection Errors
// ============================================================================

// ErrProjectionFailed creates an error when a projection fails.
func ErrProjectionFailed(operation string, projectionName string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeProjectionFailed).
		WithMeta("projection", projectionName)
}

// ============================================================================
// Error Checkers
// ============================================================================

// IsConcurrencyConflict returns true if the error is a concurrency conflict.
func IsConcurrencyConflict(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeConcurrencyConflict
	}
	return false
}

// IsAggregateNotFound returns true if the error is an aggregate not found error.
func IsAggregateNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeAggregateNotFound
	}
	return false
}

// IsEventNotFound returns true if the error is an event not found error.
func IsEventNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeEventNotFound
	}
	return false
}

// IsSnapshotNotFound returns true if the error is a snapshot not found error.
func IsSnapshotNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeSnapshotNotFound
	}
	return false
}

// IsProjectionFailed returns true if the error is a projection failure.
func IsProjectionFailed(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeProjectionFailed
	}
	return false
}
