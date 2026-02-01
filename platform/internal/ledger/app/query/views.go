package query

import (
	"time"
)

// ============================================================================
// Entry View
// ============================================================================

// EntryView represents an audit entry for read operations.
type EntryView struct {
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
// Activity Views
// ============================================================================

// ActivityView represents a paginated list of audit entries.
type ActivityView struct {
	Entries    []EntryView `json:"entries"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	HasMore    bool        `json:"has_more"`
}

// UserActivityView represents activity for a specific user.
type UserActivityView struct {
	UserID   string       `json:"user_id"`
	Activity ActivityView `json:"activity"`
}

// ============================================================================
// Credential History View
// ============================================================================

// CredentialHistoryView represents the lifecycle history of a credential.
type CredentialHistoryView struct {
	CredentialID string      `json:"credential_id"`
	Entries      []EntryView `json:"entries"`
	TotalCount   int         `json:"total_count"`
}

// ============================================================================
// Subject History View
// ============================================================================

// SubjectHistoryView represents the history of any subject.
type SubjectHistoryView struct {
	SubjectID   string      `json:"subject_id"`
	SubjectType string      `json:"subject_type"`
	Entries     []EntryView `json:"entries"`
	TotalCount  int         `json:"total_count"`
}

// ============================================================================
// Verification Log View
// ============================================================================

// VerificationLogEntry represents a single verification attempt.
type VerificationLogEntry struct {
	ID           string         `json:"id"`
	OccurredAt   time.Time      `json:"occurred_at"`
	VerifierID   string         `json:"verifier_id"`
	VerifierType string         `json:"verifier_type"`
	CredentialID string         `json:"credential_id"`
	Outcome      string         `json:"outcome"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// VerificationLogView represents a log of external verification attempts.
type VerificationLogView struct {
	UserID     string                 `json:"user_id"`
	Entries    []VerificationLogEntry `json:"entries"`
	TotalCount int                    `json:"total_count"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	HasMore    bool                   `json:"has_more"`
}

// ============================================================================
// Statistics View
// ============================================================================

// EventTypeCount represents a count for a specific event type.
type EventTypeCount struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
}

// ActivityStatsView represents aggregated activity statistics.
type ActivityStatsView struct {
	TotalEntries    int              `json:"total_entries"`
	EventTypeCounts []EventTypeCount `json:"event_type_counts"`
	FromTime        time.Time        `json:"from_time"`
	ToTime          time.Time        `json:"to_time"`
}
