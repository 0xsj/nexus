// Package domain contains the core business logic for the Identity bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Identity-specific)
// ============================================================================

const (
	// User errors
	CodeUserNotFound      pkgerrors.Code = "USER_NOT_FOUND"
	CodeUserAlreadyExists pkgerrors.Code = "USER_ALREADY_EXISTS"
	CodeUserSuspended     pkgerrors.Code = "USER_SUSPENDED"
	CodeUserNotActive     pkgerrors.Code = "USER_NOT_ACTIVE"

	// Email errors
	CodeEmailAlreadyRegistered pkgerrors.Code = "EMAIL_ALREADY_REGISTERED"
	CodeEmailNotVerified       pkgerrors.Code = "EMAIL_NOT_VERIFIED"
	CodeInvalidEmail           pkgerrors.Code = "INVALID_EMAIL"

	// Session errors
	CodeSessionNotFound pkgerrors.Code = "SESSION_NOT_FOUND"
	CodeSessionExpired  pkgerrors.Code = "SESSION_EXPIRED"
	CodeSessionRevoked  pkgerrors.Code = "SESSION_REVOKED"
	CodeInvalidToken    pkgerrors.Code = "INVALID_TOKEN"

	// Magic link errors
	CodeMagicLinkExpired  pkgerrors.Code = "MAGIC_LINK_EXPIRED"
	CodeMagicLinkUsed     pkgerrors.Code = "MAGIC_LINK_USED"
	CodeMagicLinkInvalid  pkgerrors.Code = "MAGIC_LINK_INVALID"
	CodeMagicLinkNotFound pkgerrors.Code = "MAGIC_LINK_NOT_FOUND"

	// OAuth errors
	CodeOAuthProviderNotSupported pkgerrors.Code = "OAUTH_PROVIDER_NOT_SUPPORTED"
	CodeOAuthStateMismatch        pkgerrors.Code = "OAUTH_STATE_MISMATCH"
	CodeOAuthCodeExchangeFailed   pkgerrors.Code = "OAUTH_CODE_EXCHANGE_FAILED"
	CodeOAuthAccountLinked        pkgerrors.Code = "OAUTH_ACCOUNT_ALREADY_LINKED"

	// Auth method errors
	CodeAuthMethodNotEnabled pkgerrors.Code = "AUTH_METHOD_NOT_ENABLED"
	CodeAuthMethodExists     pkgerrors.Code = "AUTH_METHOD_ALREADY_EXISTS"

	// DID errors
	CodeDIDAlreadyLinked pkgerrors.Code = "DID_ALREADY_LINKED"
	CodeDIDNotFound      pkgerrors.Code = "DID_NOT_FOUND"
	CodePrimaryDIDChange pkgerrors.Code = "PRIMARY_DID_CANNOT_CHANGE"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserSuspended     = errors.New("user is suspended")
	ErrUserNotActive     = errors.New("user is not active")

	// Email errors
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrEmailNotVerified       = errors.New("email not verified")

	// Session errors
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionRevoked  = errors.New("session revoked")
	ErrInvalidToken    = errors.New("invalid token")

	// Magic link errors
	ErrMagicLinkExpired  = errors.New("magic link expired")
	ErrMagicLinkUsed     = errors.New("magic link already used")
	ErrMagicLinkInvalid  = errors.New("magic link invalid")
	ErrMagicLinkNotFound = errors.New("magic link not found")

	// OAuth errors
	ErrOAuthProviderNotSupported = errors.New("oauth provider not supported")
	ErrOAuthStateMismatch        = errors.New("oauth state mismatch")
	ErrOAuthCodeExchangeFailed   = errors.New("oauth code exchange failed")
	ErrOAuthAccountLinked        = errors.New("oauth account already linked")

	// Auth method errors
	ErrAuthMethodNotEnabled = errors.New("auth method not enabled")
	ErrAuthMethodExists     = errors.New("auth method already exists")

	// DID errors
	ErrDIDAlreadyLinked = errors.New("DID already linked to another user")
	ErrDIDNotFound      = errors.New("DID not found")
	ErrPrimaryDIDChange = errors.New("primary DID cannot be changed")
)

// ============================================================================
// Error Constructors
// ============================================================================

// UserNotFound creates a user not found error.
func UserNotFound(operation string, userID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "user").
		WithCode(CodeUserNotFound).
		WithMeta("user_id", userID)
}

// UserAlreadyExists creates a user already exists error.
func UserAlreadyExists(operation string, identifier string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "user").
		WithCode(CodeUserAlreadyExists).
		WithMeta("identifier", identifier)
}

// UserSuspended creates a user suspended error.
func UserSuspended(operation string, userID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "user account is suspended").
		WithCode(CodeUserSuspended).
		WithMeta("user_id", userID)
}

// UserNotActive creates a user not active error.
func UserNotActive(operation string, userID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "user account is not active").
		WithCode(CodeUserNotActive).
		WithMeta("user_id", userID)
}

// EmailAlreadyRegistered creates an email already registered error.
func EmailAlreadyRegistered(operation string, email string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "email").
		WithCode(CodeEmailAlreadyRegistered).
		WithMeta("email", email)
}

// EmailNotVerified creates an email not verified error.
func EmailNotVerified(operation string, email string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "email address not verified").
		WithCode(CodeEmailNotVerified).
		WithMeta("email", email)
}

// InvalidEmail creates an invalid email error.
func InvalidEmail(operation string, email string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid email: "+reason).
		WithCode(CodeInvalidEmail).
		WithMeta("email", email)
}

// SessionNotFound creates a session not found error.
func SessionNotFound(operation string, sessionID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "session").
		WithCode(CodeSessionNotFound).
		WithMeta("session_id", sessionID)
}

// SessionExpired creates a session expired error.
func SessionExpired(operation string, sessionID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "session has expired").
		WithCode(CodeSessionExpired).
		WithMeta("session_id", sessionID)
}

// SessionRevoked creates a session revoked error.
func SessionRevoked(operation string, sessionID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "session has been revoked").
		WithCode(CodeSessionRevoked).
		WithMeta("session_id", sessionID)
}

// InvalidToken creates an invalid token error.
func InvalidToken(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Unauthorized(operation, "invalid token: "+reason).
		WithCode(CodeInvalidToken)
}

// MagicLinkExpired creates a magic link expired error.
func MagicLinkExpired(operation string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "magic link has expired").
		WithCode(CodeMagicLinkExpired)
}

// MagicLinkUsed creates a magic link already used error.
func MagicLinkUsed(operation string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "magic link has already been used").
		WithCode(CodeMagicLinkUsed)
}

// MagicLinkInvalid creates a magic link invalid error.
func MagicLinkInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid magic link: "+reason).
		WithCode(CodeMagicLinkInvalid)
}

// MagicLinkNotFound creates a magic link not found error.
func MagicLinkNotFound(operation string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "magic link").
		WithCode(CodeMagicLinkNotFound)
}

// OAuthProviderNotSupported creates an unsupported oauth provider error.
func OAuthProviderNotSupported(operation string, provider string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "oauth provider not supported: "+provider).
		WithCode(CodeOAuthProviderNotSupported).
		WithMeta("provider", provider)
}

// OAuthStateMismatch creates an oauth state mismatch error.
func OAuthStateMismatch(operation string) *pkgerrors.Error {
	return pkgerrors.Unauthorized(operation, "oauth state mismatch").
		WithCode(CodeOAuthStateMismatch)
}

// OAuthCodeExchangeFailed creates an oauth code exchange failure error.
func OAuthCodeExchangeFailed(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Infrastructure(operation, errors.New(reason)).
		WithCode(CodeOAuthCodeExchangeFailed)
}

// OAuthAccountLinked creates an oauth account already linked error.
func OAuthAccountLinked(operation string, provider string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "oauth account").
		WithCode(CodeOAuthAccountLinked).
		WithMeta("provider", provider)
}

// AuthMethodNotEnabled creates an auth method not enabled error.
func AuthMethodNotEnabled(operation string, method string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "authentication method not enabled").
		WithCode(CodeAuthMethodNotEnabled).
		WithMeta("method", method)
}

// AuthMethodExists creates an auth method already exists error.
func AuthMethodExists(operation string, method string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "auth method").
		WithCode(CodeAuthMethodExists).
		WithMeta("method", method)
}

// DIDAlreadyLinked creates a DID already linked error.
func DIDAlreadyLinked(operation string, did string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "DID").
		WithCode(CodeDIDAlreadyLinked).
		WithMeta("did", did)
}

// DIDNotFound creates a DID not found error.
func DIDNotFound(operation string, did string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "DID").
		WithCode(CodeDIDNotFound).
		WithMeta("did", did)
}

// PrimaryDIDCannotChange creates an error for attempts to change primary DID.
func PrimaryDIDCannotChange(operation string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "primary DID cannot be changed after creation").
		WithCode(CodePrimaryDIDChange)
}
