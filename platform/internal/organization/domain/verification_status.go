package domain

import "fmt"

// VerificationStatus represents the KYB verification state of an organization.
type VerificationStatus int

const (
	VerificationStatusUnverified VerificationStatus = 1
	VerificationStatusPending    VerificationStatus = 2
	VerificationStatusVerified   VerificationStatus = 3
)

// String returns the string representation.
func (s VerificationStatus) String() string {
	switch s {
	case VerificationStatusUnverified:
		return "unverified"
	case VerificationStatusPending:
		return "pending"
	case VerificationStatusVerified:
		return "verified"
	default:
		return "unknown"
	}
}

// ParseVerificationStatus parses a string into a VerificationStatus.
func ParseVerificationStatus(s string) (VerificationStatus, error) {
	switch s {
	case "unverified":
		return VerificationStatusUnverified, nil
	case "pending":
		return VerificationStatusPending, nil
	case "verified":
		return VerificationStatusVerified, nil
	default:
		return 0, fmt.Errorf("invalid verification status: %s", s)
	}
}

// IsVerified returns true if the organization is verified.
func (s VerificationStatus) IsVerified() bool { return s == VerificationStatusVerified }

// IsPending returns true if verification is in progress.
func (s VerificationStatus) IsPending() bool { return s == VerificationStatusPending }

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s VerificationStatus) CanTransitionTo(target VerificationStatus) bool {
	switch s {
	case VerificationStatusUnverified:
		return target == VerificationStatusPending
	case VerificationStatusPending:
		return target == VerificationStatusVerified || target == VerificationStatusUnverified
	case VerificationStatusVerified:
		return false
	default:
		return false
	}
}
