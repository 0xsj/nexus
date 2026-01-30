package domain

import "github.com/0xsj/nexus/platform/pkg/errors"

// Domain error codes for the Ledger context.
const (
	// CodeEntryNotFound indicates an audit entry was not found.
	CodeEntryNotFound errors.Code = "LEDGER_ENTRY_NOT_FOUND"

	// CodeInvalidEntryID indicates an invalid entry ID format.
	CodeInvalidEntryID errors.Code = "LEDGER_INVALID_ENTRY_ID"

	// CodeInvalidActorType indicates an invalid actor type.
	CodeInvalidActorType errors.Code = "LEDGER_INVALID_ACTOR_TYPE"

	// CodeInvalidSubjectType indicates an invalid subject type.
	CodeInvalidSubjectType errors.Code = "LEDGER_INVALID_SUBJECT_TYPE"

	// CodeInvalidEventType indicates an invalid event type format.
	CodeInvalidEventType errors.Code = "LEDGER_INVALID_EVENT_TYPE"

	// CodeInvalidMetadata indicates invalid metadata.
	CodeInvalidMetadata errors.Code = "LEDGER_INVALID_METADATA"

	// CodeEntryValidation indicates an entry validation failure.
	CodeEntryValidation errors.Code = "LEDGER_ENTRY_VALIDATION"
)

// Sentinel errors for the Ledger context.
var (
	// ErrEntryNotFound indicates an audit entry was not found.
	ErrEntryNotFound = errors.New(errors.KindNotFound, CodeEntryNotFound, "audit entry not found")

	// ErrInvalidEntryID indicates an invalid entry ID.
	ErrInvalidEntryID = errors.New(errors.KindValidation, CodeInvalidEntryID, "invalid entry ID")

	// ErrInvalidActorType indicates an invalid actor type.
	ErrInvalidActorType = errors.New(errors.KindValidation, CodeInvalidActorType, "invalid actor type")

	// ErrInvalidSubjectType indicates an invalid subject type.
	ErrInvalidSubjectType = errors.New(errors.KindValidation, CodeInvalidSubjectType, "invalid subject type")

	// ErrInvalidEventType indicates an invalid event type format.
	ErrInvalidEventType = errors.New(errors.KindValidation, CodeInvalidEventType, "invalid event type")

	// ErrInvalidMetadata indicates invalid metadata.
	ErrInvalidMetadata = errors.New(errors.KindValidation, CodeInvalidMetadata, "invalid metadata")
)

// EntryNotFound returns a not found error for a specific entry ID.
func EntryNotFound(entryID string) *errors.Error {
	return errors.NotFound("ledger.Domain", "audit entry").
		WithCode(CodeEntryNotFound).
		WithMeta("entry_id", entryID)
}

// InvalidEntryID returns a validation error for an invalid entry ID.
func InvalidEntryID(value string, reason string) *errors.Error {
	return errors.Validation("ledger.Domain", reason).
		WithCode(CodeInvalidEntryID).
		WithMeta("value", value)
}

// InvalidActorType returns a validation error for an invalid actor type.
func InvalidActorType(value string) *errors.Error {
	return errors.Validationf("ledger.Domain", "invalid actor type: %s", value).
		WithCode(CodeInvalidActorType).
		WithMeta("value", value)
}

// InvalidSubjectType returns a validation error for an invalid subject type.
func InvalidSubjectType(value string) *errors.Error {
	return errors.Validationf("ledger.Domain", "invalid subject type: %s", value).
		WithCode(CodeInvalidSubjectType).
		WithMeta("value", value)
}

// InvalidEventType returns a validation error for an invalid event type.
func InvalidEventType(value string, reason string) *errors.Error {
	return errors.Validation("ledger.Domain", reason).
		WithCode(CodeInvalidEventType).
		WithMeta("value", value)
}

// InvalidMetadata returns a validation error for invalid metadata.
func InvalidMetadata(reason string) *errors.Error {
	return errors.Validation("ledger.Domain", reason).
		WithCode(CodeInvalidMetadata)
}

// EntryValidation returns a validation error for entry construction.
func EntryValidation(field string, reason string) *errors.Error {
	return errors.Validationf("ledger.Domain", "invalid %s: %s", field, reason).
		WithCode(CodeEntryValidation).
		WithMeta("field", field)
}
