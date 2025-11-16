package domain

import (
	"time"
)

// Participant represents a user's participation in a conversation.
type Participant struct {
	UserID            string
	Role              ParticipantRole
	JoinedAt          time.Time
	LeftAt            *time.Time
	LastSeenAt        *time.Time
	LastReadMessageID *string
}

// ParticipantRole represents a participant's role in a conversation.
type ParticipantRole string

const (
	ParticipantRoleMember ParticipantRole = "member"
	ParticipantRoleAdmin  ParticipantRole = "admin"
	ParticipantRoleOwner  ParticipantRole = "owner"
)

// IsValid checks if the role is valid.
func (pr ParticipantRole) IsValid() bool {
	switch pr {
	case ParticipantRoleMember, ParticipantRoleAdmin, ParticipantRoleOwner:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (pr ParticipantRole) String() string {
	return string(pr)
}

// CanAddParticipants checks if the role can add participants.
func (pr ParticipantRole) CanAddParticipants() bool {
	return pr == ParticipantRoleAdmin || pr == ParticipantRoleOwner
}

// CanRemoveParticipants checks if the role can remove participants.
func (pr ParticipantRole) CanRemoveParticipants() bool {
	return pr == ParticipantRoleAdmin || pr == ParticipantRoleOwner
}

// NewParticipant creates a new participant.
func NewParticipant(userID string, role ParticipantRole) *Participant {
	return &Participant{
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now().UTC(),
	}
}

// UpdateLastSeen updates the last seen timestamp.
func (p *Participant) UpdateLastSeen() {
	now := time.Now().UTC()
	p.LastSeenAt = &now
}

// UpdateLastRead updates the last read message ID.
func (p *Participant) UpdateLastRead(messageID string) {
	p.LastReadMessageID = &messageID
}

// Leave marks the participant as having left the conversation.
func (p *Participant) Leave() {
	now := time.Now().UTC()
	p.LeftAt = &now
}

// HasLeft checks if the participant has left.
func (p *Participant) HasLeft() bool {
	return p.LeftAt != nil
}

// IsActive checks if the participant is active.
func (p *Participant) IsActive() bool {
	return !p.HasLeft()
}
