package v1

import (
	"time"
)

// ============================================================================
// Entry Response
// ============================================================================

// EntryResponse represents a single audit entry in API responses.
type EntryResponse struct {
	ID          string         `json:"id"`
	OccurredAt  time.Time      `json:"occurred_at"`
	RecordedAt  time.Time      `json:"recorded_at"`
	EventType   string         `json:"event_type"`
	ActorID     string         `json:"actor_id"`
	ActorType   string         `json:"actor_type"`
	SubjectID   string         `json:"subject_id"`
	SubjectType string         `json:"subject_type"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	ContextID   string         `json:"context_id,omitempty"`
}

// ============================================================================
// Activity Responses
// ============================================================================

// ActivityResponse represents a paginated list of audit entries.
type ActivityResponse struct {
	Entries    []EntryResponse `json:"entries"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	HasMore    bool            `json:"has_more"`
}

// UserActivityResponse represents activity for a specific user.
type UserActivityResponse struct {
	UserID   string           `json:"user_id"`
	Activity ActivityResponse `json:"activity"`
}

// ============================================================================
// History Responses
// ============================================================================

// CredentialHistoryResponse represents the lifecycle history of a credential.
type CredentialHistoryResponse struct {
	CredentialID string          `json:"credential_id"`
	Entries      []EntryResponse `json:"entries"`
	TotalCount   int             `json:"total_count"`
}

// SubjectHistoryResponse represents the history of any subject.
type SubjectHistoryResponse struct {
	SubjectID   string          `json:"subject_id"`
	SubjectType string          `json:"subject_type"`
	Entries     []EntryResponse `json:"entries"`
	TotalCount  int             `json:"total_count"`
}

// ============================================================================
// Verification Log Response
// ============================================================================

// VerificationLogEntryResponse represents a single verification attempt.
type VerificationLogEntryResponse struct {
	ID           string         `json:"id"`
	OccurredAt   time.Time      `json:"occurred_at"`
	VerifierID   string         `json:"verifier_id"`
	VerifierType string         `json:"verifier_type"`
	CredentialID string         `json:"credential_id"`
	Outcome      string         `json:"outcome"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// VerificationLogResponse represents a log of verification attempts.
type VerificationLogResponse struct {
	UserID     string                         `json:"user_id"`
	Entries    []VerificationLogEntryResponse `json:"entries"`
	TotalCount int                            `json:"total_count"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
	HasMore    bool                           `json:"has_more"`
}

// ============================================================================
// Statistics Response
// ============================================================================

// EventTypeCountResponse represents a count for a specific event type.
type EventTypeCountResponse struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
}

// ActivityStatsResponse represents aggregated activity statistics.
type ActivityStatsResponse struct {
	TotalEntries    int                      `json:"total_entries"`
	EventTypeCounts []EventTypeCountResponse `json:"event_type_counts"`
	FromTime        time.Time                `json:"from_time,omitempty"`
	ToTime          time.Time                `json:"to_time,omitempty"`
}
