package domain

import "time"

// Member represents a user's membership in an organization.
type Member struct {
	id       MemberID
	userID   string
	role     Role
	joinedAt time.Time
}

// NewMember creates a new Member.
func NewMember(id MemberID, userID string, role Role, joinedAt time.Time) Member {
	return Member{
		id:       id,
		userID:   userID,
		role:     role,
		joinedAt: joinedAt,
	}
}

// ID returns the member's ID.
func (m Member) ID() MemberID { return m.id }

// UserID returns the user's ID.
func (m Member) UserID() string { return m.userID }

// Role returns the member's role.
func (m Member) Role() Role { return m.role }

// JoinedAt returns when the member joined.
func (m Member) JoinedAt() time.Time { return m.joinedAt }

// IsOwner returns true if the member is an owner.
func (m Member) IsOwner() bool { return m.role == RoleOwner }
