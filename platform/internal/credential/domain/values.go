package domain

import "time"

// ============================================================================
// Credential Status
// ============================================================================

// Status represents the lifecycle status of a credential.
type Status string

const (
	// StatusPending indicates the credential has been requested but not yet issued.
	StatusPending Status = "pending"

	// StatusActive indicates the credential has been issued and is valid.
	StatusActive Status = "active"

	// StatusSuspended indicates the credential is temporarily suspended.
	StatusSuspended Status = "suspended"

	// StatusRevoked indicates the credential has been permanently revoked.
	StatusRevoked Status = "revoked"

	// StatusExpired indicates the credential has expired.
	StatusExpired Status = "expired"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// IsValid checks if the status is a valid status value.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusActive, StatusSuspended, StatusRevoked, StatusExpired:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the status is a terminal state (cannot transition out of).
func (s Status) IsTerminal() bool {
	return s == StatusRevoked || s == StatusExpired
}

// CanTransitionTo checks if a transition to the target status is valid.
func (s Status) CanTransitionTo(target Status) bool {
	switch s {
	case StatusPending:
		return target == StatusActive || target == StatusRevoked
	case StatusActive:
		return target == StatusSuspended || target == StatusRevoked || target == StatusExpired
	case StatusSuspended:
		return target == StatusActive || target == StatusRevoked
	case StatusRevoked:
		return false // Terminal state
	case StatusExpired:
		return false // Terminal state
	default:
		return false
	}
}

// ============================================================================
// Revocation Info
// ============================================================================

// RevocationInfo contains details about a credential revocation.
type RevocationInfo struct {
	RevokedBy string    `json:"revoked_by"`
	Reason    string    `json:"reason"`
	RevokedAt time.Time `json:"revoked_at"`
}

// ============================================================================
// Suspension Info
// ============================================================================

// SuspensionInfo contains details about a credential suspension.
type SuspensionInfo struct {
	SuspendedBy string     `json:"suspended_by"`
	Reason      string     `json:"reason"`
	SuspendedAt time.Time  `json:"suspended_at"`
	Until       *time.Time `json:"until,omitempty"`
}
