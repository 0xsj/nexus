// Package domain contains the core business logic for the Credential bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Credential-specific)
// ============================================================================

const (
	CodeCredentialNotFound      pkgerrors.Code = "CREDENTIAL_NOT_FOUND"
	CodeCredentialRevoked       pkgerrors.Code = "CREDENTIAL_REVOKED"
	CodeCredentialExpired       pkgerrors.Code = "CREDENTIAL_EXPIRED"
	CodeCredentialInvalid       pkgerrors.Code = "CREDENTIAL_INVALID"
	CodeClaimsValidationFailed  pkgerrors.Code = "CLAIMS_VALIDATION_FAILED"
	CodeSigningFailed           pkgerrors.Code = "SIGNING_FAILED"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrCredentialNotFound     = errors.New("credential not found")
	ErrCredentialRevoked      = errors.New("credential has been revoked")
	ErrCredentialExpired      = errors.New("credential has expired")
	ErrCredentialInvalid      = errors.New("credential is invalid")
	ErrClaimsValidationFailed = errors.New("claims validation failed")
	ErrSigningFailed          = errors.New("credential signing failed")
)

// ============================================================================
// Error Constructors
// ============================================================================

// CredentialNotFound creates a credential not found error.
func CredentialNotFound(operation string, credentialID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "credential").
		WithCode(CodeCredentialNotFound).
		WithMeta("credential_id", credentialID)
}

// CredentialRevoked creates a credential revoked error.
func CredentialRevoked(operation string, credentialID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "credential has been revoked").
		WithCode(CodeCredentialRevoked).
		WithMeta("credential_id", credentialID)
}

// CredentialExpired creates a credential expired error.
func CredentialExpired(operation string, credentialID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "credential has expired").
		WithCode(CodeCredentialExpired).
		WithMeta("credential_id", credentialID)
}

// CredentialInvalid creates a credential invalid error.
func CredentialInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "credential is invalid: "+reason).
		WithCode(CodeCredentialInvalid)
}

// ClaimsValidationFailed creates a claims validation failed error.
func ClaimsValidationFailed(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "claims validation failed: "+reason).
		WithCode(CodeClaimsValidationFailed)
}

// SigningFailed creates a credential signing failed error.
func SigningFailed(operation string, err error) *pkgerrors.Error {
	return pkgerrors.Infrastructure(operation, err).
		WithCode(CodeSigningFailed)
}
