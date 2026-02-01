package query

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Compile-time interface checks
// ============================================================================

var (
	_ cqrs.Query = (*GetEntry)(nil)
	_ cqrs.Query = (*GetActivity)(nil)
	_ cqrs.Query = (*GetUserActivity)(nil)
	_ cqrs.Query = (*GetCredentialHistory)(nil)
	_ cqrs.Query = (*GetSubjectHistory)(nil)
	_ cqrs.Query = (*GetVerificationLog)(nil)
	_ cqrs.Query = (*GetActivityStats)(nil)
)

// ============================================================================
// Query Names
// ============================================================================

const (
	QueryNameGetEntry             = "ledger.GetEntry"
	QueryNameGetActivity          = "ledger.GetActivity"
	QueryNameGetUserActivity      = "ledger.GetUserActivity"
	QueryNameGetCredentialHistory = "ledger.GetCredentialHistory"
	QueryNameGetSubjectHistory    = "ledger.GetSubjectHistory"
	QueryNameGetVerificationLog   = "ledger.GetVerificationLog"
	QueryNameGetActivityStats     = "ledger.GetActivityStats"
)

// ============================================================================
// Get Entry Query
// ============================================================================

// GetEntry retrieves a single audit entry by ID.
type GetEntry struct {
	ID string `json:"id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetEntry) QueryName() string {
	return QueryNameGetEntry
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

// QueryName implements cqrs.Query.
func (q GetActivity) QueryName() string {
	return QueryNameGetActivity
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

// QueryName implements cqrs.Query.
func (q GetUserActivity) QueryName() string {
	return QueryNameGetUserActivity
}

// ============================================================================
// History Queries
// ============================================================================

// GetCredentialHistory retrieves the full lifecycle history of a credential.
type GetCredentialHistory struct {
	CredentialID string `json:"credential_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetCredentialHistory) QueryName() string {
	return QueryNameGetCredentialHistory
}

// GetSubjectHistory retrieves the history of any subject.
type GetSubjectHistory struct {
	SubjectID   string `json:"subject_id" validate:"required"`
	SubjectType string `json:"subject_type" validate:"required"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
}

// QueryName implements cqrs.Query.
func (q GetSubjectHistory) QueryName() string {
	return QueryNameGetSubjectHistory
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

// QueryName implements cqrs.Query.
func (q GetVerificationLog) QueryName() string {
	return QueryNameGetVerificationLog
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

// QueryName implements cqrs.Query.
func (q GetActivityStats) QueryName() string {
	return QueryNameGetActivityStats
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
