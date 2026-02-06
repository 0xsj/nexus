package domain

import (
	"fmt"
)

// VouchStatus represents the lifecycle status of a vouch.
type VouchStatus int

const (
	// VouchStatusPending indicates the vouch has been given but not yet accepted.
	VouchStatusPending VouchStatus = 1

	// VouchStatusAccepted indicates the vouch has been accepted by the vouchee.
	VouchStatusAccepted VouchStatus = 2

	// VouchStatusRevoked indicates the vouch has been revoked.
	VouchStatusRevoked VouchStatus = 3

	// VouchStatusExpired indicates the vouch has expired.
	VouchStatusExpired VouchStatus = 4
)

// String returns the string representation of the vouch status.
func (s VouchStatus) String() string {
	switch s {
	case VouchStatusPending:
		return "pending"
	case VouchStatusAccepted:
		return "accepted"
	case VouchStatusRevoked:
		return "revoked"
	case VouchStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// IsPending returns true if the vouch is pending.
func (s VouchStatus) IsPending() bool {
	return s == VouchStatusPending
}

// IsAccepted returns true if the vouch has been accepted.
func (s VouchStatus) IsAccepted() bool {
	return s == VouchStatusAccepted
}

// IsRevoked returns true if the vouch has been revoked.
func (s VouchStatus) IsRevoked() bool {
	return s == VouchStatusRevoked
}

// IsExpired returns true if the vouch has expired.
func (s VouchStatus) IsExpired() bool {
	return s == VouchStatusExpired
}

// IsTerminal returns true if the vouch is in a terminal state.
func (s VouchStatus) IsTerminal() bool {
	return s == VouchStatusRevoked || s == VouchStatusExpired
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s VouchStatus) CanTransitionTo(target VouchStatus) bool {
	switch s {
	case VouchStatusPending:
		return target == VouchStatusAccepted || target == VouchStatusRevoked || target == VouchStatusExpired
	case VouchStatusAccepted:
		return target == VouchStatusRevoked || target == VouchStatusExpired
	case VouchStatusRevoked:
		return false
	case VouchStatusExpired:
		return false
	default:
		return false
	}
}

// ParseVouchStatus parses a string into a VouchStatus.
func ParseVouchStatus(s string) (VouchStatus, error) {
	switch s {
	case "pending":
		return VouchStatusPending, nil
	case "accepted":
		return VouchStatusAccepted, nil
	case "revoked":
		return VouchStatusRevoked, nil
	case "expired":
		return VouchStatusExpired, nil
	default:
		return 0, fmt.Errorf("invalid vouch status: %s", s)
	}
}
