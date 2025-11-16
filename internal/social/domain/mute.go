package domain

import (
	"time"

	"github.com/0xsj/result"
	"github.com/google/uuid"
)

// Mute represents a mute relationship between two users.
// Muting hides content without unfollowing.
type Mute struct {
	ID         string
	MuterID    string
	MutedID    string
	CreatedAt  time.Time
	MutedUntil *time.Time // nil = permanent mute
}

// NewMute creates a new mute relationship.
func NewMute(muterID, mutedID string, mutedUntil *time.Time) result.Result[*Mute] {
	// Validate: cannot mute yourself
	if muterID == mutedID {
		return result.Err[*Mute](ErrCannotMuteSelf())
	}

	// Validate: valid user IDs
	if muterID == "" || mutedID == "" {
		return result.Err[*Mute](ErrInvalidUserID())
	}

	// Validate: mute duration must be in future
	if mutedUntil != nil && mutedUntil.Before(time.Now().UTC()) {
		return result.Err[*Mute](ErrInvalidMuteDuration())
	}

	mute := &Mute{
		ID:         uuid.New().String(),
		MuterID:    muterID,
		MutedID:    mutedID,
		CreatedAt:  time.Now().UTC(),
		MutedUntil: mutedUntil,
	}

	return result.Ok(mute)
}

// IsExpired checks if a temporary mute has expired.
func (m *Mute) IsExpired() bool {
	if m.MutedUntil == nil {
		return false // Permanent mute
	}
	return time.Now().UTC().After(*m.MutedUntil)
}

// IsPermanent checks if the mute is permanent.
func (m *Mute) IsPermanent() bool {
	return m.MutedUntil == nil
}

// Extend extends a temporary mute or makes it permanent.
func (m *Mute) Extend(mutedUntil *time.Time) error {
	if mutedUntil != nil && mutedUntil.Before(time.Now().UTC()) {
		return ErrInvalidMuteDuration()
	}

	m.MutedUntil = mutedUntil
	return nil
}
