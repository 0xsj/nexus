package query

import "time"

// ============================================================================
// Integration Views
// ============================================================================

// IntegrationView is the full integration read model.
type IntegrationView struct {
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

// IntegrationSummaryView is a lightweight integration representation for lists.
type IntegrationSummaryView struct {
	IntegrationID    string     `json:"integration_id"`
	UserID           string     `json:"user_id"`
	ProviderType     string     `json:"provider_type"`
	Status           string     `json:"status"`
	ProviderUsername string     `json:"provider_username,omitempty"`
	ConnectedAt      *time.Time `json:"connected_at,omitempty"`
}

// IntegrationListView is a list of integrations for a user.
type IntegrationListView struct {
	Integrations []IntegrationSummaryView `json:"integrations"`
	TotalCount   int                      `json:"total_count"`
}

// SupportedProvidersView is a list of supported provider types.
type SupportedProvidersView struct {
	Providers []SupportedProviderView `json:"providers"`
}

// SupportedProviderView describes a single supported provider.
type SupportedProviderView struct {
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}
