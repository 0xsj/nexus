package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetProfile             = "profile.GetProfile"
	QueryGetProfileByUser       = "profile.GetProfileByUser"
	QueryGetProfileByVanitySlug = "profile.GetProfileByVanitySlug"
	QueryListProfiles           = "profile.ListProfiles"
)

// ============================================================================
// GetProfile
// ============================================================================

// GetProfile retrieves a single profile by ID.
type GetProfile struct {
	ProfileID types.ID `json:"profile_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetProfile) QueryName() string {
	return QueryGetProfile
}

// Validate implements cqrs.Validatable.
func (q GetProfile) Validate() error {
	if q.ProfileID.IsZero() {
		return cqrs.ErrQueryValidation("GetProfile.Validate", "profile_id is required")
	}
	return nil
}

// ============================================================================
// GetProfileByUser
// ============================================================================

// GetProfileByUser retrieves a profile by user ID.
type GetProfileByUser struct {
	UserID string `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetProfileByUser) QueryName() string {
	return QueryGetProfileByUser
}

// Validate implements cqrs.Validatable.
func (q GetProfileByUser) Validate() error {
	if q.UserID == "" {
		return cqrs.ErrQueryValidation("GetProfileByUser.Validate", "user_id is required")
	}
	return nil
}

// ============================================================================
// GetProfileByVanitySlug
// ============================================================================

// GetProfileByVanitySlug retrieves a profile by vanity URL slug.
type GetProfileByVanitySlug struct {
	VanitySlug string `json:"vanity_slug" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetProfileByVanitySlug) QueryName() string {
	return QueryGetProfileByVanitySlug
}

// Validate implements cqrs.Validatable.
func (q GetProfileByVanitySlug) Validate() error {
	if q.VanitySlug == "" {
		return cqrs.ErrQueryValidation("GetProfileByVanitySlug.Validate", "vanity_slug is required")
	}
	return nil
}

// ============================================================================
// ListProfiles
// ============================================================================

// ListProfiles lists profiles with pagination.
type ListProfiles struct {
	Limit  int `json:"limit" validate:"omitempty"`
	Offset int `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListProfiles) QueryName() string {
	return QueryListProfiles
}

// Validate implements cqrs.Validatable.
func (q ListProfiles) Validate() error {
	return nil
}
