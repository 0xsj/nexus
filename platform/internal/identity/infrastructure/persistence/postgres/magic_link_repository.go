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

// Ensure MagicLinkRepository implements domain.MagicLinkRepository.
var _ domain.MagicLinkRepository = (*MagicLinkRepository)(nil)

// MagicLinkRepository is a PostgreSQL implementation of domain.MagicLinkRepository.
// Magic links are short-lived tokens and not event-sourced.
type MagicLinkRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewMagicLinkRepository creates a new PostgreSQL magic link repository.
func NewMagicLinkRepository(pool *pgxpool.Pool) *MagicLinkRepository {
	return &MagicLinkRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save stores a magic link record.
func (r *MagicLinkRepository) Save(ctx context.Context, record *domain.MagicLinkRecord) error {
	params := MagicLinkRecordToInsertParams(record)
	if err := r.queries.InsertMagicLink(ctx, params); err != nil {
		return pkgerrors.Infrastructure("MagicLinkRepository.Save", err).
			WithMessage("failed to save magic link")
	}
	return nil
}

// GetByTokenHash retrieves a magic link by its token hash.
func (r *MagicLinkRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.MagicLinkRecord, error) {
	row, err := r.queries.GetMagicLinkByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.MagicLinkNotFoundError("MagicLinkRepository.GetByTokenHash")
		}
		return nil, pkgerrors.Infrastructure("MagicLinkRepository.GetByTokenHash", err).
			WithMessage("failed to retrieve magic link")
	}

	return MagicLinkRowToRecord(row), nil
}

// MarkUsed marks a magic link as used.
func (r *MagicLinkRepository) MarkUsed(ctx context.Context, tokenHash string) error {
	if err := r.queries.MarkMagicLinkUsed(ctx, tokenHash); err != nil {
		return pkgerrors.Infrastructure("MagicLinkRepository.MarkUsed", err).
			WithMessage("failed to mark magic link as used")
	}
	return nil
}

// DeleteExpired removes expired magic links.
func (r *MagicLinkRepository) DeleteExpired(ctx context.Context) (int64, error) {
	count, err := r.queries.DeleteExpiredMagicLinks(ctx)
	if err != nil {
		return 0, pkgerrors.Infrastructure("MagicLinkRepository.DeleteExpired", err).
			WithMessage("failed to delete expired magic links")
	}
	return count, nil
}
