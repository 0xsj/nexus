package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/issuer/domain"
	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure TemplateRepository implements domain.TemplateRepository.
var _ domain.TemplateRepository = (*TemplateRepository)(nil)

// TemplateRepository is a PostgreSQL implementation of domain.TemplateRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type TemplateRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewTemplateRepository creates a new PostgreSQL template repository.
func NewTemplateRepository(pool *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a template aggregate by appending new events and updating the projection.
func (r *TemplateRepository) Save(ctx context.Context, t *domain.Template) error {
	if !t.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("TemplateRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestTemplateEventVersion(ctx, uuidFromTemplateID(t.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("TemplateRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := t.Version() - len(t.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"TemplateRepository.Save",
			t.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range t.Changes() {
		version := expectedVersion + i + 1
		params, err := TemplateEventToInsertParams(t.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("TemplateRepository.Save", err)
		}

		if err := qtx.InsertTemplateEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("TemplateRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := TemplateToUpsertParams(t)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("TemplateRepository.Save", "template", err)
	}

	if err := qtx.UpsertTemplate(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("TemplateRepository.Save", "template", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("TemplateRepository.Save", err)
	}

	t.ClearChanges()
	return nil
}

// Get retrieves a template by ID, replaying events to rebuild state.
func (r *TemplateRepository) Get(ctx context.Context, id domain.TemplateID) (*domain.Template, error) {
	rows, err := r.queries.GetTemplateEvents(ctx, uuidFromTemplateID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("TemplateRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("TemplateRepository.Get", domain.AggregateTypeTemplate, id.String())
	}

	events, err := TemplateEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("TemplateRepository.Get", id.String(), err)
	}

	t := domain.NewTemplateFromEvents(id.String())
	eventsourcing.Hydrate(t, events)

	return t, nil
}

// Exists checks if a template with the given ID exists.
func (r *TemplateRepository) Exists(ctx context.Context, id domain.TemplateID) (bool, error) {
	exists, err := r.queries.TemplateAggregateExists(ctx, uuidFromTemplateID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("TemplateRepository.Exists", id.String(), err)
	}
	return exists, nil
}
