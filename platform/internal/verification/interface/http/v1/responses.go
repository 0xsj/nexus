package v1

import "time"

// ============================================================================
// Verification Responses
// ============================================================================

// VerificationResponse represents a full verification in API responses.
type VerificationResponse struct {
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

// VerificationSummaryResponse represents a lightweight verification in list responses.
type VerificationSummaryResponse struct {
	VerificationID string     `json:"verification_id"`
	UserID         string     `json:"user_id"`
	ProviderType   string     `json:"provider_type"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// VerificationListResponse represents a paginated list of verifications.
type VerificationListResponse struct {
	Verifications []VerificationSummaryResponse `json:"verifications"`
	TotalCount    int                           `json:"total_count"`
	Limit         int                           `json:"limit"`
	Offset        int                           `json:"offset"`
	HasMore       bool                          `json:"has_more"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// VerificationStartedResponse represents the result of starting a verification.
type VerificationStartedResponse struct {
	VerificationID string    `json:"verification_id"`
	ProviderType   string    `json:"provider_type"`
	Status         string    `json:"status"`
	OAuthState     string    `json:"oauth_state"`
	StartedAt      time.Time `json:"started_at"`
}

// OAuthCallbackResponse represents the result of receiving an OAuth callback.
type OAuthCallbackResponse struct {
	VerificationID string `json:"verification_id"`
	Status         string `json:"status"`
}
