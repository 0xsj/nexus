package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Session duration constants
const (
	// DefaultSessionDuration is the default session lifetime.
	DefaultSessionDuration = 24 * time.Hour

	// ExtendedSessionDuration is the extended session lifetime (remember me).
	ExtendedSessionDuration = 30 * 24 * time.Hour
)

// Session is the aggregate root for user sessions.
// It manages session lifecycle, token validation, and expiration.
type Session struct {
	eventsourcing.AggregateRoot

	id         SessionID
	userID     UserID
	tokenHash  string
	authMethod AuthMethod
	status     SessionStatus
	ipAddress  string
	userAgent  string
	expiresAt  time.Time
	createdAt  time.Time
	updatedAt  time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// NewSession creates a new Session aggregate.
func NewSession(
	id SessionID,
	userID UserID,
	token Token,
	authMethod AuthMethod,
	ipAddress string,
	userAgent string,
	duration time.Duration,
) (*Session, error) {
	if id.IsZero() {
		return nil, InvalidSessionID("Session.New", "")
	}

	if userID.IsZero() {
		return nil, InvalidUserID("Session.New", "")
	}

	if token.IsZero() {
		return nil, InvalidTokenError("Session.New", "token is required")
	}

	if !authMethod.IsValid() {
		return nil, AuthMethodNotEnabled("Session.New", authMethod.String())
	}

	if duration <= 0 {
		duration = DefaultSessionDuration
	}

	s := &Session{}
	s.InitAggregate(AggregateTypeSession, id.String())

	now := time.Now().UTC()
	expiresAt := now.Add(duration)

	s.Raise(s, SessionStartedEvent{
		BaseEvent:  newSessionBaseEvent(id),
		SessionID:  id.String(),
		UserID:     userID.String(),
		TokenHash:  token.HashString(),
		AuthMethod: authMethod.String(),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  expiresAt,
		StartedAt:  now,
	})

	return s, nil
}

// NewSessionFromEvents reconstructs a Session from events (for hydration).
func NewSessionFromEvents(id string) *Session {
	s := &Session{}
	s.InitAggregate(AggregateTypeSession, id)
	return s
}

// SessionFactory creates a factory for Session aggregates.
func SessionFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewSessionFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the session's ID.
func (s *Session) ID() SessionID {
	return s.id
}

// UserID returns the user ID associated with this session.
func (s *Session) UserID() UserID {
	return s.userID
}

// TokenHash returns the hashed session token.
func (s *Session) TokenHash() string {
	return s.tokenHash
}

// AuthMethod returns the authentication method used to create this session.
func (s *Session) AuthMethod() AuthMethod {
	return s.authMethod
}

// Status returns the session status.
func (s *Session) Status() SessionStatus {
	return s.status
}

// IPAddress returns the IP address that created this session.
func (s *Session) IPAddress() string {
	return s.ipAddress
}

// UserAgent returns the user agent that created this session.
func (s *Session) UserAgent() string {
	return s.userAgent
}

// ExpiresAt returns when the session expires.
func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

// CreatedAt returns when the session was created.
func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns when the session was last updated.
func (s *Session) UpdatedAt() time.Time {
	return s.updatedAt
}

// ============================================================================
// Query Methods
// ============================================================================

// IsActive returns true if the session is active and not expired.
func (s *Session) IsActive() bool {
	return s.status.IsActive() && !s.IsExpired()
}

// IsExpired returns true if the session has expired based on time.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.expiresAt)
}

// IsRevoked returns true if the session has been revoked.
func (s *Session) IsRevoked() bool {
	return s.status.IsRevoked()
}

// IsValid returns true if the session is valid for authentication.
func (s *Session) IsValid() bool {
	return s.status.IsActive() && !s.IsExpired()
}

// TTL returns the remaining time-to-live for the session.
// Returns 0 if the session has expired.
func (s *Session) TTL() time.Duration {
	if s.IsExpired() {
		return 0
	}
	return time.Until(s.expiresAt)
}

// VerifyToken verifies that the provided token matches the session's token hash.
func (s *Session) VerifyToken(token Token) bool {
	if token.IsZero() {
		return false
	}
	return token.HashString() == s.tokenHash
}

// ============================================================================
// Command Methods
// ============================================================================

// Refresh rotates the session token and extends the expiration.
func (s *Session) Refresh(newToken Token, duration time.Duration) error {
	if !s.IsValid() {
		if s.IsExpired() {
			return SessionExpiredError("Session.Refresh", s.id.String())
		}
		if s.IsRevoked() {
			return SessionRevokedError("Session.Refresh", s.id.String())
		}
		return SessionNotFound("Session.Refresh", s.id.String())
	}

	if newToken.IsZero() {
		return InvalidTokenError("Session.Refresh", "new token is required")
	}

	if duration <= 0 {
		duration = DefaultSessionDuration
	}

	now := time.Now().UTC()
	expiresAt := now.Add(duration)

	s.Raise(s, SessionRefreshedEvent{
		BaseEvent:    newSessionBaseEvent(s.id),
		SessionID:    s.id.String(),
		OldTokenHash: s.tokenHash,
		NewTokenHash: newToken.HashString(),
		ExpiresAt:    expiresAt,
		RefreshedAt:  now,
	})

	return nil
}

// Revoke revokes the session.
func (s *Session) Revoke(reason string) error {
	if !s.status.CanBeRevoked() {
		if s.IsRevoked() {
			return SessionRevokedError("Session.Revoke", s.id.String())
		}
		if s.status.IsExpired() {
			return SessionExpiredError("Session.Revoke", s.id.String())
		}
		return SessionNotFound("Session.Revoke", s.id.String())
	}

	s.Raise(s, SessionRevokedEvent{
		BaseEvent: newSessionBaseEvent(s.id),
		SessionID: s.id.String(),
		Reason:    reason,
		RevokedAt: time.Now().UTC(),
	})

	return nil
}

// MarkExpired marks the session as expired.
// This is typically called by a background job that cleans up expired sessions.
func (s *Session) MarkExpired() error {
	if s.status.IsTerminal() {
		return nil // Already in terminal state
	}

	if !s.IsExpired() {
		return SessionExpiredError("Session.MarkExpired", s.id.String()).
			WithMessage("session has not yet expired")
	}

	s.Raise(s, SessionExpiredEvent{
		BaseEvent: newSessionBaseEvent(s.id),
		SessionID: s.id.String(),
		ExpiredAt: time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (s *Session) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *SessionStartedEvent:
		s.onSessionStarted(e)
	case *SessionRefreshedEvent:
		s.onSessionRefreshed(e)
	case *SessionRevokedEvent:
		s.onSessionRevoked(e)
	case *SessionExpiredEvent:
		s.onSessionExpired(e)
	}
}

func (s *Session) onSessionStarted(e *SessionStartedEvent) {
	s.id, _ = ParseSessionID(e.SessionID)
	s.userID, _ = ParseUserID(e.UserID)
	s.tokenHash = e.TokenHash
	s.authMethod, _ = ParseAuthMethod(e.AuthMethod)
	s.status = SessionStatusActive
	s.ipAddress = e.IPAddress
	s.userAgent = e.UserAgent
	s.expiresAt = e.ExpiresAt
	s.createdAt = e.StartedAt
	s.updatedAt = e.StartedAt
}

func (s *Session) onSessionRefreshed(e *SessionRefreshedEvent) {
	s.tokenHash = e.NewTokenHash
	s.expiresAt = e.ExpiresAt
	s.updatedAt = e.RefreshedAt
}

func (s *Session) onSessionRevoked(e *SessionRevokedEvent) {
	s.status = SessionStatusRevoked
	s.updatedAt = e.RevokedAt
}

func (s *Session) onSessionExpired(e *SessionExpiredEvent) {
	s.status = SessionStatusExpired
	s.updatedAt = e.ExpiredAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (s *Session) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &s.AggregateRoot
}
