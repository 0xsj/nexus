package domain

import (
	"fmt"
)

// AuthMethod represents the method used for authentication.
type AuthMethod int

const (
	// AuthMethodUnknown indicates an unknown authentication method.
	AuthMethodUnknown AuthMethod = iota

	// AuthMethodMagicLink indicates email magic link authentication.
	AuthMethodMagicLink

	// AuthMethodWallet indicates wallet-based authentication (SIWE).
	AuthMethodWallet

	// AuthMethodOAuth indicates OAuth provider authentication.
	AuthMethodOAuth
)

// String returns the string representation of the auth method.
func (m AuthMethod) String() string {
	switch m {
	case AuthMethodMagicLink:
		return "magic_link"
	case AuthMethodWallet:
		return "wallet"
	case AuthMethodOAuth:
		return "oauth"
	default:
		return "unknown"
	}
}

// ParseAuthMethod parses a string into an AuthMethod.
func ParseAuthMethod(s string) (AuthMethod, error) {
	switch s {
	case "magic_link":
		return AuthMethodMagicLink, nil
	case "wallet":
		return AuthMethodWallet, nil
	case "oauth":
		return AuthMethodOAuth, nil
	case "unknown":
		return AuthMethodUnknown, nil
	default:
		return AuthMethodUnknown, fmt.Errorf("invalid auth method: %s", s)
	}
}

// MustParseAuthMethod parses a string into an AuthMethod and panics if invalid.
// Only use for constants or tests.
func MustParseAuthMethod(s string) AuthMethod {
	m, err := ParseAuthMethod(s)
	if err != nil {
		panic(err)
	}
	return m
}

// IsValid returns true if the auth method is a known, valid method.
func (m AuthMethod) IsValid() bool {
	switch m {
	case AuthMethodMagicLink, AuthMethodWallet, AuthMethodOAuth:
		return true
	default:
		return false
	}
}

// IsZero returns true if the auth method is the zero value (unknown).
func (m AuthMethod) IsZero() bool {
	return m == AuthMethodUnknown
}

// RequiresEmail returns true if the auth method requires an email address.
func (m AuthMethod) RequiresEmail() bool {
	switch m {
	case AuthMethodMagicLink, AuthMethodOAuth:
		return true
	default:
		return false
	}
}

// RequiresWallet returns true if the auth method requires a wallet address.
func (m AuthMethod) RequiresWallet() bool {
	return m == AuthMethodWallet
}
