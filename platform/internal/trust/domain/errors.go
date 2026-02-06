// Package domain contains the core business logic for the Trust bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Trust-specific)
// ============================================================================

const (
	CodeVouchNotFound        pkgerrors.Code = "VOUCH_NOT_FOUND"
	CodeVouchAlreadyAccepted pkgerrors.Code = "VOUCH_ALREADY_ACCEPTED"
	CodeVouchRevoked         pkgerrors.Code = "VOUCH_REVOKED"
	CodeVouchExpired         pkgerrors.Code = "VOUCH_EXPIRED"
	CodeSelfVouch            pkgerrors.Code = "SELF_VOUCH"
	CodeVouchInvalid         pkgerrors.Code = "VOUCH_INVALID"
	CodeVouchUnauthorized    pkgerrors.Code = "VOUCH_UNAUTHORIZED"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrVouchNotFound        = errors.New("vouch not found")
	ErrVouchAlreadyAccepted = errors.New("vouch has already been accepted")
	ErrVouchRevoked         = errors.New("vouch has been revoked")
	ErrVouchExpired         = errors.New("vouch has expired")
	ErrSelfVouch            = errors.New("cannot vouch for yourself")
	ErrVouchInvalid         = errors.New("vouch is invalid")
	ErrVouchUnauthorized    = errors.New("unauthorized vouch operation")
)

// ============================================================================
// Error Constructors
// ============================================================================

// VouchNotFound creates a vouch not found error.
func VouchNotFound(operation string, vouchID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "vouch").
		WithCode(CodeVouchNotFound).
		WithMeta("vouch_id", vouchID)
}

// VouchAlreadyAccepted creates a vouch already accepted error.
func VouchAlreadyAccepted(operation string, vouchID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "vouch has already been accepted").
		WithCode(CodeVouchAlreadyAccepted).
		WithMeta("vouch_id", vouchID)
}

// VouchRevoked creates a vouch revoked error.
func VouchRevoked(operation string, vouchID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "vouch has been revoked").
		WithCode(CodeVouchRevoked).
		WithMeta("vouch_id", vouchID)
}

// VouchExpired creates a vouch expired error.
func VouchExpired(operation string, vouchID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "vouch has expired").
		WithCode(CodeVouchExpired).
		WithMeta("vouch_id", vouchID)
}

// SelfVouch creates a self-vouch error.
func SelfVouch(operation string, userID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "cannot vouch for yourself").
		WithCode(CodeSelfVouch).
		WithMeta("user_id", userID)
}

// VouchInvalid creates a vouch invalid error.
func VouchInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "vouch is invalid: "+reason).
		WithCode(CodeVouchInvalid)
}

// VouchUnauthorized creates a vouch unauthorized error.
func VouchUnauthorized(operation string, vouchID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "unauthorized vouch operation").
		WithCode(CodeVouchUnauthorized).
		WithMeta("vouch_id", vouchID)
}
