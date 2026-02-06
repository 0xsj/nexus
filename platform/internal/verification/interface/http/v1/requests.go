package v1

// ============================================================================
// Verification Requests
// ============================================================================

// StartVerificationRequest represents a request to start a new verification.
type StartVerificationRequest struct {
	ProviderType string `json:"provider_type" validate:"required"`
}

// OAuthCallbackRequest represents an OAuth callback from a provider.
type OAuthCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
}
