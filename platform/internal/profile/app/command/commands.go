package command

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandCreateProfile         = "profile.CreateProfile"
	CommandUpdateProfile         = "profile.UpdateProfile"
	CommandAddBadge              = "profile.AddBadge"
	CommandRemoveBadge           = "profile.RemoveBadge"
	CommandChangeBadgeVisibility = "profile.ChangeBadgeVisibility"
	CommandClaimVanityURL        = "profile.ClaimVanityURL"
)

// ============================================================================
// CreateProfile
// ============================================================================

// CreateProfile creates a new profile for a user.
type CreateProfile struct {
	UserID      string `json:"user_id" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
	Headline    string `json:"headline" validate:"omitempty"`
	VanitySlug  string `json:"vanity_slug" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c CreateProfile) CommandName() string {
	return CommandCreateProfile
}

// Validate implements cqrs.Validatable.
func (c CreateProfile) Validate() error {
	if c.UserID == "" {
		return cqrs.ErrCommandValidation("CreateProfile.Validate", "user_id is required")
	}
	if c.DisplayName == "" {
		return cqrs.ErrCommandValidation("CreateProfile.Validate", "display_name is required")
	}
	if c.VanitySlug == "" {
		return cqrs.ErrCommandValidation("CreateProfile.Validate", "vanity_slug is required")
	}
	return nil
}

// CreateProfileResult is the result data for CreateProfile.
type CreateProfileResult struct {
	ProfileID  string `json:"profile_id"`
	VanitySlug string `json:"vanity_slug"`
}

// ============================================================================
// UpdateProfile
// ============================================================================

// UpdateProfile updates a profile's metadata.
type UpdateProfile struct {
	ProfileID   types.ID `json:"profile_id" validate:"required"`
	DisplayName string   `json:"display_name" validate:"required"`
	Headline    string   `json:"headline" validate:"omitempty"`
	Bio         string   `json:"bio" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c UpdateProfile) CommandName() string {
	return CommandUpdateProfile
}

// Validate implements cqrs.Validatable.
func (c UpdateProfile) Validate() error {
	if c.ProfileID.IsZero() {
		return cqrs.ErrCommandValidation("UpdateProfile.Validate", "profile_id is required")
	}
	if c.DisplayName == "" {
		return cqrs.ErrCommandValidation("UpdateProfile.Validate", "display_name is required")
	}
	return nil
}

// UpdateProfileResult is the result data for UpdateProfile.
type UpdateProfileResult struct {
	ProfileID string `json:"profile_id"`
}

// ============================================================================
// AddBadge
// ============================================================================

// AddBadge adds a badge to a profile.
type AddBadge struct {
	ProfileID      types.ID  `json:"profile_id" validate:"required"`
	CredentialID   string    `json:"credential_id" validate:"required"`
	BadgeType      string    `json:"badge_type" validate:"required"`
	DisplayName    string    `json:"display_name" validate:"required"`
	PrimaryValue   string    `json:"primary_value" validate:"required"`
	SecondaryValue string    `json:"secondary_value" validate:"omitempty"`
	Icon           string    `json:"icon" validate:"omitempty"`
	VerifiedAt     time.Time `json:"verified_at" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c AddBadge) CommandName() string {
	return CommandAddBadge
}

// Validate implements cqrs.Validatable.
func (c AddBadge) Validate() error {
	if c.ProfileID.IsZero() {
		return cqrs.ErrCommandValidation("AddBadge.Validate", "profile_id is required")
	}
	if c.CredentialID == "" {
		return cqrs.ErrCommandValidation("AddBadge.Validate", "credential_id is required")
	}
	if c.BadgeType == "" {
		return cqrs.ErrCommandValidation("AddBadge.Validate", "badge_type is required")
	}
	if c.DisplayName == "" {
		return cqrs.ErrCommandValidation("AddBadge.Validate", "display_name is required")
	}
	if c.PrimaryValue == "" {
		return cqrs.ErrCommandValidation("AddBadge.Validate", "primary_value is required")
	}
	if c.VerifiedAt.IsZero() {
		return cqrs.ErrCommandValidation("AddBadge.Validate", "verified_at is required")
	}
	return nil
}

// AddBadgeResult is the result data for AddBadge.
type AddBadgeResult struct {
	ProfileID string `json:"profile_id"`
	BadgeID   string `json:"badge_id"`
}

// ============================================================================
// RemoveBadge
// ============================================================================

// RemoveBadge removes a badge from a profile.
type RemoveBadge struct {
	ProfileID types.ID `json:"profile_id" validate:"required"`
	BadgeID   types.ID `json:"badge_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c RemoveBadge) CommandName() string {
	return CommandRemoveBadge
}

// Validate implements cqrs.Validatable.
func (c RemoveBadge) Validate() error {
	if c.ProfileID.IsZero() {
		return cqrs.ErrCommandValidation("RemoveBadge.Validate", "profile_id is required")
	}
	if c.BadgeID.IsZero() {
		return cqrs.ErrCommandValidation("RemoveBadge.Validate", "badge_id is required")
	}
	return nil
}

// RemoveBadgeResult is the result data for RemoveBadge.
type RemoveBadgeResult struct {
	ProfileID string `json:"profile_id"`
	BadgeID   string `json:"badge_id"`
}

// ============================================================================
// ChangeBadgeVisibility
// ============================================================================

// ChangeBadgeVisibility changes a badge's visibility on a profile.
type ChangeBadgeVisibility struct {
	ProfileID  types.ID `json:"profile_id" validate:"required"`
	BadgeID    types.ID `json:"badge_id" validate:"required"`
	Visibility string   `json:"visibility" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ChangeBadgeVisibility) CommandName() string {
	return CommandChangeBadgeVisibility
}

// Validate implements cqrs.Validatable.
func (c ChangeBadgeVisibility) Validate() error {
	if c.ProfileID.IsZero() {
		return cqrs.ErrCommandValidation("ChangeBadgeVisibility.Validate", "profile_id is required")
	}
	if c.BadgeID.IsZero() {
		return cqrs.ErrCommandValidation("ChangeBadgeVisibility.Validate", "badge_id is required")
	}
	if c.Visibility == "" {
		return cqrs.ErrCommandValidation("ChangeBadgeVisibility.Validate", "visibility is required")
	}
	return nil
}

// ChangeBadgeVisibilityResult is the result data for ChangeBadgeVisibility.
type ChangeBadgeVisibilityResult struct {
	ProfileID  string `json:"profile_id"`
	BadgeID    string `json:"badge_id"`
	Visibility string `json:"visibility"`
}

// ============================================================================
// ClaimVanityURL
// ============================================================================

// ClaimVanityURL claims a new vanity URL slug for a profile.
type ClaimVanityURL struct {
	ProfileID types.ID `json:"profile_id" validate:"required"`
	Slug      string   `json:"slug" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ClaimVanityURL) CommandName() string {
	return CommandClaimVanityURL
}

// Validate implements cqrs.Validatable.
func (c ClaimVanityURL) Validate() error {
	if c.ProfileID.IsZero() {
		return cqrs.ErrCommandValidation("ClaimVanityURL.Validate", "profile_id is required")
	}
	if c.Slug == "" {
		return cqrs.ErrCommandValidation("ClaimVanityURL.Validate", "slug is required")
	}
	return nil
}

// ClaimVanityURLResult is the result data for ClaimVanityURL.
type ClaimVanityURLResult struct {
	ProfileID string `json:"profile_id"`
	Slug      string `json:"slug"`
}
