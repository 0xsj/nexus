package domain

import (
	"fmt"
)

// Visibility represents the privacy level for profile elements.
type Visibility int

const (
	// VisibilityPublic indicates the element is visible to anyone.
	VisibilityPublic Visibility = 1

	// VisibilityConnectionsOnly indicates the element is visible only to connected users.
	VisibilityConnectionsOnly Visibility = 2

	// VisibilityRequestRequired indicates the viewer must request access.
	VisibilityRequestRequired Visibility = 3

	// VisibilityPrivate indicates the element is hidden from the profile.
	VisibilityPrivate Visibility = 4
)

// String returns the string representation of the Visibility.
func (v Visibility) String() string {
	switch v {
	case VisibilityPublic:
		return "public"
	case VisibilityConnectionsOnly:
		return "connections_only"
	case VisibilityRequestRequired:
		return "request_required"
	case VisibilityPrivate:
		return "private"
	default:
		return "unknown"
	}
}

// ParseVisibility parses a string into a Visibility.
func ParseVisibility(s string) (Visibility, error) {
	switch s {
	case "public":
		return VisibilityPublic, nil
	case "connections_only":
		return VisibilityConnectionsOnly, nil
	case "request_required":
		return VisibilityRequestRequired, nil
	case "private":
		return VisibilityPrivate, nil
	default:
		return 0, fmt.Errorf("unknown visibility: %s", s)
	}
}

// IsPublic returns true if the visibility is public.
func (v Visibility) IsPublic() bool {
	return v == VisibilityPublic
}

// IsVisible returns true if the visibility is not private.
// Elements with this visibility may still require additional checks
// (e.g., connection status for ConnectionsOnly, or approval for RequestRequired).
func (v Visibility) IsVisible() bool {
	return v != VisibilityPrivate
}

// IsValid returns true if the Visibility is a known, valid value.
func (v Visibility) IsValid() bool {
	switch v {
	case VisibilityPublic,
		VisibilityConnectionsOnly,
		VisibilityRequestRequired,
		VisibilityPrivate:
		return true
	default:
		return false
	}
}
