package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
)

// IssuerLookup provides read-side queries on the issuers projection table.
type IssuerLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewIssuerLookup creates a new IssuerLookup.
func NewIssuerLookup(pool *pgxpool.Pool) *IssuerLookup {
	return &IssuerLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns an issuer projection by ID.
func (l *IssuerLookup) GetByID(ctx context.Context, id string) (*IssuerProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetIssuerByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := IssuerRowToProjection(row)
	return &proj, nil
}

// GetByOrganizationID returns an issuer projection by organization ID.
func (l *IssuerLookup) GetByOrganizationID(ctx context.Context, orgID string) (*IssuerProjection, error) {
	row, err := l.queries.GetIssuerByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	proj := IssuerRowToProjection(row)
	return &proj, nil
}

// ListIssuers returns issuers with optional status filter and pagination.
func (l *IssuerLookup) ListIssuers(ctx context.Context, status *string, limit, offset int) ([]IssuerProjection, error) {
	rows, err := l.queries.ListIssuers(ctx, generated.ListIssuersParams{
		Status: status,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]IssuerProjection, len(rows))
	for i, row := range rows {
		projections[i] = IssuerRowToProjection(row)
	}
	return projections, nil
}

// CountIssuers returns the count of issuers with optional status filter.
func (l *IssuerLookup) CountIssuers(ctx context.Context, status *string) (int, error) {
	count, err := l.queries.CountIssuers(ctx, status)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
