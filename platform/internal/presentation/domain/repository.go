package domain

import (
	"context"
)

// ============================================================================
// Presentation Repository
// ============================================================================

// PresentationRepository defines the persistence operations for Presentation aggregates.
type PresentationRepository interface {
	// Save persists a presentation aggregate (appends new events).
	Save(ctx context.Context, presentation *Presentation) error

	// Get retrieves a presentation by ID (replays events to rebuild state).
	Get(ctx context.Context, id PresentationID) (*Presentation, error)

	// Exists checks if a presentation with the given ID exists.
	Exists(ctx context.Context, id PresentationID) (bool, error)
}

// ============================================================================
// ShareLink Repository
// ============================================================================

// ShareLinkRepository defines the persistence operations for ShareLink aggregates.
type ShareLinkRepository interface {
	// Save persists a share link aggregate (appends new events).
	Save(ctx context.Context, shareLink *ShareLink) error

	// Get retrieves a share link by ID (replays events to rebuild state).
	Get(ctx context.Context, id ShareLinkID) (*ShareLink, error)

	// Exists checks if a share link with the given ID exists.
	Exists(ctx context.Context, id ShareLinkID) (bool, error)
}
