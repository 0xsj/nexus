package domain

import (
	"fmt"
)

// Category represents a notification category for preference management.
type Category int

const (
	// CategorySecurity covers login alerts, API key events, session revocations.
	CategorySecurity Category = 1

	// CategoryCredentials covers credential issuance, verification, expiration, revocation.
	CategoryCredentials Category = 2

	// CategoryVerification covers verification started, completed, failed.
	CategoryVerification Category = 3

	// CategorySocial covers vouches, profile views, social interactions.
	CategorySocial Category = 4

	// CategoryOrganization covers invitations, role changes, member events.
	CategoryOrganization Category = 5

	// CategoryIssuer covers credential requests, batch completions.
	CategoryIssuer Category = 6

	// CategorySystem covers maintenance, policy updates, new features.
	CategorySystem Category = 7
)

// String returns the string representation of the category.
func (c Category) String() string {
	switch c {
	case CategorySecurity:
		return "security"
	case CategoryCredentials:
		return "credentials"
	case CategoryVerification:
		return "verification"
	case CategorySocial:
		return "social"
	case CategoryOrganization:
		return "organization"
	case CategoryIssuer:
		return "issuer"
	case CategorySystem:
		return "system"
	default:
		return "unknown"
	}
}

// ParseCategory parses a string into a Category.
func ParseCategory(s string) (Category, error) {
	switch s {
	case "security":
		return CategorySecurity, nil
	case "credentials":
		return CategoryCredentials, nil
	case "verification":
		return CategoryVerification, nil
	case "social":
		return CategorySocial, nil
	case "organization":
		return CategoryOrganization, nil
	case "issuer":
		return CategoryIssuer, nil
	case "system":
		return CategorySystem, nil
	default:
		return 0, fmt.Errorf("invalid category: %s", s)
	}
}

// IsValid returns true if the Category is a known, valid category.
func (c Category) IsValid() bool {
	switch c {
	case CategorySecurity, CategoryCredentials, CategoryVerification,
		CategorySocial, CategoryOrganization, CategoryIssuer, CategorySystem:
		return true
	default:
		return false
	}
}
