package query

import "time"

// ============================================================================
// Verification Views
// ============================================================================

// VerificationView is the full verification read model.
type VerificationView struct {
	VerificationID string     `json:"verification_id"`
	UserID         string     `json:"user_id"`
	ProviderType   string     `json:"provider_type"`
	Status         string     `json:"status"`
	OAuthState     string     `json:"oauth_state"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	ErrorCode      string     `json:"error_code,omitempty"`
	CredentialID   string     `json:"credential_id,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// VerificationSummaryView is a lightweight verification representation for lists.
type VerificationSummaryView struct {
	VerificationID string     `json:"verification_id"`
	UserID         string     `json:"user_id"`
	ProviderType   string     `json:"provider_type"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// VerificationListView is a paginated list of verifications.
type VerificationListView struct {
	Verifications []VerificationSummaryView `json:"verifications"`
	TotalCount    int                       `json:"total_count"`
	Limit         int                       `json:"limit"`
	Offset        int                       `json:"offset"`
	HasMore       bool                      `json:"has_more"`
}
