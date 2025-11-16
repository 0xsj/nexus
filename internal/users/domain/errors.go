// Package domain contains the core business logic for the users domain.
package domain

import (
	"fmt"

	"github.com/0xsj/result"
)

// ============================================================================
// Email Validation Errors (Kind: KindValidation)
// ============================================================================

// ErrInvalidEmail is returned when an email address is invalid.
type ErrInvalidEmail struct {
	Email  string
	Reason string
}

func (e ErrInvalidEmail) Error() string {
	return fmt.Sprintf("invalid email '%s': %s", e.Email, e.Reason)
}

func (e ErrInvalidEmail) Unwrap() error {
	return result.Validation("user.email.validate", e.Reason, map[string]any{
		"email":  e.Email,
		"reason": e.Reason,
	})
}

// ============================================================================
// Username Validation Errors (Kind: KindValidation)
// ============================================================================

// ErrInvalidUsername is returned when a username is invalid.
type ErrInvalidUsername struct {
	Username string
	Reason   string
}

func (e ErrInvalidUsername) Error() string {
	return fmt.Sprintf("invalid username '%s': %s", e.Username, e.Reason)
}

func (e ErrInvalidUsername) Unwrap() error {
	return result.Validation("user.username.validate", e.Reason, map[string]any{
		"username": e.Username,
		"reason":   e.Reason,
	})
}

// ============================================================================
// Password Validation Errors (Kind: KindValidation)
// ============================================================================

// ErrInvalidPassword is returned when a password is invalid.
type ErrInvalidPassword struct {
	Reason string
}

func (e ErrInvalidPassword) Error() string {
	return fmt.Sprintf("invalid password: %s", e.Reason)
}

func (e ErrInvalidPassword) Unwrap() error {
	return result.Validation("user.password.validate", e.Reason, nil)
}

// ErrWeakPassword is returned when a password doesn't meet strength requirements.
type ErrWeakPassword struct {
	Reason string
}

func (e ErrWeakPassword) Error() string {
	return fmt.Sprintf("weak password: %s", e.Reason)
}

func (e ErrWeakPassword) Unwrap() error {
	return result.Validation("user.password.validate", e.Reason, nil)
}

// ============================================================================
// User ID Validation Errors (Kind: KindValidation)
// ============================================================================

// ErrInvalidUserID is returned when a user ID is invalid.
type ErrInvalidUserID struct {
	ID     string
	Reason string
}

func (e ErrInvalidUserID) Error() string {
	return fmt.Sprintf("invalid user id '%s': %s", e.ID, e.Reason)
}

func (e ErrInvalidUserID) Unwrap() error {
	return result.Validation("user.id.validate", e.Reason, map[string]any{
		"id":     e.ID,
		"reason": e.Reason,
	})
}

// ============================================================================
// Not Found Errors (Kind: KindNotFound)
// ============================================================================

// ErrUserNotFound is returned when a user cannot be found.
type ErrUserNotFound struct {
	ID string
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user not found: %s", e.ID)
}

func (e ErrUserNotFound) Unwrap() error {
	return result.NotFound("user.repository.find", "user")
}

// ============================================================================
// Conflict Errors (Kind: KindConflict)
// ============================================================================

// ErrEmailAlreadyExists is returned when attempting to create a user with an email that already exists.
type ErrEmailAlreadyExists struct {
	Email string
}

func (e ErrEmailAlreadyExists) Error() string {
	return fmt.Sprintf("email already exists: %s", e.Email)
}

func (e ErrEmailAlreadyExists) Unwrap() error {
	return result.Conflict("user.repository.create", "email")
}

// ErrUsernameAlreadyExists is returned when attempting to create a user with a username that already exists.
type ErrUsernameAlreadyExists struct {
	Username string
}

func (e ErrUsernameAlreadyExists) Error() string {
	return fmt.Sprintf("username already exists: %s", e.Username)
}

func (e ErrUsernameAlreadyExists) Unwrap() error {
	return result.Conflict("user.repository.create", "username")
}

// ============================================================================
// Domain Rule Violations (Kind: KindDomain)
// ============================================================================

// ErrEmailNotVerified is returned when an operation requires email verification.
type ErrEmailNotVerified struct {
	Email string
}

func (e ErrEmailNotVerified) Error() string {
	return fmt.Sprintf("email not verified: %s", e.Email)
}

func (e ErrEmailNotVerified) Unwrap() error {
	return result.Domain("user.verify_email", "email verification required")
}

// ErrUserInactive is returned when attempting to perform operations on an inactive user.
type ErrUserInactive struct {
	ID string
}

func (e ErrUserInactive) Error() string {
	return fmt.Sprintf("user is inactive: %s", e.ID)
}

func (e ErrUserInactive) Unwrap() error {
	return result.Domain("user.check_active", "user is not active")
}

// ErrUserDeleted is returned when attempting to perform operations on a deleted user.
type ErrUserDeleted struct {
	ID string
}

func (e ErrUserDeleted) Error() string {
	return fmt.Sprintf("user is deleted: %s", e.ID)
}

func (e ErrUserDeleted) Unwrap() error {
	return result.Domain("user.check_deleted", "user has been deleted")
}

// ============================================================================
// Authentication Errors (Kind: KindUnauthorized)
// ============================================================================

// ErrPasswordMismatch is returned when password verification fails.
type ErrPasswordMismatch struct{}

func (e ErrPasswordMismatch) Error() string {
	return "password does not match"
}

func (e ErrPasswordMismatch) Unwrap() error {
	return result.Unauthorized("user.verify_password", "invalid credentials")
}
