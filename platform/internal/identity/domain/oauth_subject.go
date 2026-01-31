package domain

import (
	"fmt"
	"strings"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// OAuthSubject represents a unique identifier for a user within an OAuth provider.
// Combines the provider and the provider's unique subject ID.
type OAuthSubject struct {
	provider   OAuthProvider
	externalID string
}

// NewOAuthSubject creates a new OAuthSubject.
func NewOAuthSubject(provider OAuthProvider, externalID string) (OAuthSubject, error) {
	if !provider.IsValid() {
		return OAuthSubject{}, pkgerrors.Validation("OAuthSubject.New", "invalid oauth provider")
	}

	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		return OAuthSubject{}, pkgerrors.Validation("OAuthSubject.New", "external ID cannot be empty")
	}

	return OAuthSubject{
		provider:   provider,
		externalID: externalID,
	}, nil
}

// MustNewOAuthSubject creates a new OAuthSubject and panics if invalid.
// Only use for constants or tests.
func MustNewOAuthSubject(provider OAuthProvider, externalID string) OAuthSubject {
	s, err := NewOAuthSubject(provider, externalID)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseOAuthSubject parses a string in the format "provider:external_id" into an OAuthSubject.
func ParseOAuthSubject(s string) (OAuthSubject, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return OAuthSubject{}, pkgerrors.Validation("OAuthSubject.Parse", "oauth subject cannot be empty")
	}

	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return OAuthSubject{}, pkgerrors.Validation("OAuthSubject.Parse", "invalid oauth subject format, expected 'provider:external_id'").
			WithMeta("value", s)
	}

	provider, err := ParseOAuthProvider(parts[0])
	if err != nil {
		return OAuthSubject{}, pkgerrors.Validation("OAuthSubject.Parse", "invalid oauth provider").
			WithMeta("provider", parts[0])
	}

	return NewOAuthSubject(provider, parts[1])
}

// MustParseOAuthSubject parses a string into an OAuthSubject and panics if invalid.
// Only use for constants or tests.
func MustParseOAuthSubject(s string) OAuthSubject {
	subject, err := ParseOAuthSubject(s)
	if err != nil {
		panic(err)
	}
	return subject
}

// String returns the string representation in the format "provider:external_id".
func (s OAuthSubject) String() string {
	if s.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s:%s", s.provider.String(), s.externalID)
}

// Provider returns the OAuth provider.
func (s OAuthSubject) Provider() OAuthProvider {
	return s.provider
}

// ExternalID returns the external ID from the provider.
func (s OAuthSubject) ExternalID() string {
	return s.externalID
}

// IsZero returns true if the OAuthSubject is the zero value.
func (s OAuthSubject) IsZero() bool {
	return s.provider.IsZero() && s.externalID == ""
}

// IsValid returns true if the OAuthSubject is valid.
func (s OAuthSubject) IsValid() bool {
	return s.provider.IsValid() && s.externalID != ""
}

// Equals checks if two OAuthSubjects are equal.
func (s OAuthSubject) Equals(other OAuthSubject) bool {
	return s.provider == other.provider && s.externalID == other.externalID
}
