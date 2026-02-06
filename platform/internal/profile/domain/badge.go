package domain

import (
	"time"
)

// Badge is a value object representing a visual representation of a verified
// credential or achievement. Badges are derived from credentials but optimized
// for display on a user's public profile.
type Badge struct {
	id             BadgeID
	credentialID   string
	badgeType      string
	displayName    string
	primaryValue   string
	secondaryValue string
	icon           string
	verifiedAt     time.Time
	visibility     Visibility
}

// NewBadge creates a new Badge value object with required fields.
// Optional fields can be set using With* methods which return new Badge copies.
func NewBadge(
	id BadgeID,
	credentialID string,
	badgeType string,
	displayName string,
	primaryValue string,
	verifiedAt time.Time,
) Badge {
	return Badge{
		id:           id,
		credentialID: credentialID,
		badgeType:    badgeType,
		displayName:  displayName,
		primaryValue: primaryValue,
		verifiedAt:   verifiedAt,
		visibility:   VisibilityPublic,
	}
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the badge's unique identifier.
func (b Badge) ID() BadgeID {
	return b.id
}

// CredentialID returns the ID of the credential this badge was derived from.
func (b Badge) CredentialID() string {
	return b.credentialID
}

// BadgeType returns the type of badge (e.g., "GitHub", "LinkedIn", "Certification").
func (b Badge) BadgeType() string {
	return b.badgeType
}

// DisplayName returns the display name of the badge.
func (b Badge) DisplayName() string {
	return b.displayName
}

// PrimaryValue returns the primary display value (e.g., "1337 commits").
func (b Badge) PrimaryValue() string {
	return b.primaryValue
}

// SecondaryValue returns the optional secondary display value (e.g., "42 repositories").
func (b Badge) SecondaryValue() string {
	return b.secondaryValue
}

// Icon returns the icon identifier or URL for the badge.
func (b Badge) Icon() string {
	return b.icon
}

// VerifiedAt returns when the underlying credential was verified.
func (b Badge) VerifiedAt() time.Time {
	return b.verifiedAt
}

// Visibility returns the badge's visibility setting.
func (b Badge) Visibility() Visibility {
	return b.visibility
}

// ============================================================================
// Immutable Mutation Methods (return new copies)
// ============================================================================

// WithSecondaryValue returns a new Badge with the secondary value set.
func (b Badge) WithSecondaryValue(secondaryValue string) Badge {
	b.secondaryValue = secondaryValue
	return b
}

// WithIcon returns a new Badge with the icon set.
func (b Badge) WithIcon(icon string) Badge {
	b.icon = icon
	return b
}

// WithVisibility returns a new Badge with the visibility set.
func (b Badge) WithVisibility(visibility Visibility) Badge {
	b.visibility = visibility
	return b
}
