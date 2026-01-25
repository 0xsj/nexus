package domain

import (
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// Integration domain error codes.
const (
	// Verification errors
	CodeVerificationNotFound      errors.Code = "VERIFICATION_NOT_FOUND"
	CodeVerificationAlreadyExists errors.Code = "VERIFICATION_ALREADY_EXISTS"
	CodeVerificationExpired       errors.Code = "VERIFICATION_EXPIRED"
	CodeVerificationFailed        errors.Code = "VERIFICATION_FAILED"
	CodeVerificationInvalidState  errors.Code = "VERIFICATION_INVALID_STATE"

	// Provider errors
	CodeProviderNotSupported errors.Code = "PROVIDER_NOT_SUPPORTED"
	CodeProviderUnavailable  errors.Code = "PROVIDER_UNAVAILABLE"
	CodeProviderAuthFailed   errors.Code = "PROVIDER_AUTH_FAILED"
	CodeProviderRateLimited  errors.Code = "PROVIDER_RATE_LIMITED"
	CodeProviderTokenExpired errors.Code = "PROVIDER_TOKEN_EXPIRED"
	CodeProviderTokenInvalid errors.Code = "PROVIDER_TOKEN_INVALID"
	CodeProviderFetchFailed  errors.Code = "PROVIDER_FETCH_FAILED"

	// OAuth errors
	CodeOAuthStateMismatch errors.Code = "OAUTH_STATE_MISMATCH"
	CodeOAuthCodeInvalid   errors.Code = "OAUTH_CODE_INVALID"
	CodeOAuthCodeExpired   errors.Code = "OAUTH_CODE_EXPIRED"

	// Credential issuance errors
	CodeCredentialIssuanceFailed errors.Code = "CREDENTIAL_ISSUANCE_FAILED"
	CodeInsufficientData         errors.Code = "INSUFFICIENT_DATA"
)

// ============================================================================
// Verification Errors
// ============================================================================

// ErrVerificationNotFound creates a verification not found error.
func ErrVerificationNotFound(operation string, id string) *errors.Error {
	return errors.NotFound(operation, "verification: "+id).
		WithCode(CodeVerificationNotFound).
		WithMeta("verification_id", id)
}

// ErrVerificationAlreadyExists creates a verification already exists error.
func ErrVerificationAlreadyExists(operation string, userID string, provider string) *errors.Error {
	return errors.Conflict(operation, "verification already exists for this provider").
		WithCode(CodeVerificationAlreadyExists).
		WithMeta("user_id", userID).
		WithMeta("provider", provider)
}

// ErrVerificationExpired creates a verification expired error.
func ErrVerificationExpired(operation string, id string) *errors.Error {
	return errors.Validation(operation, "verification has expired").
		WithCode(CodeVerificationExpired).
		WithMeta("verification_id", id)
}

// ErrVerificationFailed creates a verification failed error.
func ErrVerificationFailed(operation string, reason string) *errors.Error {
	return errors.Domain(operation, "verification failed: "+reason).
		WithCode(CodeVerificationFailed)
}

// ErrVerificationInvalidState creates an invalid state transition error.
func ErrVerificationInvalidState(operation string, current string, target string) *errors.Error {
	return errors.Domain(operation, "cannot transition from "+current+" to "+target).
		WithCode(CodeVerificationInvalidState).
		WithMeta("current_state", current).
		WithMeta("target_state", target)
}

// ============================================================================
// Provider Errors
// ============================================================================

// ErrProviderNotSupported creates a provider not supported error.
func ErrProviderNotSupported(operation string, provider string) *errors.Error {
	return errors.Validation(operation, "provider not supported: "+provider).
		WithCode(CodeProviderNotSupported).
		WithMeta("provider", provider)
}

// ErrProviderUnavailable creates a provider unavailable error.
func ErrProviderUnavailable(operation string, provider string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeProviderUnavailable).
		WithMeta("provider", provider).
		WithMessage("provider is temporarily unavailable: " + provider)
}

// ErrProviderAuthFailed creates a provider authentication failed error.
func ErrProviderAuthFailed(operation string, provider string, reason string) *errors.Error {
	return errors.Unauthorized(operation, "authentication with "+provider+" failed: "+reason).
		WithCode(CodeProviderAuthFailed).
		WithMeta("provider", provider)
}

// ErrProviderRateLimited creates a provider rate limited error.
func ErrProviderRateLimited(operation string, provider string) *errors.Error {
	return errors.Infrastructure(operation, nil).
		WithCode(CodeProviderRateLimited).
		WithMeta("provider", provider).
		WithMessage("rate limited by " + provider)
}

// ErrProviderTokenExpired creates a provider token expired error.
func ErrProviderTokenExpired(operation string, provider string) *errors.Error {
	return errors.Unauthorized(operation, "access token for "+provider+" has expired").
		WithCode(CodeProviderTokenExpired).
		WithMeta("provider", provider)
}

// ErrProviderTokenInvalid creates a provider token invalid error.
func ErrProviderTokenInvalid(operation string, provider string) *errors.Error {
	return errors.Unauthorized(operation, "access token for "+provider+" is invalid").
		WithCode(CodeProviderTokenInvalid).
		WithMeta("provider", provider)
}

// ErrProviderFetchFailed creates a provider data fetch failed error.
func ErrProviderFetchFailed(operation string, provider string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeProviderFetchFailed).
		WithMeta("provider", provider).
		WithMessage("failed to fetch data from " + provider)
}

// ============================================================================
// OAuth Errors
// ============================================================================

// ErrOAuthStateMismatch creates an OAuth state mismatch error.
func ErrOAuthStateMismatch(operation string) *errors.Error {
	return errors.Validation(operation, "OAuth state parameter mismatch").
		WithCode(CodeOAuthStateMismatch)
}

// ErrOAuthCodeInvalid creates an OAuth code invalid error.
func ErrOAuthCodeInvalid(operation string, provider string) *errors.Error {
	return errors.Validation(operation, "invalid authorization code").
		WithCode(CodeOAuthCodeInvalid).
		WithMeta("provider", provider)
}

// ErrOAuthCodeExpired creates an OAuth code expired error.
func ErrOAuthCodeExpired(operation string, provider string) *errors.Error {
	return errors.Validation(operation, "authorization code has expired").
		WithCode(CodeOAuthCodeExpired).
		WithMeta("provider", provider)
}

// ============================================================================
// Credential Issuance Errors
// ============================================================================

// ErrCredentialIssuanceFailed creates a credential issuance failed error.
func ErrCredentialIssuanceFailed(operation string, reason string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeCredentialIssuanceFailed).
		WithMessage("failed to issue credential: " + reason)
}

// ErrInsufficientData creates an insufficient data error.
func ErrInsufficientData(operation string, provider string, missing string) *errors.Error {
	return errors.Validation(operation, "insufficient data from "+provider+" to issue credential").
		WithCode(CodeInsufficientData).
		WithMeta("provider", provider).
		WithMeta("missing", missing)
}

// ============================================================================
// Error Checkers
// ============================================================================

// IsVerificationNotFound checks if error is verification not found.
func IsVerificationNotFound(err error) bool {
	return errors.GetCode(err) == CodeVerificationNotFound
}

// IsVerificationAlreadyExists checks if error is verification already exists.
func IsVerificationAlreadyExists(err error) bool {
	return errors.GetCode(err) == CodeVerificationAlreadyExists
}

// IsVerificationExpired checks if error is verification expired.
func IsVerificationExpired(err error) bool {
	return errors.GetCode(err) == CodeVerificationExpired
}

// IsVerificationFailed checks if error is verification failed.
func IsVerificationFailed(err error) bool {
	return errors.GetCode(err) == CodeVerificationFailed
}

// IsVerificationInvalidState checks if error is invalid state transition.
func IsVerificationInvalidState(err error) bool {
	return errors.GetCode(err) == CodeVerificationInvalidState
}

// IsProviderNotSupported checks if error is provider not supported.
func IsProviderNotSupported(err error) bool {
	return errors.GetCode(err) == CodeProviderNotSupported
}

// IsProviderUnavailable checks if error is provider unavailable.
func IsProviderUnavailable(err error) bool {
	return errors.GetCode(err) == CodeProviderUnavailable
}

// IsProviderAuthFailed checks if error is provider auth failed.
func IsProviderAuthFailed(err error) bool {
	return errors.GetCode(err) == CodeProviderAuthFailed
}

// IsProviderRateLimited checks if error is provider rate limited.
func IsProviderRateLimited(err error) bool {
	return errors.GetCode(err) == CodeProviderRateLimited
}

// IsProviderTokenExpired checks if error is provider token expired.
func IsProviderTokenExpired(err error) bool {
	return errors.GetCode(err) == CodeProviderTokenExpired
}

// IsProviderTokenInvalid checks if error is provider token invalid.
func IsProviderTokenInvalid(err error) bool {
	return errors.GetCode(err) == CodeProviderTokenInvalid
}

// IsProviderFetchFailed checks if error is provider fetch failed.
func IsProviderFetchFailed(err error) bool {
	return errors.GetCode(err) == CodeProviderFetchFailed
}

// IsOAuthStateMismatch checks if error is OAuth state mismatch.
func IsOAuthStateMismatch(err error) bool {
	return errors.GetCode(err) == CodeOAuthStateMismatch
}

// IsOAuthCodeInvalid checks if error is OAuth code invalid.
func IsOAuthCodeInvalid(err error) bool {
	return errors.GetCode(err) == CodeOAuthCodeInvalid
}

// IsOAuthCodeExpired checks if error is OAuth code expired.
func IsOAuthCodeExpired(err error) bool {
	return errors.GetCode(err) == CodeOAuthCodeExpired
}

// IsCredentialIssuanceFailed checks if error is credential issuance failed.
func IsCredentialIssuanceFailed(err error) bool {
	return errors.GetCode(err) == CodeCredentialIssuanceFailed
}

// IsInsufficientData checks if error is insufficient data.
func IsInsufficientData(err error) bool {
	return errors.GetCode(err) == CodeInsufficientData
}
