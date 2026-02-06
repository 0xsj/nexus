package domain

import (
	"fmt"
)

// TemplateStatus represents the lifecycle status of a template.
type TemplateStatus int

const (
	// TemplateStatusActive indicates the template is active and can be used.
	TemplateStatusActive TemplateStatus = 1

	// TemplateStatusArchived indicates the template has been archived.
	TemplateStatusArchived TemplateStatus = 2
)

// String returns the string representation of the template status.
func (s TemplateStatus) String() string {
	switch s {
	case TemplateStatusActive:
		return "active"
	case TemplateStatusArchived:
		return "archived"
	default:
		return "unknown"
	}
}

// IsActive returns true if the template is active.
func (s TemplateStatus) IsActive() bool {
	return s == TemplateStatusActive
}

// IsArchived returns true if the template is archived.
func (s TemplateStatus) IsArchived() bool {
	return s == TemplateStatusArchived
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s TemplateStatus) CanTransitionTo(target TemplateStatus) bool {
	switch s {
	case TemplateStatusActive:
		return target == TemplateStatusArchived
	case TemplateStatusArchived:
		return false
	default:
		return false
	}
}

// ParseTemplateStatus parses a string into a TemplateStatus.
func ParseTemplateStatus(s string) (TemplateStatus, error) {
	switch s {
	case "active":
		return TemplateStatusActive, nil
	case "archived":
		return TemplateStatusArchived, nil
	default:
		return 0, fmt.Errorf("invalid template status: %s", s)
	}
}
