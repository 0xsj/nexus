package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// ============================================================================
// Session Entity
// ============================================================================

// Session represents an authenticated session.
// Sessions are created when a user authenticates and can be revoked.
type Session struct {
	id        string
	userID    string
	status    SessionStatus
	method    AuthMethod
	tokenHash string

	// Client info
	userAgent string
	ipAddress string
	device    string

	// Timestamps
	createdAt  time.Time
	expiresAt  time.Time
	revokedAt  time.Time
	lastSeenAt time.Time
}

// ============================================================================
// Constructor
// ============================================================================

// NewSession creates a new active session.
func NewSession(
	id string,
	userID string,
	method AuthMethod,
	tokenHash string,
	expiresAt time.Time,
	userAgent string,
	ipAddress string,
) *Session {
	now := time.Now()
	return &Session{
		id:         id,
		userID:     userID,
		status:     SessionStatusActive,
		method:     method,
		tokenHash:  tokenHash,
		userAgent:  userAgent,
		ipAddress:  ipAddress,
		createdAt:  now,
		expiresAt:  expiresAt,
		lastSeenAt: now,
	}
}

// ReconstituteSession creates a Session from persisted data.
func ReconstituteSession(
	id string,
	userID string,
	status SessionStatus,
	method AuthMethod,
	tokenHash string,
	userAgent string,
	ipAddress string,
	device string,
	createdAt time.Time,
	expiresAt time.Time,
	revokedAt time.Time,
	lastSeenAt time.Time,
) *Session {
	return &Session{
		id:         id,
		userID:     userID,
		status:     status,
		method:     method,
		tokenHash:  tokenHash,
		userAgent:  userAgent,
		ipAddress:  ipAddress,
		device:     device,
		createdAt:  createdAt,
		expiresAt:  expiresAt,
		revokedAt:  revokedAt,
		lastSeenAt: lastSeenAt,
	}
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the session ID.
func (s *Session) ID() string {
	return s.id
}

// UserID returns the user ID.
func (s *Session) UserID() string {
	return s.userID
}

// Status returns the session status.
func (s *Session) Status() SessionStatus {
	return s.status
}

// Method returns the authentication method used.
func (s *Session) Method() AuthMethod {
	return s.method
}

// TokenHash returns the token hash.
func (s *Session) TokenHash() string {
	return s.tokenHash
}

// UserAgent returns the user agent.
func (s *Session) UserAgent() string {
	return s.userAgent
}

// IPAddress returns the IP address.
func (s *Session) IPAddress() string {
	return s.ipAddress
}

// Device returns the device identifier.
func (s *Session) Device() string {
	return s.device
}

// CreatedAt returns when the session was created.
func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

// ExpiresAt returns when the session expires.
func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

// RevokedAt returns when the session was revoked (zero if not revoked).
func (s *Session) RevokedAt() time.Time {
	return s.revokedAt
}

// LastSeenAt returns when the session was last used.
func (s *Session) LastSeenAt() time.Time {
	return s.lastSeenAt
}

// ============================================================================
// Status Checks
// ============================================================================

// IsActive returns true if the session is active and not expired.
func (s *Session) IsActive() bool {
	if s.status != SessionStatusActive {
		return false
	}
	return time.Now().Before(s.expiresAt)
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

// IsRevoked returns true if the session was revoked.
func (s *Session) IsRevoked() bool {
	return s.status == SessionStatusRevoked
}

// IsValid returns true if the session is valid (active and not expired).
func (s *Session) IsValid() bool {
	return s.IsActive() && !s.IsExpired()
}

// TimeUntilExpiry returns the duration until the session expires.
func (s *Session) TimeUntilExpiry() time.Duration {
	return time.Until(s.expiresAt)
}

// ============================================================================
// Commands
// ============================================================================

// Touch updates the last seen timestamp.
func (s *Session) Touch() {
	s.lastSeenAt = time.Now()
}

// Extend extends the session expiration.
func (s *Session) Extend(duration time.Duration) error {
	if !s.IsActive() {
		return ErrSessionExpired("Session.Extend", s.id)
	}

	s.expiresAt = time.Now().Add(duration)
	s.lastSeenAt = time.Now()
	return nil
}

// Revoke revokes the session.
func (s *Session) Revoke() error {
	if s.status == SessionStatusRevoked {
		return nil // Already revoked
	}

	s.status = SessionStatusRevoked
	s.revokedAt = time.Now()
	return nil
}

// Expire marks the session as expired.
func (s *Session) Expire() {
	s.status = SessionStatusExpired
}

// SetDevice sets the device identifier.
func (s *Session) SetDevice(device string) {
	s.device = device
}

// ============================================================================
// Validation
// ============================================================================

// Validate validates the session for use.
func (s *Session) Validate() error {
	if s.status == SessionStatusRevoked {
		return ErrSessionRevoked("Session.Validate", s.id)
	}

	if s.IsExpired() {
		return ErrSessionExpired("Session.Validate", s.id)
	}

	return nil
}

// ValidateToken validates that the provided token hash matches.
func (s *Session) ValidateToken(tokenHash string) error {
	if err := s.Validate(); err != nil {
		return err
	}

	if s.tokenHash != tokenHash {
		return ErrTokenInvalid("Session.ValidateToken", "token mismatch")
	}

	return nil
}

// ============================================================================
// Session Token Generation
// ============================================================================

// GenerateSessionToken generates a cryptographically secure session token.
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateSessionID generates a new session ID.
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
