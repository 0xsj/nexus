package session

import (
	"time"
)

// Session represents an active user session.
type Session struct {
	ID           string    // Unique session ID (UUID)
	UserID       string    // User this session belongs to
	RefreshToken string    // JWT refresh token associated with this session
	IPAddress    string    // IP address of the client
	UserAgent    string    // User agent string
	ExpiresAt    time.Time // When the session expires
	CreatedAt    time.Time // When the session was created
	UpdatedAt    time.Time // When the session was last updated
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired).
func (s *Session) IsValid() bool {
	return !s.IsExpired()
}

// TimeUntilExpiry returns the duration until the session expires.
// Returns 0 if already expired.
func (s *Session) TimeUntilExpiry() time.Duration {
	if s.IsExpired() {
		return 0
	}
	return time.Until(s.ExpiresAt)
}

// UpdateRefreshToken updates the session's refresh token and timestamp.
func (s *Session) UpdateRefreshToken(newToken string) {
	s.RefreshToken = newToken
	s.UpdatedAt = time.Now()
}

// Extend extends the session expiry by the given duration.
func (s *Session) Extend(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
	s.UpdatedAt = time.Now()
}

// Refresh updates both the refresh token and extends the session.
func (s *Session) Refresh(newToken string, duration time.Duration) {
	s.RefreshToken = newToken
	s.ExpiresAt = time.Now().Add(duration)
	s.UpdatedAt = time.Now()
}
