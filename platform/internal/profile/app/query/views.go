package query

import "time"

// ============================================================================
// Profile Views
// ============================================================================

// ProfileView is the full profile read model.
type ProfileView struct {
	ProfileID   string      `json:"profile_id"`
	UserID      string      `json:"user_id"`
	DisplayName string      `json:"display_name"`
	Headline    string      `json:"headline"`
	Bio         string      `json:"bio"`
	VanitySlug  string      `json:"vanity_slug"`
	Badges      []BadgeView `json:"badges"`
	BadgeCount  int         `json:"badge_count"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// ProfileSummaryView is a lightweight profile representation for lists.
type ProfileSummaryView struct {
	ProfileID   string    `json:"profile_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Headline    string    `json:"headline"`
	VanitySlug  string    `json:"vanity_slug"`
	BadgeCount  int       `json:"badge_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProfileListView is a paginated list of profiles.
type ProfileListView struct {
	Profiles   []ProfileSummaryView `json:"profiles"`
	TotalCount int                  `json:"total_count"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	HasMore    bool                 `json:"has_more"`
}

// ============================================================================
// Badge Views
// ============================================================================

// BadgeView is a badge read model.
type BadgeView struct {
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
