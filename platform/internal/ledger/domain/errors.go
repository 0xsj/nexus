package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Ledger-specific)
// ============================================================================

const (
	// Entry errors
	CodeEntryNotFound   pkgerrors.Code = "LEDGER_ENTRY_NOT_FOUND"
	CodeInvalidEntryID  pkgerrors.Code = "LEDGER_INVALID_ENTRY_ID"
	CodeEntryValidation pkgerrors.Code = "LEDGER_ENTRY_VALIDATION"

	// Actor errors
	CodeInvalidActorID   pkgerrors.Code = "LEDGER_INVALID_ACTOR_ID"
	CodeInvalidActorType pkgerrors.Code = "LEDGER_INVALID_ACTOR_TYPE"

	// Subject errors
	CodeInvalidSubjectID   pkgerrors.Code = "LEDGER_INVALID_SUBJECT_ID"
	CodeInvalidSubjectType pkgerrors.Code = "LEDGER_INVALID_SUBJECT_TYPE"

	// Event type errors
	CodeInvalidEventType pkgerrors.Code = "LEDGER_INVALID_EVENT_TYPE"

	// Metadata errors
	CodeInvalidMetadata pkgerrors.Code = "LEDGER_INVALID_METADATA"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	// Entry errors
	ErrEntryNotFound   = errors.New("audit entry not found")
	ErrInvalidEntryID  = errors.New("invalid entry ID")
	ErrEntryValidation = errors.New("entry validation failed")

	// Actor errors
	ErrInvalidActorID   = errors.New("invalid actor ID")
	ErrInvalidActorType = errors.New("invalid actor type")

	// Subject errors
	ErrInvalidSubjectID   = errors.New("invalid subject ID")
	ErrInvalidSubjectType = errors.New("invalid subject type")

	// Event type errors
	ErrInvalidEventType = errors.New("invalid event type")

	// Metadata errors
	ErrInvalidMetadata = errors.New("invalid metadata")
)

// ============================================================================
// Error Constructors
// ============================================================================

// EntryNotFound creates an entry not found error.
func EntryNotFound(operation string, entryID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "audit entry").
		WithCode(CodeEntryNotFound).
		WithMeta("entry_id", entryID)
}

// InvalidEntryID creates an invalid entry ID error.
func InvalidEntryID(operation string, value string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid entry ID: "+reason).
		WithCode(CodeInvalidEntryID).
		WithMeta("value", value)
}

// EntryValidation creates an entry validation error.
func EntryValidation(operation string, field string, reason string) *pkgerrors.Error {
	return pkgerrors.Validationf(operation, "invalid %s: %s", field, reason).
		WithCode(CodeEntryValidation).
		WithMeta("field", field)
}

// InvalidActorID creates an invalid actor ID error.
func InvalidActorID(operation string, value string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid actor ID: "+reason).
		WithCode(CodeInvalidActorID).
		WithMeta("value", value)
}

// InvalidActorType creates an invalid actor type error.
func InvalidActorType(operation string, value string) *pkgerrors.Error {
	return pkgerrors.Validationf(operation, "invalid actor type: %s", value).
		WithCode(CodeInvalidActorType).
		WithMeta("value", value)
}

// InvalidSubjectID creates an invalid subject ID error.
func InvalidSubjectID(operation string, value string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid subject ID: "+reason).
		WithCode(CodeInvalidSubjectID).
		WithMeta("value", value)
}

// InvalidSubjectType creates an invalid subject type error.
func InvalidSubjectType(operation string, value string) *pkgerrors.Error {
	return pkgerrors.Validationf(operation, "invalid subject type: %s", value).
		WithCode(CodeInvalidSubjectType).
		WithMeta("value", value)
}

// InvalidEventType creates an invalid event type error.
func InvalidEventType(operation string, value string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid event type: "+reason).
		WithCode(CodeInvalidEventType).
		WithMeta("value", value)
}

// InvalidMetadata creates an invalid metadata error.
func InvalidMetadata(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid metadata: "+reason).
		WithCode(CodeInvalidMetadata)
}
