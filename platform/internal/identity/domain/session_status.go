package domain

import (
	"fmt"
)

// SessionStatus represents the current status of a session.
type SessionStatus int

const (
	// SessionStatusUnknown indicates an unknown status.
	SessionStatusUnknown SessionStatus = iota

	// SessionStatusActive indicates the session is active and valid.
	SessionStatusActive

	// SessionStatusExpired indicates the session has expired.
	SessionStatusExpired

	// SessionStatusRevoked indicates the session has been explicitly revoked.
	SessionStatusRevoked
)

// String returns the string representation of the session status.
func (s SessionStatus) String() string {
	switch s {
	case SessionStatusActive:
		return "active"
	case SessionStatusExpired:
		return "expired"
	case SessionStatusRevoked:
		return "revoked"
	default:
		return "unknown"
	}
}

// ParseSessionStatus parses a string into a SessionStatus.
func ParseSessionStatus(str string) (SessionStatus, error) {
	switch str {
	case "active":
		return SessionStatusActive, nil
	case "expired":
		return SessionStatusExpired, nil
	case "revoked":
		return SessionStatusRevoked, nil
	case "unknown":
		return SessionStatusUnknown, nil
	default:
		return SessionStatusUnknown, fmt.Errorf("invalid session status: %s", str)
	}
}

// MustParseSessionStatus parses a string into a SessionStatus and panics if invalid.
// Only use for constants or tests.
func MustParseSessionStatus(str string) SessionStatus {
	s, err := ParseSessionStatus(str)
	if err != nil {
		panic(err)
	}
	return s
}

// IsValid returns true if the status is a known, valid status.
func (s SessionStatus) IsValid() bool {
	switch s {
	case SessionStatusActive, SessionStatusExpired, SessionStatusRevoked:
		return true
	default:
		return false
	}
}

// IsZero returns true if the status is the zero value (unknown).
func (s SessionStatus) IsZero() bool {
	return s == SessionStatusUnknown
}

// IsActive returns true if the session is active.
func (s SessionStatus) IsActive() bool {
	return s == SessionStatusActive
}

// IsExpired returns true if the session has expired.
func (s SessionStatus) IsExpired() bool {
	return s == SessionStatusExpired
}

// IsRevoked returns true if the session has been revoked.
func (s SessionStatus) IsRevoked() bool {
	return s == SessionStatusRevoked
}

// IsTerminal returns true if the session is in a terminal state (expired or revoked).
func (s SessionStatus) IsTerminal() bool {
	return s == SessionStatusExpired || s == SessionStatusRevoked
}

// CanBeRevoked returns true if the session can be revoked.
func (s SessionStatus) CanBeRevoked() bool {
	return s == SessionStatusActive
}
