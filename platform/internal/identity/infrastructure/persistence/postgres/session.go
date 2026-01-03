package postgres

import (
	"context"
	"database/sql"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Session Repository
// ============================================================================

// SessionRepository implements domain.SessionRepository using PostgreSQL.
type SessionRepository struct {
	adapter *postgres.BaseAdapter
}

// Ensure SessionRepository implements domain.SessionRepository.
var _ domain.SessionRepository = (*SessionRepository)(nil)

// NewSessionRepository creates a new session repository.
func NewSessionRepository(adapter *postgres.BaseAdapter) *SessionRepository {
	return &SessionRepository{
		adapter: adapter,
	}
}

// ============================================================================
// Save
// ============================================================================

// Save persists a session (create or update).
func (r *SessionRepository) Save(ctx context.Context, session *domain.Session) error {
	const op = "postgres.SessionRepository.Save"

	exists, err := r.existsByID(ctx, session.ID())
	if err != nil {
		return errors.Wrap(err, op)
	}

	row := ToSessionRow(session)

	if exists {
		return r.update(ctx, row)
	}

	return r.insert(ctx, row)
}

func (r *SessionRepository) insert(ctx context.Context, row *SessionRow) error {
	const op = "postgres.SessionRepository.insert"

	_, err := r.adapter.Exec(ctx, querySessionInsert,
		row.ID,
		row.UserID,
		row.TokenHash,
		row.AuthMethod,
		row.IPAddress,
		row.UserAgent,
		row.DeviceID,
		row.Status,
		row.CreatedAt,
		row.ExpiresAt,
		row.LastActiveAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *SessionRepository) update(ctx context.Context, row *SessionRow) error {
	const op = "postgres.SessionRepository.update"

	_, err := r.adapter.Exec(ctx, querySessionUpdate,
		row.ID,
		row.Status,
		row.LastActiveAt,
		row.RevokedAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// Find
// ============================================================================

// FindByID finds a session by ID.
func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	const op = "postgres.SessionRepository.FindByID"

	row, err := r.scanSession(ctx, querySessionByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound(op, id)
		}
		return nil, errors.Wrap(err, op)
	}

	return row.ToDomainSession(), nil
}

// FindByTokenHash finds a session by token hash.
func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	const op = "postgres.SessionRepository.FindByTokenHash"

	row, err := r.scanSession(ctx, querySessionByTokenHash, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound(op, tokenHash)
		}
		return nil, errors.Wrap(err, op)
	}

	return row.ToDomainSession(), nil
}

// FindByUserID finds all sessions for a user.
func (r *SessionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	const op = "postgres.SessionRepository.FindByUserID"

	return r.findSessions(ctx, op, querySessionsByUserID, userID)
}

// FindActiveByUserID finds all active sessions for a user.
func (r *SessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	const op = "postgres.SessionRepository.FindActiveByUserID"

	return r.findSessions(ctx, op, querySessionsActiveByUserID, userID)
}

// ============================================================================
// Delete
// ============================================================================

// Delete removes a session.
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	const op = "postgres.SessionRepository.Delete"

	result, err := r.adapter.Exec(ctx, querySessionDelete, id)
	if err != nil {
		return errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound(op, id)
	}

	return nil
}

// DeleteByUserID removes all sessions for a user.
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	const op = "postgres.SessionRepository.DeleteByUserID"

	_, err := r.adapter.Exec(ctx, querySessionDeleteByUserID, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// DeleteExpired removes all expired sessions.
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	const op = "postgres.SessionRepository.DeleteExpired"

	result, err := r.adapter.Exec(ctx, querySessionDeleteExpired)
	if err != nil {
		return 0, errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, op)
	}

	return rowsAffected, nil
}

// ============================================================================
// Count
// ============================================================================

// CountByUserID counts sessions for a user.
func (r *SessionRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	const op = "postgres.SessionRepository.CountByUserID"

	var count int
	row := r.adapter.Executor().QueryRow(ctx, querySessionCountByUserID, userID)
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

// CountActiveByUserID counts active sessions for a user.
func (r *SessionRepository) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	const op = "postgres.SessionRepository.CountActiveByUserID"

	var count int
	row := r.adapter.Executor().QueryRow(ctx, querySessionCountActiveByUserID, userID)
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

// ============================================================================
// Helpers
// ============================================================================

func (r *SessionRepository) existsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1)", id)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *SessionRepository) scanSession(ctx context.Context, query string, args ...interface{}) (*SessionRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)

	var sessionRow SessionRow
	err := row.Scan(
		&sessionRow.ID,
		&sessionRow.UserID,
		&sessionRow.TokenHash,
		&sessionRow.AuthMethod,
		&sessionRow.IPAddress,
		&sessionRow.UserAgent,
		&sessionRow.DeviceID,
		&sessionRow.Status,
		&sessionRow.CreatedAt,
		&sessionRow.ExpiresAt,
		&sessionRow.LastActiveAt,
		&sessionRow.RevokedAt,
	)
	if err != nil {
		return nil, err
	}

	return &sessionRow, nil
}

func (r *SessionRepository) findSessions(ctx context.Context, op, query string, args ...interface{}) ([]*domain.Session, error) {
	rows, err := r.adapter.Select(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		var sessionRow SessionRow
		err := rows.Scan(
			&sessionRow.ID,
			&sessionRow.UserID,
			&sessionRow.TokenHash,
			&sessionRow.AuthMethod,
			&sessionRow.IPAddress,
			&sessionRow.UserAgent,
			&sessionRow.DeviceID,
			&sessionRow.Status,
			&sessionRow.CreatedAt,
			&sessionRow.ExpiresAt,
			&sessionRow.LastActiveAt,
			&sessionRow.RevokedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		sessions = append(sessions, sessionRow.ToDomainSession())
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	return sessions, nil
}
