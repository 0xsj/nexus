package v1

import (
	"time"
)

// ============================================================================
// Activity Requests
// ============================================================================

// GetActivityRequest represents a request to list audit entries.
type GetActivityRequest struct {
	EventTypes  []string  `json:"event_types,omitempty"`
	ActorID     string    `json:"actor_id,omitempty"`
	ActorType   string    `json:"actor_type,omitempty"`
	SubjectID   string    `json:"subject_id,omitempty"`
	SubjectType string    `json:"subject_type,omitempty"`
	ContextID   string    `json:"context_id,omitempty"`
	FromTime    time.Time `json:"from_time,omitempty"`
	ToTime      time.Time `json:"to_time,omitempty"`
	Page        int       `json:"page,omitempty"`
	PageSize    int       `json:"page_size,omitempty"`
}

// GetUserActivityRequest represents a request to get activity for a specific user.
type GetUserActivityRequest struct {
	UserID     string    `json:"user_id" validate:"required"`
	EventTypes []string  `json:"event_types,omitempty"`
	FromTime   time.Time `json:"from_time,omitempty"`
	ToTime     time.Time `json:"to_time,omitempty"`
	Page       int       `json:"page,omitempty"`
	PageSize   int       `json:"page_size,omitempty"`
}

// ============================================================================
// History Requests
// ============================================================================

// GetCredentialHistoryRequest represents a request to get credential history.
type GetCredentialHistoryRequest struct {
	CredentialID string `json:"credential_id" validate:"required"`
}

// GetSubjectHistoryRequest represents a request to get subject history.
type GetSubjectHistoryRequest struct {
	SubjectID   string `json:"subject_id" validate:"required"`
	SubjectType string `json:"subject_type" validate:"required"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
}

// ============================================================================
// Verification Log Request
// ============================================================================

// GetVerificationLogRequest represents a request to get verification logs.
type GetVerificationLogRequest struct {
	UserID   string    `json:"user_id" validate:"required"`
	FromTime time.Time `json:"from_time,omitempty"`
	ToTime   time.Time `json:"to_time,omitempty"`
	Page     int       `json:"page,omitempty"`
	PageSize int       `json:"page_size,omitempty"`
}

// ============================================================================
// Statistics Request
// ============================================================================

// GetActivityStatsRequest represents a request to get activity statistics.
type GetActivityStatsRequest struct {
	ActorID     string    `json:"actor_id,omitempty"`
	SubjectID   string    `json:"subject_id,omitempty"`
	SubjectType string    `json:"subject_type,omitempty"`
	FromTime    time.Time `json:"from_time,omitempty"`
	ToTime      time.Time `json:"to_time,omitempty"`
}
