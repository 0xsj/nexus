package domain

import (
	"fmt"
)

// PresentationStatus represents the lifecycle status of a presentation.
type PresentationStatus int

const (
	// PresentationStatusActive indicates the presentation is active and valid.
	PresentationStatusActive PresentationStatus = 1

	// PresentationStatusRevoked indicates the presentation has been revoked.
	PresentationStatusRevoked PresentationStatus = 2
)

// String returns the string representation of the presentation status.
func (s PresentationStatus) String() string {
	switch s {
	case PresentationStatusActive:
		return "active"
	case PresentationStatusRevoked:
		return "revoked"
	default:
		return "unknown"
	}
}

// IsActive returns true if the presentation is active.
func (s PresentationStatus) IsActive() bool {
	return s == PresentationStatusActive
}

// IsRevoked returns true if the presentation has been revoked.
func (s PresentationStatus) IsRevoked() bool {
	return s == PresentationStatusRevoked
}

// IsTerminal returns true if the presentation is in a terminal state.
func (s PresentationStatus) IsTerminal() bool {
	return s == PresentationStatusRevoked
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s PresentationStatus) CanTransitionTo(target PresentationStatus) bool {
	switch s {
	case PresentationStatusActive:
		return target == PresentationStatusRevoked
	case PresentationStatusRevoked:
		return false
	default:
		return false
	}
}

// ParsePresentationStatus parses a string into a PresentationStatus.
func ParsePresentationStatus(s string) (PresentationStatus, error) {
	switch s {
	case "active":
		return PresentationStatusActive, nil
	case "revoked":
		return PresentationStatusRevoked, nil
	default:
		return 0, fmt.Errorf("invalid presentation status: %s", s)
	}
}
