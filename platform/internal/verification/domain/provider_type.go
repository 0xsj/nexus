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
	case ProviderTypeGitHub, ProviderTypeLinkedIn, ProviderTypeCoursera:
		return true
	default:
		return false
	}
}
