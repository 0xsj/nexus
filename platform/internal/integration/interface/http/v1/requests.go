package v1

// ============================================================================
// Integration Requests
// ============================================================================

// ConnectProviderRequest represents a request to connect a provider.
type ConnectProviderRequest struct {
	UserID           string   `json:"user_id" validate:"required"`
	ProviderType     string   `json:"provider_type" validate:"required"`
	ProviderUserID   string   `json:"provider_user_id" validate:"required"`
	ProviderUsername string   `json:"provider_username,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
}

// SuspendIntegrationRequest represents a request to suspend an integration.
type SuspendIntegrationRequest struct {
	Reason string `json:"reason" validate:"required,max=500"`
}

// RefreshCredentialsRequest represents a request to refresh credentials.
type RefreshCredentialsRequest struct {
	Scopes []string `json:"scopes,omitempty"`
}
