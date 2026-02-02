package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/persistence/postgres/generated"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// Ensure OAuthStateRepository implements domain.OAuthStateRepository.
var _ domain.OAuthStateRepository = (*OAuthStateRepository)(nil)

// OAuthStateRepository is a PostgreSQL implementation of domain.OAuthStateRepository.
// OAuth states are short-lived CSRF tokens and not event-sourced.
type OAuthStateRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewOAuthStateRepository creates a new PostgreSQL OAuth state repository.
func NewOAuthStateRepository(pool *pgxpool.Pool) *OAuthStateRepository {
	return &OAuthStateRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save stores an OAuth state record.
func (r *OAuthStateRepository) Save(ctx context.Context, record *domain.OAuthStateRecord) error {
	params := OAuthStateRecordToInsertParams(record)
	if err := r.queries.InsertOAuthState(ctx, params); err != nil {
		return pkgerrors.Infrastructure("OAuthStateRepository.Save", err).
			WithMessage("failed to save OAuth state")
	}
	return nil
}

// GetByState retrieves an OAuth state by its state value.
func (r *OAuthStateRepository) GetByState(ctx context.Context, state string) (*domain.OAuthStateRecord, error) {
	row, err := r.queries.GetOAuthStateByState(ctx, state)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.OAuthStateMismatch("OAuthStateRepository.GetByState")
		}
		return nil, pkgerrors.Infrastructure("OAuthStateRepository.GetByState", err).
			WithMessage("failed to retrieve OAuth state")
	}

	return OAuthStateRowToRecord(row), nil
}

// Delete removes an OAuth state (after use or expiration).
func (r *OAuthStateRepository) Delete(ctx context.Context, state string) error {
	if err := r.queries.DeleteOAuthState(ctx, state); err != nil {
		return pkgerrors.Infrastructure("OAuthStateRepository.Delete", err).
			WithMessage("failed to delete OAuth state")
	}
	return nil
}

// DeleteExpired removes expired OAuth states.
func (r *OAuthStateRepository) DeleteExpired(ctx context.Context) (int64, error) {
	count, err := r.queries.DeleteExpiredOAuthStates(ctx)
	if err != nil {
		return 0, pkgerrors.Infrastructure("OAuthStateRepository.DeleteExpired", err).
			WithMessage("failed to delete expired OAuth states")
	}
	return count, nil
}
