package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/integration/infrastructure/persistence/postgres/generated"
)

// IntegrationLookup provides read-side queries on the integrations projection table.
type IntegrationLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewIntegrationLookup creates a new IntegrationLookup.
func NewIntegrationLookup(pool *pgxpool.Pool) *IntegrationLookup {
	return &IntegrationLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns an integration projection by ID.
func (l *IntegrationLookup) GetByID(ctx context.Context, id string) (*IntegrationProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetIntegrationByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := IntegrationRowToProjection(row)
	return &proj, nil
}

// ListByUserID returns integration projections for a user.
func (l *IntegrationLookup) ListByUserID(ctx context.Context, userID string) ([]IntegrationProjection, error) {
	rows, err := l.queries.ListIntegrationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	projections := make([]IntegrationProjection, len(rows))
	for i, row := range rows {
		projections[i] = IntegrationRowToProjection(row)
	}
	return projections, nil
}

// CountByUserID returns the count of integrations for a user.
func (l *IntegrationLookup) CountByUserID(ctx context.Context, userID string) (int, error) {
	count, err := l.queries.CountIntegrationsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetByUserAndProvider returns an integration projection for a user and provider type.
func (l *IntegrationLookup) GetByUserAndProvider(ctx context.Context, userID string, providerType string) (*IntegrationProjection, error) {
	row, err := l.queries.GetIntegrationByUserAndProvider(ctx, generated.GetIntegrationByUserAndProviderParams{
		UserID:       userID,
		ProviderType: providerType,
	})
	if err != nil {
		return nil, err
	}

	proj := IntegrationRowToProjection(row)
	return &proj, nil
}
