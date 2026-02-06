package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/notification/domain"
	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/persistence/postgres/generated"
)

// Ensure PreferencesRepository implements domain.PreferencesRepository.
var _ domain.PreferencesRepository = (*PreferencesRepository)(nil)

// PreferencesRepository is a PostgreSQL implementation of domain.PreferencesRepository.
// Preferences are stored using simple CRUD operations (not event-sourced).
type PreferencesRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewPreferencesRepository creates a new PostgreSQL preferences repository.
func NewPreferencesRepository(pool *pgxpool.Pool) *PreferencesRepository {
	return &PreferencesRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// SavePreferences persists notification preferences for a user.
func (r *PreferencesRepository) SavePreferences(ctx context.Context, prefs domain.NotificationPreferences) error {
	params, err := PreferencesToUpsertParams(prefs)
	if err != nil {
		return err
	}
	return r.queries.UpsertPreferences(ctx, params)
}

// GetPreferences retrieves notification preferences for a user.
// Returns default preferences if none are found.
func (r *PreferencesRepository) GetPreferences(ctx context.Context, userID string) (domain.NotificationPreferences, error) {
	row, err := r.queries.GetPreferencesByUserID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.NewDefaultPreferences(userID), nil
		}
		return domain.NotificationPreferences{}, err
	}

	return PreferencesFromRow(row)
}
