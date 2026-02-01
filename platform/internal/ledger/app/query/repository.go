package query

import (
	"context"
	"time"
)

// ============================================================================
// Reader Interface
// ============================================================================

// Reader defines the interface for reading audit entries.
type Reader interface {
	// GetByID retrieves a single entry by ID.
	GetByID(ctx context.Context, id string) (*EntryView, error)

	// List retrieves paginated entries with filters.
	List(ctx context.Context, filter ListFilter) (*ActivityView, error)

	// GetBySubject retrieves entries for a specific subject.
	GetBySubject(ctx context.Context, filter SubjectFilter) (*SubjectHistoryView, error)

	// GetByActor retrieves entries for a specific actor.
	GetByActor(ctx context.Context, filter ActorFilter) (*ActivityView, error)

	// GetVerificationLog retrieves verification attempts for a user.
	GetVerificationLog(ctx context.Context, filter VerificationLogFilter) (*VerificationLogView, error)

	// GetStats retrieves aggregated statistics.
	GetStats(ctx context.Context, filter StatsFilter) (*ActivityStatsView, error)
}

// ============================================================================
// Filters
// ============================================================================

// ListFilter contains parameters for listing entries.
type ListFilter struct {
	EventTypes  []string
	ActorID     string
	ActorType   string
	SubjectID   string
	SubjectType string
	ContextID   string
	FromTime    time.Time
	ToTime      time.Time
	Page        int
	PageSize    int
}

// SubjectFilter contains parameters for filtering by subject.
type SubjectFilter struct {
	SubjectID   string
	SubjectType string
	EventTypes  []string
	FromTime    time.Time
	ToTime      time.Time
	Page        int
	PageSize    int
}

// ActorFilter contains parameters for filtering by actor.
type ActorFilter struct {
	ActorID    string
	ActorType  string
	EventTypes []string
	FromTime   time.Time
	ToTime     time.Time
	Page       int
	PageSize   int
}

// VerificationLogFilter contains parameters for filtering verification logs.
type VerificationLogFilter struct {
	UserID   string
	FromTime time.Time
	ToTime   time.Time
	Page     int
	PageSize int
}

// StatsFilter contains parameters for statistics queries.
type StatsFilter struct {
	ActorID     string
	SubjectID   string
	SubjectType string
	FromTime    time.Time
	ToTime      time.Time
}
