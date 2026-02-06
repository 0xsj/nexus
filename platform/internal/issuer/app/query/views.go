package query

import "time"

// ============================================================================
// Issuer Views
// ============================================================================

// IssuerView is the full issuer read model.
type IssuerView struct {
	IssuerID       string         `json:"issuer_id"`
	OrganizationID string         `json:"organization_id"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	DID            string         `json:"did,omitempty"`
	WebhookURL     string         `json:"webhook_url,omitempty"`
	Status         string         `json:"status"`
	Branding       map[string]any `json:"branding,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// IssuerSummaryView is a lightweight issuer representation for lists.
type IssuerSummaryView struct {
	IssuerID       string    `json:"issuer_id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// IssuerListView is a paginated list of issuers.
type IssuerListView struct {
	Issuers    []IssuerSummaryView `json:"issuers"`
	TotalCount int                 `json:"total_count"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	HasMore    bool                `json:"has_more"`
}

// ============================================================================
// Template Views
// ============================================================================

// TemplateView is the full template read model.
type TemplateView struct {
	TemplateID     string            `json:"template_id"`
	IssuerID       string            `json:"issuer_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	SchemaType     string            `json:"schema_type"`
	ClaimMappings  map[string]string `json:"claim_mappings,omitempty"`
	DefaultValues  map[string]any    `json:"default_values,omitempty"`
	ExpirationDays int               `json:"expiration_days"`
	AutoApprove    bool              `json:"auto_approve"`
	Status         string            `json:"status"`
	Version        int               `json:"version"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// TemplateSummaryView is a lightweight template representation for lists.
type TemplateSummaryView struct {
	TemplateID string    `json:"template_id"`
	IssuerID   string    `json:"issuer_id"`
	Name       string    `json:"name"`
	SchemaType string    `json:"schema_type"`
	Status     string    `json:"status"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
}

// TemplateListView is a paginated list of templates.
type TemplateListView struct {
	Templates  []TemplateSummaryView `json:"templates"`
	TotalCount int                   `json:"total_count"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
	HasMore    bool                  `json:"has_more"`
}
