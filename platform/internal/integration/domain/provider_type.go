package domain

import (
	"fmt"
)

// ProviderType represents the type of external provider.
type ProviderType string

const (
	// ProviderTypeGitHub represents the GitHub provider.
	ProviderTypeGitHub ProviderType = "github"

	// ProviderTypeLinkedIn represents the LinkedIn provider.
	ProviderTypeLinkedIn ProviderType = "linkedin"

	// ProviderTypeTwitter represents the Twitter/X provider.
	ProviderTypeTwitter ProviderType = "twitter"

	// ProviderTypeCoursera represents the Coursera provider.
	ProviderTypeCoursera ProviderType = "coursera"

	// ProviderTypeAWS represents the AWS provider.
	ProviderTypeAWS ProviderType = "aws"

	// ProviderTypeGoogle represents the Google provider.
	ProviderTypeGoogle ProviderType = "google"
)

// ParseProviderType parses a string into a ProviderType.
func ParseProviderType(s string) (ProviderType, error) {
	switch s {
	case string(ProviderTypeGitHub):
		return ProviderTypeGitHub, nil
	case string(ProviderTypeLinkedIn):
		return ProviderTypeLinkedIn, nil
	case string(ProviderTypeTwitter):
		return ProviderTypeTwitter, nil
	case string(ProviderTypeCoursera):
		return ProviderTypeCoursera, nil
	case string(ProviderTypeAWS):
		return ProviderTypeAWS, nil
	case string(ProviderTypeGoogle):
		return ProviderTypeGoogle, nil
	default:
		return "", fmt.Errorf("unknown provider type: %s", s)
	}
}

// String returns the string representation of the ProviderType.
func (t ProviderType) String() string {
	return string(t)
}

// IsValid returns true if the ProviderType is a known, valid type.
func (t ProviderType) IsValid() bool {
	switch t {
	case ProviderTypeGitHub,
		ProviderTypeLinkedIn,
		ProviderTypeTwitter,
		ProviderTypeCoursera,
		ProviderTypeAWS,
		ProviderTypeGoogle:
		return true
	default:
		return false
	}
}

// AllProviderTypes returns all valid provider types.
func AllProviderTypes() []ProviderType {
	return []ProviderType{
		ProviderTypeGitHub,
		ProviderTypeLinkedIn,
		ProviderTypeTwitter,
		ProviderTypeCoursera,
		ProviderTypeAWS,
		ProviderTypeGoogle,
	}
}
