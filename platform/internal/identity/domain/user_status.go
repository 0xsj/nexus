package domain

import (
	"fmt"
)

// UserStatus represents the current status of a user account.
type UserStatus int

const (
	// UserStatusUnknown indicates an unknown status.
	UserStatusUnknown UserStatus = iota

	// UserStatusPending indicates the user has registered but not yet verified.
	UserStatusPending

	// UserStatusActive indicates the user account is active and verified.
	UserStatusActive

	// UserStatusSuspended indicates the user account has been suspended.
	UserStatusSuspended

	// UserStatusDeleted indicates the user account has been soft-deleted.
	UserStatusDeleted
)

// String returns the string representation of the user status.
func (s UserStatus) String() string {
	switch s {
	case UserStatusPending:
		return "pending"
	case UserStatusActive:
		return "active"
	case UserStatusSuspended:
		return "suspended"
	case UserStatusDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

// ParseUserStatus parses a string into a UserStatus.
func ParseUserStatus(str string) (UserStatus, error) {
	switch str {
	case "pending":
		return UserStatusPending, nil
	case "active":
		return UserStatusActive, nil
	case "suspended":
		return UserStatusSuspended, nil
	case "deleted":
		return UserStatusDeleted, nil
	case "unknown":
		return UserStatusUnknown, nil
	default:
		return UserStatusUnknown, fmt.Errorf("invalid user status: %s", str)
	}
}

// MustParseUserStatus parses a string into a UserStatus and panics if invalid.
// Only use for constants or tests.
func MustParseUserStatus(str string) UserStatus {
	s, err := ParseUserStatus(str)
	if err != nil {
		panic(err)
	}
	return s
}

// IsValid returns true if the status is a known, valid status.
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusPending, UserStatusActive, UserStatusSuspended, UserStatusDeleted:
		return true
	default:
		return false
	}
}

// IsZero returns true if the status is the zero value (unknown).
func (s UserStatus) IsZero() bool {
	return s == UserStatusUnknown
}

// IsActive returns true if the user account is active.
func (s UserStatus) IsActive() bool {
	return s == UserStatusActive
}

// IsPending returns true if the user account is pending verification.
func (s UserStatus) IsPending() bool {
	return s == UserStatusPending
}

// IsSuspended returns true if the user account is suspended.
func (s UserStatus) IsSuspended() bool {
	return s == UserStatusSuspended
}

// IsDeleted returns true if the user account is soft-deleted.
func (s UserStatus) IsDeleted() bool {
	return s == UserStatusDeleted
}

// CanAuthenticate returns true if the user is allowed to authenticate.
func (s UserStatus) CanAuthenticate() bool {
	return s == UserStatusActive || s == UserStatusPending
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s UserStatus) CanTransitionTo(target UserStatus) bool {
	switch s {
	case UserStatusPending:
		// Pending can transition to active, suspended, or deleted
		return target == UserStatusActive || target == UserStatusSuspended || target == UserStatusDeleted
	case UserStatusActive:
		// Active can transition to suspended or deleted
		return target == UserStatusSuspended || target == UserStatusDeleted
	case UserStatusSuspended:
		// Suspended can transition to active or deleted
		return target == UserStatusActive || target == UserStatusDeleted
	case UserStatusDeleted:
		// Deleted is terminal (could allow reactivation if needed)
		return false
	default:
		return false
	}
}
