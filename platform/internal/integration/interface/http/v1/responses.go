package v1

import "time"

// ============================================================================
// Integration Responses
// ============================================================================

// IntegrationResponse represents a full integration in API responses.
type IntegrationResponse struct {
	IntegrationID    string            `json:"integration_id"`
	UserID           string            `json:"user_id"`
	ProviderType     string            `json:"provider_type"`
	Status           string            `json:"status"`
	ProviderUserID   string            `json:"provider_user_id,omitempty"`
	ProviderUsername string            `json:"provider_username,omitempty"`
	Scopes           []string          `json:"scopes,omitempty"`
	LastFetchAt      *time.Time        `json:"last_fetch_at,omitempty"`
	FetchCount       int               `json:"fetch_count"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	ConnectedAt      *time.Time        `json:"connected_at,omitempty"`
	DisconnectedAt   *time.Time        `json:"disconnected_at,omitempty"`
	SuspendedAt      *time.Time        `json:"suspended_at,omitempty"`
	SuspensionReason string            `json:"suspension_reason,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// IntegrationSummaryResponse represents a lightweight integration in list responses.
type IntegrationSummaryResponse struct {
	IntegrationID    string     `json:"integration_id"`
	UserID           string     `json:"user_id"`
	ProviderType     string     `json:"provider_type"`
	Status           string     `json:"status"`
	ProviderUsername string     `json:"provider_username,omitempty"`
	ConnectedAt      *time.Time `json:"connected_at,omitempty"`
}

// IntegrationListResponse represents a list of integrations.
type IntegrationListResponse struct {
	Integrations []IntegrationSummaryResponse `json:"integrations"`
	TotalCount   int                          `json:"total_count"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// ConnectProviderResponse represents the result of connecting a provider.
type ConnectProviderResponse struct {
	IntegrationID string    `json:"integration_id"`
	ProviderType  string    `json:"provider_type"`
	Status        string    `json:"status"`
	ConnectedAt   time.Time `json:"connected_at"`
}

// DisconnectProviderResponse represents the result of disconnecting a provider.
type DisconnectProviderResponse struct {
	IntegrationID  string    `json:"integration_id"`
	Status         string    `json:"status"`
	DisconnectedAt time.Time `json:"disconnected_at"`
}

// RefreshCredentialsResponse represents the result of refreshing credentials.
type RefreshCredentialsResponse struct {
	IntegrationID string    `json:"integration_id"`
	Status        string    `json:"status"`
	RefreshedAt   time.Time `json:"refreshed_at"`
}

// SuspendIntegrationResponse represents the result of suspending an integration.
type SuspendIntegrationResponse struct {
	IntegrationID string    `json:"integration_id"`
	Status        string    `json:"status"`
	SuspendedAt   time.Time `json:"suspended_at"`
}

// ============================================================================
// Supported Providers Response
// ============================================================================

// SupportedProviderResponse describes a single supported provider.
type SupportedProviderResponse struct {
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// SupportedProvidersResponse represents the list of all supported providers.
type SupportedProvidersResponse struct {
	Providers []SupportedProviderResponse `json:"providers"`
}
