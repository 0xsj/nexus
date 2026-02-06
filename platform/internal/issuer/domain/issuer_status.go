package domain

import (
	"fmt"
)

// IssuerStatus represents the lifecycle status of an issuer.
type IssuerStatus int

const (
	// IssuerStatusPending indicates the issuer is pending activation.
	IssuerStatusPending IssuerStatus = 1

	// IssuerStatusActive indicates the issuer is active and can issue credentials.
	IssuerStatusActive IssuerStatus = 2

	// IssuerStatusSuspended indicates the issuer has been suspended.
	IssuerStatusSuspended IssuerStatus = 3

	// IssuerStatusRevoked indicates the issuer has been permanently revoked.
	IssuerStatusRevoked IssuerStatus = 4
)

// String returns the string representation of the issuer status.
func (s IssuerStatus) String() string {
	switch s {
	case IssuerStatusPending:
		return "pending"
	case IssuerStatusActive:
		return "active"
	case IssuerStatusSuspended:
		return "suspended"
	case IssuerStatusRevoked:
		return "revoked"
	default:
		return "unknown"
	}
}

// IsPending returns true if the issuer is pending.
func (s IssuerStatus) IsPending() bool {
	return s == IssuerStatusPending
}

// IsActive returns true if the issuer is active.
func (s IssuerStatus) IsActive() bool {
	return s == IssuerStatusActive
}

// IsSuspended returns true if the issuer is suspended.
func (s IssuerStatus) IsSuspended() bool {
	return s == IssuerStatusSuspended
}

// IsRevoked returns true if the issuer is revoked.
func (s IssuerStatus) IsRevoked() bool {
	return s == IssuerStatusRevoked
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s IssuerStatus) CanTransitionTo(target IssuerStatus) bool {
	switch s {
	case IssuerStatusPending:
		return target == IssuerStatusActive || target == IssuerStatusRevoked
	case IssuerStatusActive:
		return target == IssuerStatusSuspended || target == IssuerStatusRevoked
	case IssuerStatusSuspended:
		return target == IssuerStatusActive || target == IssuerStatusRevoked
	case IssuerStatusRevoked:
		return false
	default:
		return false
	}
}

// ParseIssuerStatus parses a string into an IssuerStatus.
func ParseIssuerStatus(s string) (IssuerStatus, error) {
	switch s {
	case "pending":
		return IssuerStatusPending, nil
	case "active":
		return IssuerStatusActive, nil
	case "suspended":
		return IssuerStatusSuspended, nil
	case "revoked":
		return IssuerStatusRevoked, nil
	default:
		return 0, fmt.Errorf("invalid issuer status: %s", s)
	}
}
