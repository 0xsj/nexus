package domain

import (
	"fmt"
)

// CredentialType represents the type of a verifiable credential.
type CredentialType string

const (
	// CredentialTypeGitHubContributor represents a GitHub contributor credential.
	CredentialTypeGitHubContributor CredentialType = "GitHubContributor"

	// CredentialTypeProfessionalExperience represents a professional experience credential.
	CredentialTypeProfessionalExperience CredentialType = "ProfessionalExperience"

	// CredentialTypeCourseCompletion represents a course completion credential.
	CredentialTypeCourseCompletion CredentialType = "CourseCompletion"

	// CredentialTypeCustom represents a custom credential type.
	CredentialTypeCustom CredentialType = "Custom"
)

// ParseCredentialType parses a string into a CredentialType.
func ParseCredentialType(s string) (CredentialType, error) {
	switch s {
	case string(CredentialTypeGitHubContributor):
		return CredentialTypeGitHubContributor, nil
	case string(CredentialTypeProfessionalExperience):
		return CredentialTypeProfessionalExperience, nil
	case string(CredentialTypeCourseCompletion):
		return CredentialTypeCourseCompletion, nil
	case string(CredentialTypeCustom):
		return CredentialTypeCustom, nil
	default:
		return "", fmt.Errorf("unknown credential type: %s", s)
	}
}

// String returns the string representation of the CredentialType.
func (t CredentialType) String() string {
	return string(t)
}

// IsValid returns true if the CredentialType is a known, valid type.
func (t CredentialType) IsValid() bool {
	switch t {
	case CredentialTypeGitHubContributor,
		CredentialTypeProfessionalExperience,
		CredentialTypeCourseCompletion,
		CredentialTypeCustom:
		return true
	default:
		return false
	}
}

// IsBuiltIn returns true if the CredentialType is a built-in (non-custom) type.
func (t CredentialType) IsBuiltIn() bool {
	return t.IsValid() && t != CredentialTypeCustom
}
