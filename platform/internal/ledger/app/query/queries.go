package query

import (
	"time"
)

// ============================================================================
// Get Entry Query
// ============================================================================

// GetEntry retrieves a single audit entry by ID.
type GetEntry struct {
	ID string `json:"id" validate:"required"`
}

// ============================================================================
// Activity Queries
// ============================================================================

// GetActivity retrieves paginated audit entries with optional filters.
type GetActivity struct {
	// Filters
	EventTypes  []string  `json:"event_types,omitempty"`
	ActorID     string    `json:"actor_id,omitempty"`
	ActorType   string    `json:"actor_type,omitempty"`
	SubjectID   string    `json:"subject_id,omitempty"`
	SubjectType string    `json:"subject_type,omitempty"`
	ContextID   string    `json:"context_id,omitempty"`
	FromTime    time.Time `json:"from_time,omitempty"`
	ToTime      time.Time `json:"to_time,omitempty"`

	// Pagination
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// GetUserActivity retrieves activity for a specific user.
type GetUserActivity struct {
	UserID     string    `json:"user_id" validate:"required"`
	EventTypes []string  `json:"event_types,omitempty"`
	FromTime   time.Time `json:"from_time,omitempty"`
	ToTime     time.Time `json:"to_time,omitempty"`
	Page       int       `json:"page,omitempty"`
	PageSize   int       `json:"page_size,omitempty"`
}

// ============================================================================
// History Queries
// ============================================================================

// GetCredentialHistory retrieves the full lifecycle history of a credential.
type GetCredentialHistory struct {
	CredentialID string `json:"credential_id" validate:"required"`
}

// GetSubjectHistory retrieves the history of any subject.
type GetSubjectHistory struct {
	SubjectID   string `json:"subject_id" validate:"required"`
	SubjectType string `json:"subject_type" validate:"required"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
}

// ============================================================================
// Verification Log Query
// ============================================================================

// GetVerificationLog retrieves external verification attempts against a user's credentials.
type GetVerificationLog struct {
	UserID   string    `json:"user_id" validate:"required"`
	FromTime time.Time `json:"from_time,omitempty"`
	ToTime   time.Time `json:"to_time,omitempty"`
	Page     int       `json:"page,omitempty"`
	PageSize int       `json:"page_size,omitempty"`
}

// ============================================================================
// Statistics Query
// ============================================================================

// GetActivityStats retrieves aggregated activity statistics.
type GetActivityStats struct {
	// Filters
	ActorID     string `json:"actor_id,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`
	SubjectType string `json:"subject_type,omitempty"`

	// Time range
	FromTime time.Time `json:"from_time,omitempty"`
	ToTime   time.Time `json:"to_time,omitempty"`
}

// ============================================================================
// Pagination Defaults
// ============================================================================

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// NormalizePagination ensures page and page size have valid values.
func NormalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}
