package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
)

// TemplateLookup provides read-side queries on the templates projection table.
type TemplateLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewTemplateLookup creates a new TemplateLookup.
func NewTemplateLookup(pool *pgxpool.Pool) *TemplateLookup {
	return &TemplateLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a template projection by ID.
func (l *TemplateLookup) GetByID(ctx context.Context, id string) (*TemplateProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetTemplateByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := TemplateRowToProjection(row)
	return &proj, nil
}

// ListByIssuerID returns templates for an issuer with pagination.
func (l *TemplateLookup) ListByIssuerID(ctx context.Context, issuerID string, limit, offset int) ([]TemplateProjection, error) {
	uid, err := uuid.Parse(issuerID)
	if err != nil {
		return nil, err
	}

	rows, err := l.queries.ListTemplatesByIssuerID(ctx, generated.ListTemplatesByIssuerIDParams{
		IssuerID: uid,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]TemplateProjection, len(rows))
	for i, row := range rows {
		projections[i] = TemplateRowToProjection(row)
	}
	return projections, nil
}

// CountByIssuerID returns the count of templates for an issuer.
func (l *TemplateLookup) CountByIssuerID(ctx context.Context, issuerID string) (int, error) {
	uid, err := uuid.Parse(issuerID)
	if err != nil {
		return 0, err
	}

	count, err := l.queries.CountTemplatesByIssuerID(ctx, uid)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
