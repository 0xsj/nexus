package domain

import (
	"fmt"
)

// OAuthProvider represents a supported OAuth authentication provider.
type OAuthProvider int

const (
	// OAuthProviderUnknown indicates an unknown provider.
	OAuthProviderUnknown OAuthProvider = iota

	// OAuthProviderGoogle indicates Google OAuth.
	OAuthProviderGoogle

	// OAuthProviderGitHub indicates GitHub OAuth.
	OAuthProviderGitHub
)

// String returns the string representation of the OAuth provider.
func (p OAuthProvider) String() string {
	switch p {
	case OAuthProviderGoogle:
		return "google"
	case OAuthProviderGitHub:
		return "github"
	default:
		return "unknown"
	}
}

// ParseOAuthProvider parses a string into an OAuthProvider.
func ParseOAuthProvider(s string) (OAuthProvider, error) {
	switch s {
	case "google":
		return OAuthProviderGoogle, nil
	case "github":
		return OAuthProviderGitHub, nil
	case "unknown":
		return OAuthProviderUnknown, nil
	default:
		return OAuthProviderUnknown, fmt.Errorf("invalid oauth provider: %s", s)
	}
}

// MustParseOAuthProvider parses a string into an OAuthProvider and panics if invalid.
// Only use for constants or tests.
func MustParseOAuthProvider(s string) OAuthProvider {
	p, err := ParseOAuthProvider(s)
	if err != nil {
		panic(err)
	}
	return p
}

// IsValid returns true if the provider is a known, valid provider.
func (p OAuthProvider) IsValid() bool {
	switch p {
	case OAuthProviderGoogle, OAuthProviderGitHub:
		return true
	default:
		return false
	}
}

// IsZero returns true if the provider is the zero value (unknown).
func (p OAuthProvider) IsZero() bool {
	return p == OAuthProviderUnknown
}

// ProvidesEmail returns true if the provider typically provides an email address.
func (p OAuthProvider) ProvidesEmail() bool {
	switch p {
	case OAuthProviderGoogle, OAuthProviderGitHub:
		return true
	default:
		return false
	}
}

// DisplayName returns a human-readable name for the provider.
func (p OAuthProvider) DisplayName() string {
	switch p {
	case OAuthProviderGoogle:
		return "Google"
	case OAuthProviderGitHub:
		return "GitHub"
	default:
		return "Unknown"
	}
}
