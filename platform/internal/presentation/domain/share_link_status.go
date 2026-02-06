package domain

import (
	"fmt"
)

// ShareLinkStatus represents the lifecycle status of a share link.
type ShareLinkStatus int

const (
	// ShareLinkStatusActive indicates the share link is active and accessible.
	ShareLinkStatusActive ShareLinkStatus = 1

	// ShareLinkStatusRevoked indicates the share link has been revoked.
	ShareLinkStatusRevoked ShareLinkStatus = 2

	// ShareLinkStatusExpired indicates the share link has expired.
	ShareLinkStatusExpired ShareLinkStatus = 3
)

// String returns the string representation of the share link status.
func (s ShareLinkStatus) String() string {
	switch s {
	case ShareLinkStatusActive:
		return "active"
	case ShareLinkStatusRevoked:
		return "revoked"
	case ShareLinkStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// IsActive returns true if the share link is active.
func (s ShareLinkStatus) IsActive() bool {
	return s == ShareLinkStatusActive
}

// IsRevoked returns true if the share link has been revoked.
func (s ShareLinkStatus) IsRevoked() bool {
	return s == ShareLinkStatusRevoked
}

// IsExpired returns true if the share link has expired.
func (s ShareLinkStatus) IsExpired() bool {
	return s == ShareLinkStatusExpired
}

// IsTerminal returns true if the share link is in a terminal state.
func (s ShareLinkStatus) IsTerminal() bool {
	return s == ShareLinkStatusRevoked || s == ShareLinkStatusExpired
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s ShareLinkStatus) CanTransitionTo(target ShareLinkStatus) bool {
	switch s {
	case ShareLinkStatusActive:
		return target == ShareLinkStatusRevoked || target == ShareLinkStatusExpired
	case ShareLinkStatusRevoked:
		return false
	case ShareLinkStatusExpired:
		return false
	default:
		return false
	}
}

// ParseShareLinkStatus parses a string into a ShareLinkStatus.
func ParseShareLinkStatus(s string) (ShareLinkStatus, error) {
	switch s {
	case "active":
		return ShareLinkStatusActive, nil
	case "revoked":
		return ShareLinkStatusRevoked, nil
	case "expired":
		return ShareLinkStatusExpired, nil
	default:
		return 0, fmt.Errorf("invalid share link status: %s", s)
	}
}
