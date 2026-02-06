// Package domain contains the core business logic for the Profile bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Profile-specific)
// ============================================================================

const (
	CodeProfileNotFound      pkgerrors.Code = "PROFILE_NOT_FOUND"
	CodeProfileAlreadyExists pkgerrors.Code = "PROFILE_ALREADY_EXISTS"
	CodeProfileInvalid       pkgerrors.Code = "PROFILE_INVALID"
	CodeBadgeNotFound        pkgerrors.Code = "BADGE_NOT_FOUND"
	CodeVanityURLTaken       pkgerrors.Code = "VANITY_URL_TAKEN"
	CodeVanityURLInvalid     pkgerrors.Code = "VANITY_URL_INVALID"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrProfileNotFound      = errors.New("profile not found")
	ErrProfileAlreadyExists = errors.New("profile already exists")
	ErrProfileInvalid       = errors.New("profile is invalid")
	ErrBadgeNotFound        = errors.New("badge not found")
	ErrVanityURLTaken       = errors.New("vanity URL is already taken")
	ErrVanityURLInvalid     = errors.New("vanity URL is invalid")
)

// ============================================================================
// Error Constructors
// ============================================================================

// ProfileNotFound creates a profile not found error.
func ProfileNotFound(operation string, profileID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "profile").
		WithCode(CodeProfileNotFound).
		WithMeta("profile_id", profileID)
}

// ProfileAlreadyExists creates a profile already exists error.
func ProfileAlreadyExists(operation string, userID string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "profile").
		WithCode(CodeProfileAlreadyExists).
		WithMeta("user_id", userID)
}

// ProfileInvalid creates a profile invalid error.
func ProfileInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "profile is invalid: "+reason).
		WithCode(CodeProfileInvalid)
}

// BadgeNotFound creates a badge not found error.
func BadgeNotFound(operation string, badgeID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "badge").
		WithCode(CodeBadgeNotFound).
		WithMeta("badge_id", badgeID)
}

// VanityURLTaken creates a vanity URL taken error.
func VanityURLTaken(operation string, slug string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "vanity URL").
		WithCode(CodeVanityURLTaken).
		WithMeta("slug", slug)
}

// VanityURLInvalid creates a vanity URL invalid error.
func VanityURLInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "vanity URL is invalid: "+reason).
		WithCode(CodeVanityURLInvalid)
}
