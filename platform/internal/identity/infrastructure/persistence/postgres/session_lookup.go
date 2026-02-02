package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/persistence/postgres/generated"
)

// SessionLookup provides query-side operations for sessions.
// It reads from the projection tables (not event store).
type SessionLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewSessionLookup creates a new SessionLookup.
func NewSessionLookup(pool *pgxpool.Pool) *SessionLookup {
	return &SessionLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// FindByID finds a session by ID from the projection.
func (l *SessionLookup) FindByID(ctx context.Context, id domain.SessionID) (*SessionProjection, error) {
	row, err := l.queries.GetSessionByID(ctx, uuidFromSessionID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.SessionNotFound("SessionLookup.FindByID", id.String())
		}
		return nil, err
	}

	return sessionRowToProjection(row), nil
}

// FindByTokenHash finds a session by token hash.
func (l *SessionLookup) FindByTokenHash(ctx context.Context, tokenHash string) (*SessionProjection, error) {
	row, err := l.queries.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.SessionNotFound("SessionLookup.FindByTokenHash", "")
		}
		return nil, err
	}

	return sessionRowToProjection(row), nil
}

// FindActiveByUserID finds all active sessions for a user.
func (l *SessionLookup) FindActiveByUserID(ctx context.Context, userID domain.UserID) ([]SessionProjection, error) {
	rows, err := l.queries.GetActiveSessionsForUser(ctx, uuidFromUserID(userID))
	if err != nil {
		return nil, err
	}

	sessions := make([]SessionProjection, len(rows))
	for i, row := range rows {
		sessions[i] = *sessionRowToProjection(row)
	}

	return sessions, nil
}

// CountActiveByUserID counts active sessions for a user.
func (l *SessionLookup) CountActiveByUserID(ctx context.Context, userID domain.UserID) (int, error) {
	count, err := l.queries.CountActiveSessionsForUser(ctx, uuidFromUserID(userID))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// ExistsByID checks if a session exists by ID.
func (l *SessionLookup) ExistsByID(ctx context.Context, id domain.SessionID) (bool, error) {
	return l.queries.SessionExistsByID(ctx, uuidFromSessionID(id))
}

// ============================================================================
// Projection Types
// ============================================================================

// SessionProjection represents a session from the read model.
type SessionProjection struct {
	ID         domain.SessionID
	UserID     domain.UserID
	TokenHash  string
	AuthMethod domain.AuthMethod
	Status     domain.SessionStatus
	IPAddress  string
	UserAgent  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Version    int
}

// IsActive returns true if the session is active and not expired.
func (s *SessionProjection) IsActive() bool {
	return s.Status == domain.SessionStatusActive && time.Now().UTC().Before(s.ExpiresAt)
}

// IsExpired returns true if the session has expired.
func (s *SessionProjection) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// TTL returns the remaining time-to-live for the session.
func (s *SessionProjection) TTL() time.Duration {
	if s.IsExpired() {
		return 0
	}
	return time.Until(s.ExpiresAt)
}

// ============================================================================
// Helpers
// ============================================================================

// sessionRowToProjection converts a database row to a SessionProjection.
func sessionRowToProjection(row generated.IdentitySession) *SessionProjection {
	sessionID := sessionIDFromUUID(row.ID)
	userID := userIDFromUUID(row.UserID)
	authMethod, _ := domain.ParseAuthMethod(row.AuthMethod)
	status, _ := domain.ParseSessionStatus(row.Status)

	var ipAddress, userAgent string
	if row.IpAddress != nil {
		ipAddress = *row.IpAddress
	}
	if row.UserAgent != nil {
		userAgent = *row.UserAgent
	}

	return &SessionProjection{
		ID:         sessionID,
		UserID:     userID,
		TokenHash:  row.TokenHash,
		AuthMethod: authMethod,
		Status:     status,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		Version:    int(row.Version),
	}
}
