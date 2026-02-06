package domain

import (
	"fmt"
	"time"
)

// InvitationStatus represents the status of an invitation.
type InvitationStatus int

const (
	InvitationStatusPending  InvitationStatus = 1
	InvitationStatusAccepted InvitationStatus = 2
	InvitationStatusDeclined InvitationStatus = 3
	InvitationStatusExpired  InvitationStatus = 4
)

// String returns the string representation.
func (s InvitationStatus) String() string {
	switch s {
	case InvitationStatusPending:
		return "pending"
	case InvitationStatusAccepted:
		return "accepted"
	case InvitationStatusDeclined:
		return "declined"
	case InvitationStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// ParseInvitationStatus parses a string into an InvitationStatus.
func ParseInvitationStatus(s string) (InvitationStatus, error) {
	switch s {
	case "pending":
		return InvitationStatusPending, nil
	case "accepted":
		return InvitationStatusAccepted, nil
	case "declined":
		return InvitationStatusDeclined, nil
	case "expired":
		return InvitationStatusExpired, nil
	default:
		return 0, fmt.Errorf("invalid invitation status: %s", s)
	}
}

// IsPending returns true if the invitation is pending.
func (s InvitationStatus) IsPending() bool { return s == InvitationStatusPending }

// IsTerminal returns true if the invitation is in a terminal state.
func (s InvitationStatus) IsTerminal() bool {
	return s == InvitationStatusAccepted || s == InvitationStatusDeclined || s == InvitationStatusExpired
}

// Invitation represents a pending invitation to join an organization.
type Invitation struct {
	id        InvitationID
	email     string
	role      Role
	inviterID string
	status    InvitationStatus
	expiresAt time.Time
	createdAt time.Time
}

// NewInvitation creates a new Invitation.
func NewInvitation(id InvitationID, email string, role Role, inviterID string, expiresAt time.Time) Invitation {
	return Invitation{
		id:        id,
		email:     email,
		role:      role,
		inviterID: inviterID,
		status:    InvitationStatusPending,
		expiresAt: expiresAt,
		createdAt: time.Now().UTC(),
	}
}

// ID returns the invitation ID.
func (i Invitation) ID() InvitationID { return i.id }

// Email returns the invitee's email.
func (i Invitation) Email() string { return i.email }

// Role returns the offered role.
func (i Invitation) Role() Role { return i.role }

// InviterID returns who sent the invitation.
func (i Invitation) InviterID() string { return i.inviterID }

// Status returns the invitation status.
func (i Invitation) Status() InvitationStatus { return i.status }

// ExpiresAt returns when the invitation expires.
func (i Invitation) ExpiresAt() time.Time { return i.expiresAt }

// CreatedAt returns when the invitation was created.
func (i Invitation) CreatedAt() time.Time { return i.createdAt }

// IsExpired returns true if the invitation has expired.
func (i Invitation) IsExpired() bool {
	return time.Now().After(i.expiresAt)
}
