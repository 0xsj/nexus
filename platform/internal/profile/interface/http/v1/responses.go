package v1

import "time"

// ============================================================================
// Profile Responses
// ============================================================================

// ProfileResponse represents a full profile in API responses.
type ProfileResponse struct {
	ProfileID   string          `json:"profile_id"`
	UserID      string          `json:"user_id"`
	DisplayName string          `json:"display_name"`
	Headline    string          `json:"headline"`
	Bio         string          `json:"bio"`
	VanitySlug  string          `json:"vanity_slug"`
	Badges      []BadgeResponse `json:"badges"`
	BadgeCount  int             `json:"badge_count"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ProfileSummaryResponse represents a lightweight profile in list responses.
type ProfileSummaryResponse struct {
	ProfileID   string    `json:"profile_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Headline    string    `json:"headline"`
	VanitySlug  string    `json:"vanity_slug"`
	BadgeCount  int       `json:"badge_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProfileListResponse represents a paginated list of profiles.
type ProfileListResponse struct {
	Profiles   []ProfileSummaryResponse `json:"profiles"`
	TotalCount int                      `json:"total_count"`
	Limit      int                      `json:"limit"`
	Offset     int                      `json:"offset"`
	HasMore    bool                     `json:"has_more"`
}

// ============================================================================
// Badge Responses
// ============================================================================

// BadgeResponse represents a badge in API responses.
type BadgeResponse struct {
	BadgeID        string    `json:"badge_id"`
	CredentialID   string    `json:"credential_id"`
	BadgeType      string    `json:"badge_type"`
	DisplayName    string    `json:"display_name"`
	PrimaryValue   string    `json:"primary_value"`
	SecondaryValue string    `json:"secondary_value,omitempty"`
	Icon           string    `json:"icon,omitempty"`
	VerifiedAt     time.Time `json:"verified_at"`
	Visibility     string    `json:"visibility"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// ProfileCreatedResponse represents the result of creating a profile.
type ProfileCreatedResponse struct {
	ProfileID  string    `json:"profile_id"`
	VanitySlug string    `json:"vanity_slug"`
	CreatedAt  time.Time `json:"created_at"`
}

// ProfileUpdatedResponse represents the result of updating a profile.
type ProfileUpdatedResponse struct {
	ProfileID string    `json:"profile_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BadgeAddedResponse represents the result of adding a badge.
type BadgeAddedResponse struct {
	ProfileID string    `json:"profile_id"`
	BadgeID   string    `json:"badge_id"`
	AddedAt   time.Time `json:"added_at"`
}

// BadgeRemovedResponse represents the result of removing a badge.
type BadgeRemovedResponse struct {
	ProfileID string    `json:"profile_id"`
	BadgeID   string    `json:"badge_id"`
	RemovedAt time.Time `json:"removed_at"`
}

// BadgeVisibilityChangedResponse represents the result of changing badge visibility.
type BadgeVisibilityChangedResponse struct {
	ProfileID  string    `json:"profile_id"`
	BadgeID    string    `json:"badge_id"`
	Visibility string    `json:"visibility"`
	ChangedAt  time.Time `json:"changed_at"`
}

// VanityURLClaimedResponse represents the result of claiming a vanity URL.
type VanityURLClaimedResponse struct {
	ProfileID string    `json:"profile_id"`
	Slug      string    `json:"slug"`
	ClaimedAt time.Time `json:"claimed_at"`
}
