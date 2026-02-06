package domain

import (
	"context"
)

// ============================================================================
// Notification Repository
// ============================================================================

// NotificationRepository defines the persistence operations for Notification aggregates.
type NotificationRepository interface {
	// Save persists a notification aggregate (appends new events).
	Save(ctx context.Context, notification *Notification) error

	// Get retrieves a notification by ID (replays events to rebuild state).
	Get(ctx context.Context, id NotificationID) (*Notification, error)

	// Exists checks if a notification with the given ID exists.
	Exists(ctx context.Context, id NotificationID) (bool, error)
}

// ============================================================================
// Preferences Repository
// ============================================================================

// PreferencesRepository defines the persistence operations for NotificationPreferences.
type PreferencesRepository interface {
	// SavePreferences persists notification preferences for a user.
	SavePreferences(ctx context.Context, prefs NotificationPreferences) error

	// GetPreferences retrieves notification preferences for a user.
	GetPreferences(ctx context.Context, userID string) (NotificationPreferences, error)
}
