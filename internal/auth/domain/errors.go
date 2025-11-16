package domain

import (
	"fmt"

	"github.com/0xsj/result"
)

// ============================================================================
// Validation Errors (Kind: KindValidation)
// ============================================================================

func ErrInvalidCredentials() error {
	return result.Validation("auth.credentials.validate", "invalid credentials", nil)
}

func ErrEmptyToken() error {
	return result.Validation("auth.token.validate", "token cannot be empty", nil)
}

func ErrInvalidDeviceID() error {
	return result.Validation("auth.device.validate", "invalid device ID", nil)
}

// ============================================================================
// Not Found Errors (Kind: KindNotFound)
// ============================================================================

func ErrSessionNotFound(sessionID string) error {
	return result.NotFound("auth.session.find", "session not found")
}

func ErrMagicLinkNotFound(token string) error {
	return result.NotFound("auth.magic_link.find", "magic link not found or expired")
}

func ErrRefreshTokenNotFound() error {
	return result.NotFound("auth.token.find", "refresh token not found")
}

func ErrOAuthConnectionNotFound(provider, userID string) error {
	return result.NotFound(
		"auth.oauth.find",
		fmt.Sprintf("oauth connection not found: %s for user %s", provider, userID),
	)
}

// ============================================================================
// Conflict Errors (Kind: KindConflict)
// ============================================================================

func ErrOAuthConnectionAlreadyExists(provider string) error {
	return result.Conflict(
		"auth.oauth.create",
		fmt.Sprintf("oauth connection already exists for provider: %s", provider),
	)
}

// ============================================================================
// Domain Errors (Kind: KindDomain)
// ============================================================================

func ErrSessionExpired(sessionID string) error {
	return result.Domain(
		"auth.session.validate",
		fmt.Sprintf("session expired: %s", sessionID),
	)
}

func ErrTokenExpired() error {
	return result.Domain("auth.token.validate", "token has expired")
}

func ErrMagicLinkExpired() error {
	return result.Domain("auth.magic_link.validate", "magic link has expired")
}

func ErrMagicLinkAlreadyUsed() error {
	return result.Domain("auth.magic_link.validate", "magic link has already been used")
}

func ErrRefreshTokenRevoked() error {
	return result.Domain("auth.token.validate", "refresh token has been revoked")
}

func ErrInvalidToken(reason string) error {
	return result.Domain(
		"auth.token.validate",
		fmt.Sprintf("invalid token: %s", reason),
	)
}

func ErrOAuthProviderNotSupported(provider string) error {
	return result.Domain(
		"auth.oauth.validate",
		fmt.Sprintf("oauth provider not supported: %s", provider),
	)
}

func ErrInvalidOAuthState() error {
	return result.Unauthorized("auth.oauth.validate", "invalid or expired OAuth state")
}

func ErrOAuthCodeExchangeFailed() error {
	return result.Domain("auth.oauth.exchange", "failed to exchange OAuth code for tokens")
}

func ErrOAuthUserInfoFailed() error {
	return result.Domain("auth.oauth.userinfo", "failed to fetch user info from OAuth provider")
}

// ============================================================================
// Unauthorized Errors (Kind: KindUnauthorized)
// ============================================================================

func ErrUnauthorized() error {
	return result.Unauthorized("auth.authorize", "unauthorized access")
}

func ErrInvalidPassword() error {
	return result.Unauthorized("auth.login.password", "invalid email or password")
}

func ErrTooManyLoginAttempts(retryAfter int) error {
	return result.Domain(
		"auth.login",
		fmt.Sprintf("too many login attempts, retry after %d seconds", retryAfter),
	)
}
