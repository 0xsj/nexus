package domain

import (
	"fmt"
)

// ProviderType represents a supported external verification provider.
type ProviderType string

const (
	// ProviderTypeGitHub indicates GitHub as the verification provider.
	ProviderTypeGitHub ProviderType = "github"

	// ProviderTypeLinkedIn indicates LinkedIn as the verification provider.
	ProviderTypeLinkedIn ProviderType = "linkedin"

	// ProviderTypeCoursera indicates Coursera as the verification provider.
	ProviderTypeCoursera ProviderType = "coursera"

	// ProviderTypeGoogle indicates Google as the verification provider.
	ProviderTypeGoogle ProviderType = "google"

	// ProviderTypeTwitter indicates Twitter as the verification provider.
	ProviderTypeTwitter ProviderType = "twitter"

	// ProviderTypeAWS indicates AWS as the verification provider.
	ProviderTypeAWS ProviderType = "aws"
)

// ParseProviderType parses a string into a ProviderType.
func ParseProviderType(s string) (ProviderType, error) {
	switch s {
	case "github":
		return ProviderTypeGitHub, nil
	case "linkedin":
		return ProviderTypeLinkedIn, nil
	case "coursera":
		return ProviderTypeCoursera, nil
	case "google":
		return ProviderTypeGoogle, nil
	case "twitter":
		return ProviderTypeTwitter, nil
	case "aws":
		return ProviderTypeAWS, nil
	default:
		return "", fmt.Errorf("invalid provider type: %s", s)
	}
}

// String returns the string representation of the ProviderType.
func (p ProviderType) String() string {
	return string(p)
}

// IsValid returns true if the ProviderType is a known, valid type.
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderTypeGitHub, ProviderTypeLinkedIn, ProviderTypeCoursera, ProviderTypeGoogle, ProviderTypeTwitter, ProviderTypeAWS:
		return true
	default:
		return false
	}
}
