package domain

import (
	"time"

	"github.com/google/uuid"
)

// MagicLink represents a magic link token for passwordless authentication.
// This is an in-memory/Redis entity, not stored in PostgreSQL.
type MagicLink struct {
	ID        string
	Email     string
	Token     string
	Used      bool
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
}

// NewMagicLink creates a new magic link.
func NewMagicLink(email, token, ipAddress, userAgent string, ttl time.Duration) *MagicLink {
	now := time.Now().UTC()

	return &MagicLink{
		ID:        uuid.New().String(),
		Email:     email,
		Token:     token,
		Used:      false,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// IsExpired checks if the magic link has expired.
func (m *MagicLink) IsExpired() bool {
	return time.Now().UTC().After(m.ExpiresAt)
}

// IsValid checks if the magic link is valid (not expired and not used).
func (m *MagicLink) IsValid() bool {
	return !m.IsExpired() && !m.Used
}

// MarkAsUsed marks the magic link as used.
func (m *MagicLink) MarkAsUsed() {
	m.Used = true
}
