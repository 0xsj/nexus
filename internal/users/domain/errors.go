package domain

import (
	"fmt"

	"github.com/0xsj/result"
)

// ============================================================================
// Validation Errors (Kind: KindValidation)
// ============================================================================

// Email validation errors
func ErrInvalidEmail() error {
	return result.Validation("user.email.validate", "invalid email format", nil)
}

func ErrEmptyEmail() error {
	return result.Validation("user.email.validate", "email cannot be empty", nil)
}

func ErrEmailTooLong(length, max int) error {
	return result.Validation(
		"user.email.validate",
		fmt.Sprintf("email too long: %d characters (max: %d)", length, max),
		map[string]any{
			"length": length,
			"max":    max,
		},
	)
}

// Username validation errors
func ErrInvalidUsername() error {
	return result.Validation("user.username.validate", "invalid username", nil)
}

func ErrEmptyUsername() error {
	return result.Validation("user.username.validate", "username cannot be empty", nil)
}

func ErrUsernameTooShort(length, min int) error {
	return result.Validation(
		"user.username.validate",
		fmt.Sprintf("username too short: %d characters (min: %d)", length, min),
		map[string]any{
			"length": length,
			"min":    min,
		},
	)
}

func ErrUsernameTooLong(length, max int) error {
	return result.Validation(
		"user.username.validate",
		fmt.Sprintf("username too long: %d characters (max: %d)", length, max),
		map[string]any{
			"length": length,
			"max":    max,
		},
	)
}

func ErrInvalidUsernameChar(username string) error {
	return result.Validation(
		"user.username.validate",
		"username contains invalid characters (only alphanumeric and underscore allowed)",
		map[string]any{
			"username": username,
		},
	)
}

func ErrUsernameStartsWithNumber() error {
	return result.Validation(
		"user.username.validate",
		"username cannot start with a number",
		nil,
	)
}

func ErrReservedUsername(username string) error {
	return result.Validation(
		"user.username.validate",
		fmt.Sprintf("username is reserved: %s", username),
		map[string]any{
			"username": username,
		},
	)
}

// Password validation errors
func ErrInvalidPassword(reason string) error {
	return result.Validation(
		"user.password.validate",
		reason,
		nil,
	)
}

func ErrWeakPassword(reason string) error {
	return result.Validation(
		"user.password.validate",
		fmt.Sprintf("weak password: %s", reason),
		nil,
	)
}

// User ID validation errors
func ErrEmptyUserID() error {
	return result.Validation("user.user.validate", "user ID cannot be empty", nil)
}

// ============================================================================
// Not Found Errors (Kind: KindNotFound)
// ============================================================================

func ErrUserNotFound(identifier string) error {
	return result.NotFound("user.repository.find", "user not found")
}

// ============================================================================
// Conflict Errors (Kind: KindConflict)
// ============================================================================

func ErrEmailAlreadyExists(email string) error {
	return result.Conflict("user.repository.create", "email already exists")
}

func ErrUsernameAlreadyExists(username string) error {
	return result.Conflict("user.repository.create", "username already exists")
}

// ============================================================================
// Domain Rule Violations (Kind: KindDomain)
// ============================================================================

func ErrEmailNotVerified(email string) error {
	return result.Domain("user.verify_email", "email verification required")
}

func ErrUserInactive(id string) error {
	return result.Domain("user.check_active", "user is not active")
}

// ============================================================================
// Authentication Errors (Kind: KindUnauthorized)
// ============================================================================

func ErrPasswordMismatch() error {
	return result.Unauthorized("user.verify_password", "invalid credentials")
}
