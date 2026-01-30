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
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	// Entry errors
	ErrEntryNotFound   = errors.New("audit entry not found")
	ErrInvalidEntryID  = errors.New("invalid entry ID")
	ErrEntryValidation = errors.New("entry validation failed")
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
