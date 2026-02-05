package domain

import (
	"fmt"
)

// CredentialStatus represents the lifecycle status of a credential.
type CredentialStatus int

const (
	// CredentialStatusActive indicates the credential is active and valid.
	CredentialStatusActive CredentialStatus = 1

	// CredentialStatusRevoked indicates the credential has been revoked.
	CredentialStatusRevoked CredentialStatus = 2

	// CredentialStatusExpired indicates the credential has expired.
	CredentialStatusExpired CredentialStatus = 3
)

// String returns the string representation of the credential status.
func (s CredentialStatus) String() string {
	switch s {
	case CredentialStatusActive:
		return "active"
	case CredentialStatusRevoked:
		return "revoked"
	case CredentialStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// IsActive returns true if the credential is active.
func (s CredentialStatus) IsActive() bool {
	return s == CredentialStatusActive
}

// IsRevoked returns true if the credential has been revoked.
func (s CredentialStatus) IsRevoked() bool {
	return s == CredentialStatusRevoked
}

// IsExpired returns true if the credential has expired.
func (s CredentialStatus) IsExpired() bool {
	return s == CredentialStatusExpired
}

// IsTerminal returns true if the credential is in a terminal state.
func (s CredentialStatus) IsTerminal() bool {
	return s == CredentialStatusRevoked || s == CredentialStatusExpired
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s CredentialStatus) CanTransitionTo(target CredentialStatus) bool {
	switch s {
	case CredentialStatusActive:
		return target == CredentialStatusRevoked || target == CredentialStatusExpired
	case CredentialStatusRevoked:
		return false
	case CredentialStatusExpired:
		return false
	default:
		return false
	}
}

// ParseCredentialStatus parses a string into a CredentialStatus.
func ParseCredentialStatus(s string) (CredentialStatus, error) {
	switch s {
	case "active":
		return CredentialStatusActive, nil
	case "revoked":
		return CredentialStatusRevoked, nil
	case "expired":
		return CredentialStatusExpired, nil
	default:
		return 0, fmt.Errorf("invalid credential status: %s", s)
	}
}
