// Package domain contains the core business logic for the Verification bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Verification-specific)
// ============================================================================

const (
	CodeVerificationNotFound         pkgerrors.Code = "VERIFICATION_NOT_FOUND"
	CodeVerificationFailed           pkgerrors.Code = "VERIFICATION_FAILED"
	CodeOAuthStateMismatch           pkgerrors.Code = "OAUTH_STATE_MISMATCH"
	CodeOAuthCallbackFailed          pkgerrors.Code = "OAUTH_CALLBACK_FAILED"
	CodeDataFetchFailed              pkgerrors.Code = "DATA_FETCH_FAILED"
	CodeProviderNotSupported         pkgerrors.Code = "PROVIDER_NOT_SUPPORTED"
	CodeVerificationAlreadyCompleted pkgerrors.Code = "VERIFICATION_ALREADY_COMPLETED"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrVerificationNotFound         = errors.New("verification not found")
	ErrVerificationFailed           = errors.New("verification failed")
	ErrOAuthStateMismatch           = errors.New("oauth state mismatch")
	ErrOAuthCallbackFailed          = errors.New("oauth callback failed")
	ErrDataFetchFailed              = errors.New("data fetch failed")
	ErrProviderNotSupported         = errors.New("provider not supported")
	ErrVerificationAlreadyCompleted = errors.New("verification already completed")
)

// ============================================================================
// Error Constructors
// ============================================================================

// VerificationNotFound creates a verification not found error.
func VerificationNotFound(operation string, verificationID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "verification").
		WithCode(CodeVerificationNotFound).
		WithMeta("verification_id", verificationID)
}

// VerificationFailed creates a verification failed error.
func VerificationFailed(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "verification failed: "+reason).
		WithCode(CodeVerificationFailed)
}

// OAuthStateMismatch creates an oauth state mismatch error.
func OAuthStateMismatch(operation string) *pkgerrors.Error {
	return pkgerrors.Unauthorized(operation, "oauth state mismatch").
		WithCode(CodeOAuthStateMismatch)
}

// OAuthCallbackFailed creates an oauth callback failure error.
func OAuthCallbackFailed(operation string, err error) *pkgerrors.Error {
	return pkgerrors.Infrastructure(operation, err).
		WithCode(CodeOAuthCallbackFailed)
}

// DataFetchFailed creates a data fetch failure error.
func DataFetchFailed(operation string, err error) *pkgerrors.Error {
	return pkgerrors.Infrastructure(operation, err).
		WithCode(CodeDataFetchFailed)
}

// ProviderNotSupported creates a provider not supported error.
func ProviderNotSupported(operation string, provider string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "provider not supported: "+provider).
		WithCode(CodeProviderNotSupported).
		WithMeta("provider", provider)
}

// VerificationAlreadyCompleted creates a verification already completed error.
func VerificationAlreadyCompleted(operation string, verificationID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "verification has already completed").
		WithCode(CodeVerificationAlreadyCompleted).
		WithMeta("verification_id", verificationID)
}
