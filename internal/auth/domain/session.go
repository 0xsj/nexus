package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an active user session.
// Matches the sessions table schema in the database.
type Session struct {
	ID           string
	UserID       string
	DeviceID     *string
	DeviceName   *string
	IPAddress    *string
	UserAgent    *string
	CreatedAt    time.Time
	LastActiveAt time.Time
	ExpiresAt    time.Time
}

// NewSession creates a new session for a user.
func NewSession(
	userID string,
	deviceID *string,
	deviceName *string,
	ipAddress *string,
	userAgent *string,
	duration time.Duration,
) *Session {
	now := time.Now().UTC()

	return &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		DeviceID:     deviceID,
		DeviceName:   deviceName,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(duration),
	}
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired).
func (s *Session) IsValid() bool {
	return !s.IsExpired()
}

// UpdateActivity updates the last activity timestamp.
func (s *Session) UpdateActivity() {
	s.LastActiveAt = time.Now().UTC()
}

// ShouldRefresh checks if the session should be refreshed.
// Returns true if more than half the session duration has passed.
func (s *Session) ShouldRefresh() bool {
	elapsed := time.Since(s.CreatedAt)
	duration := s.ExpiresAt.Sub(s.CreatedAt)
	return elapsed > duration/2
}

// Extend extends the session expiration.
func (s *Session) Extend(duration time.Duration) {
	s.ExpiresAt = time.Now().UTC().Add(duration)
}

// TimeUntilExpiry returns the duration until the session expires.
func (s *Session) TimeUntilExpiry() time.Duration {
	return time.Until(s.ExpiresAt)
}
