package v1

import "time"

// ============================================================================
// Profile Requests
// ============================================================================

// CreateProfileRequest represents a request to create a new profile.
type CreateProfileRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
	Headline    string `json:"headline,omitempty"`
	VanitySlug  string `json:"vanity_slug" validate:"required"`
}

// UpdateProfileRequest represents a request to update profile metadata.
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" validate:"required"`
	Headline    string `json:"headline,omitempty"`
	Bio         string `json:"bio,omitempty"`
}

// AddBadgeRequest represents a request to add a badge to a profile.
type AddBadgeRequest struct {
	CredentialID   string    `json:"credential_id" validate:"required"`
	BadgeType      string    `json:"badge_type" validate:"required"`
	DisplayName    string    `json:"display_name" validate:"required"`
	PrimaryValue   string    `json:"primary_value" validate:"required"`
	SecondaryValue string    `json:"secondary_value,omitempty"`
	Icon           string    `json:"icon,omitempty"`
	VerifiedAt     time.Time `json:"verified_at" validate:"required"`
}

// ChangeBadgeVisibilityRequest represents a request to change a badge's visibility.
type ChangeBadgeVisibilityRequest struct {
	Visibility string `json:"visibility" validate:"required"`
}

// ClaimVanityURLRequest represents a request to claim a vanity URL slug.
type ClaimVanityURLRequest struct {
	Slug string `json:"slug" validate:"required"`
}
