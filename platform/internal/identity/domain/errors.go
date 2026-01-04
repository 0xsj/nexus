package domain

import (
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// Identity domain error codes.
const (
	CodeUserNotFound       errors.Code = "USER_NOT_FOUND"
	CodeUserAlreadyExists  errors.Code = "USER_ALREADY_EXISTS"
	CodeUserDisabled       errors.Code = "USER_DISABLED"
	CodeUserNotVerified    errors.Code = "USER_NOT_VERIFIED"
	CodeInvalidCredentials errors.Code = "INVALID_CREDENTIALS"
	CodeInvalidEmail       errors.Code = "INVALID_EMAIL"
	CodeInvalidPassword    errors.Code = "INVALID_PASSWORD"
	CodePasswordTooWeak    errors.Code = "PASSWORD_TOO_WEAK"

	CodeSessionNotFound errors.Code = "SESSION_NOT_FOUND"
	CodeSessionExpired  errors.Code = "SESSION_EXPIRED"
	CodeSessionRevoked  errors.Code = "SESSION_REVOKED"

	CodeTokenInvalid errors.Code = "TOKEN_INVALID"
	CodeTokenExpired errors.Code = "TOKEN_EXPIRED"

	CodeConnectionNotFound      errors.Code = "CONNECTION_NOT_FOUND"
	CodeConnectionAlreadyExists errors.Code = "CONNECTION_ALREADY_EXISTS"
	CodeConnectionFailed        errors.Code = "CONNECTION_FAILED"

	CodeAPIKeyNotFound errors.Code = "API_KEY_NOT_FOUND"
	CodeAPIKeyExpired  errors.Code = "API_KEY_EXPIRED"
	CodeAPIKeyRevoked  errors.Code = "API_KEY_REVOKED"

	CodeChallengeNotFound errors.Code = "CHALLENGE_NOT_FOUND"
	CodeChallengeExpired  errors.Code = "CHALLENGE_EXPIRED"
	CodeChallengeInvalid  errors.Code = "CHALLENGE_INVALID"
	CodeChallengeUsed     errors.Code = "CHALLENGE_USED"

	CodeWalletNotFound      errors.Code = "WALLET_NOT_FOUND"
	CodeWalletAlreadyLinked errors.Code = "WALLET_ALREADY_LINKED"
	CodeInvalidSignature    errors.Code = "INVALID_SIGNATURE"
)

// ============================================================================
// User Errors
// ============================================================================

// ErrUserNotFound creates a user not found error.
func ErrUserNotFound(operation string, identifier string) *errors.Error {
	return errors.NotFound(operation, "user: "+identifier).
		WithCode(CodeUserNotFound).
		WithMeta("identifier", identifier)
}

// ErrUserAlreadyExists creates a user already exists error.
func ErrUserAlreadyExists(operation string, email string) *errors.Error {
	return errors.Conflict(operation, "user with email already exists").
		WithCode(CodeUserAlreadyExists).
		WithMeta("email", email)
}

// ErrUserDisabled creates a user disabled error.
func ErrUserDisabled(operation string, userID string) *errors.Error {
	return errors.Forbidden(operation, "user account is disabled").
		WithCode(CodeUserDisabled).
		WithMeta("user_id", userID)
}

// ErrUserNotVerified creates a user not verified error.
func ErrUserNotVerified(operation string, userID string) *errors.Error {
	return errors.Forbidden(operation, "user email not verified").
		WithCode(CodeUserNotVerified).
		WithMeta("user_id", userID)
}

// ErrInvalidCredentials creates an invalid credentials error.
func ErrInvalidCredentials(operation string) *errors.Error {
	return errors.Unauthorized(operation, "invalid email or password").
		WithCode(CodeInvalidCredentials)
}

// ErrInvalidEmail creates an invalid email error.
func ErrInvalidEmail(operation string, email string) *errors.Error {
	return errors.Validation(operation, "invalid email format").
		WithCode(CodeInvalidEmail).
		WithMeta("email", email)
}

// ErrInvalidPassword creates an invalid password error.
func ErrInvalidPassword(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidPassword)
}

// ErrPasswordTooWeak creates a password too weak error.
func ErrPasswordTooWeak(operation string, requirements string) *errors.Error {
	return errors.Validation(operation, "password does not meet requirements: "+requirements).
		WithCode(CodePasswordTooWeak)
}

// ============================================================================
// Session Errors
// ============================================================================

// ErrSessionNotFound creates a session not found error.
func ErrSessionNotFound(operation string, sessionID string) *errors.Error {
	return errors.NotFound(operation, "session: "+sessionID).
		WithCode(CodeSessionNotFound).
		WithMeta("session_id", sessionID)
}

// ErrSessionExpired creates a session expired error.
func ErrSessionExpired(operation string, sessionID string) *errors.Error {
	return errors.Unauthorized(operation, "session has expired").
		WithCode(CodeSessionExpired).
		WithMeta("session_id", sessionID)
}

// ErrSessionRevoked creates a session revoked error.
func ErrSessionRevoked(operation string, sessionID string) *errors.Error {
	return errors.Unauthorized(operation, "session has been revoked").
		WithCode(CodeSessionRevoked).
		WithMeta("session_id", sessionID)
}

// ============================================================================
// Token Errors
// ============================================================================

// ErrTokenInvalid creates a token invalid error.
func ErrTokenInvalid(operation string, reason string) *errors.Error {
	return errors.Unauthorized(operation, "invalid token: "+reason).
		WithCode(CodeTokenInvalid)
}

// ErrTokenExpired creates a token expired error.
func ErrTokenExpired(operation string) *errors.Error {
	return errors.Unauthorized(operation, "token has expired").
		WithCode(CodeTokenExpired)
}

// ============================================================================
// Challenge Errors
// ============================================================================

// ErrChallengeNotFound creates a challenge not found error.
func ErrChallengeNotFound(operation string, nonce string) *errors.Error {
	return errors.NotFound(operation, "challenge").
		WithCode(CodeChallengeNotFound).
		WithMeta("nonce", nonce)
}

// ErrChallengeExpired creates a challenge expired error.
func ErrChallengeExpired(operation string, nonce string) *errors.Error {
	return errors.Unauthorized(operation, "challenge has expired").
		WithCode(CodeChallengeExpired).
		WithMeta("nonce", nonce)
}

// ErrChallengeInvalid creates a challenge invalid error.
func ErrChallengeInvalid(operation string, nonce string) *errors.Error {
	return errors.Unauthorized(operation, "invalid challenge").
		WithCode(CodeChallengeInvalid).
		WithMeta("nonce", nonce)
}

// ErrChallengeUsed creates a challenge already used error.
func ErrChallengeUsed(operation string, nonce string) *errors.Error {
	return errors.Unauthorized(operation, "challenge has already been used").
		WithCode(CodeChallengeUsed).
		WithMeta("nonce", nonce)
}

// ============================================================================
// Connection Errors
// ============================================================================

// ErrConnectionNotFound creates a connection not found error.
func ErrConnectionNotFound(operation string, provider string, userID string) *errors.Error {
	return errors.NotFound(operation, "connection not found").
		WithCode(CodeConnectionNotFound).
		WithMeta("provider", provider).
		WithMeta("user_id", userID)
}

// ErrConnectionAlreadyExists creates a connection already exists error.
func ErrConnectionAlreadyExists(operation string, provider string, userID string) *errors.Error {
	return errors.Conflict(operation, "connection already exists for this provider").
		WithCode(CodeConnectionAlreadyExists).
		WithMeta("provider", provider).
		WithMeta("user_id", userID)
}

// ErrConnectionFailed creates a connection failed error.
func ErrConnectionFailed(operation string, provider string, reason string) *errors.Error {
	return errors.Infrastructure(operation, nil).
		WithCode(CodeConnectionFailed).
		WithMeta("provider", provider).
		WithMessage("failed to connect to " + provider + ": " + reason)
}

// ============================================================================
// API Key Errors
// ============================================================================

// ErrAPIKeyNotFound creates an API key not found error.
func ErrAPIKeyNotFound(operation string, keyID string) *errors.Error {
	return errors.NotFound(operation, "api key: "+keyID).
		WithCode(CodeAPIKeyNotFound).
		WithMeta("key_id", keyID)
}

// ErrAPIKeyExpired creates an API key expired error.
func ErrAPIKeyExpired(operation string, keyID string) *errors.Error {
	return errors.Unauthorized(operation, "api key has expired").
		WithCode(CodeAPIKeyExpired).
		WithMeta("key_id", keyID)
}

// ErrAPIKeyRevoked creates an API key revoked error.
func ErrAPIKeyRevoked(operation string, keyID string) *errors.Error {
	return errors.Unauthorized(operation, "api key has been revoked").
		WithCode(CodeAPIKeyRevoked).
		WithMeta("key_id", keyID)
}

// ============================================================================
// Wallet Errors
// ============================================================================

// ErrWalletNotFound creates a wallet not found error.
func ErrWalletNotFound(operation string, address string) *errors.Error {
	return errors.NotFound(operation, "wallet: "+address).
		WithCode(CodeWalletNotFound).
		WithMeta("address", address)
}

// ErrWalletAlreadyLinked creates a wallet already linked error.
func ErrWalletAlreadyLinked(operation string, address string) *errors.Error {
	return errors.Conflict(operation, "wallet is already linked to an account").
		WithCode(CodeWalletAlreadyLinked).
		WithMeta("address", address)
}

// ErrInvalidSignature creates an invalid signature error.
func ErrInvalidSignature(operation string) *errors.Error {
	return errors.Unauthorized(operation, "invalid signature").
		WithCode(CodeInvalidSignature)
}

// ============================================================================
// Error Checkers
// ============================================================================

// IsUserNotFound checks if error is user not found.
func IsUserNotFound(err error) bool {
	return errors.GetCode(err) == CodeUserNotFound
}

// IsUserAlreadyExists checks if error is user already exists.
func IsUserAlreadyExists(err error) bool {
	return errors.GetCode(err) == CodeUserAlreadyExists
}

// IsUserDisabled checks if error is user disabled.
func IsUserDisabled(err error) bool {
	return errors.GetCode(err) == CodeUserDisabled
}

// IsInvalidCredentials checks if error is invalid credentials.
func IsInvalidCredentials(err error) bool {
	return errors.GetCode(err) == CodeInvalidCredentials
}

// IsSessionNotFound checks if error is session not found.
func IsSessionNotFound(err error) bool {
	return errors.GetCode(err) == CodeSessionNotFound
}

// IsSessionExpired checks if error is session expired.
func IsSessionExpired(err error) bool {
	return errors.GetCode(err) == CodeSessionExpired
}

// IsSessionRevoked checks if error is session revoked.
func IsSessionRevoked(err error) bool {
	return errors.GetCode(err) == CodeSessionRevoked
}

// IsTokenExpired checks if error is token expired.
func IsTokenExpired(err error) bool {
	return errors.GetCode(err) == CodeTokenExpired
}

// IsTokenInvalid checks if error is token invalid.
func IsTokenInvalid(err error) bool {
	return errors.GetCode(err) == CodeTokenInvalid
}

// IsChallengeNotFound checks if error is challenge not found.
func IsChallengeNotFound(err error) bool {
	return errors.GetCode(err) == CodeChallengeNotFound
}

// IsChallengeExpired checks if error is challenge expired.
func IsChallengeExpired(err error) bool {
	return errors.GetCode(err) == CodeChallengeExpired
}

// IsChallengeInvalid checks if error is challenge invalid.
func IsChallengeInvalid(err error) bool {
	return errors.GetCode(err) == CodeChallengeInvalid
}

// IsChallengeUsed checks if error is challenge already used.
func IsChallengeUsed(err error) bool {
	return errors.GetCode(err) == CodeChallengeUsed
}

// IsConnectionNotFound checks if error is connection not found.
func IsConnectionNotFound(err error) bool {
	return errors.GetCode(err) == CodeConnectionNotFound
}

// IsConnectionAlreadyExists checks if error is connection already exists.
func IsConnectionAlreadyExists(err error) bool {
	return errors.GetCode(err) == CodeConnectionAlreadyExists
}

// IsAPIKeyNotFound checks if error is API key not found.
func IsAPIKeyNotFound(err error) bool {
	return errors.GetCode(err) == CodeAPIKeyNotFound
}

// IsAPIKeyExpired checks if error is API key expired.
func IsAPIKeyExpired(err error) bool {
	return errors.GetCode(err) == CodeAPIKeyExpired
}

// IsAPIKeyRevoked checks if error is API key revoked.
func IsAPIKeyRevoked(err error) bool {
	return errors.GetCode(err) == CodeAPIKeyRevoked
}

// IsWalletNotFound checks if error is wallet not found.
func IsWalletNotFound(err error) bool {
	return errors.GetCode(err) == CodeWalletNotFound
}

// IsWalletAlreadyLinked checks if error is wallet already linked.
func IsWalletAlreadyLinked(err error) bool {
	return errors.GetCode(err) == CodeWalletAlreadyLinked
}

// IsInvalidSignature checks if error is invalid signature.
func IsInvalidSignature(err error) bool {
	return errors.GetCode(err) == CodeInvalidSignature
}
