package v1

import "time"

// ============================================================================
// Issuer Responses
// ============================================================================

// IssuerResponse represents a full issuer in API responses.
type IssuerResponse struct {
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

// IssuerSummaryResponse represents a lightweight issuer in list responses.
type IssuerSummaryResponse struct {
	IssuerID       string    `json:"issuer_id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// IssuerListResponse represents a paginated list of issuers.
type IssuerListResponse struct {
	Issuers    []IssuerSummaryResponse `json:"issuers"`
	TotalCount int                     `json:"total_count"`
	Limit      int                     `json:"limit"`
	Offset     int                     `json:"offset"`
	HasMore    bool                    `json:"has_more"`
}

// ============================================================================
// Template Responses
// ============================================================================

// TemplateResponse represents a full template in API responses.
type TemplateResponse struct {
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

// TemplateSummaryResponse represents a lightweight template in list responses.
type TemplateSummaryResponse struct {
	TemplateID string    `json:"template_id"`
	IssuerID   string    `json:"issuer_id"`
	Name       string    `json:"name"`
	SchemaType string    `json:"schema_type"`
	Status     string    `json:"status"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
}

// TemplateListResponse represents a paginated list of templates.
type TemplateListResponse struct {
	Templates  []TemplateSummaryResponse `json:"templates"`
	TotalCount int                       `json:"total_count"`
	Limit      int                       `json:"limit"`
	Offset     int                       `json:"offset"`
	HasMore    bool                      `json:"has_more"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// IssuerRegisteredResponse represents the result of registering an issuer.
type IssuerRegisteredResponse struct {
	IssuerID  string    `json:"issuer_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// IssuerActivatedResponse represents the result of activating an issuer.
type IssuerActivatedResponse struct {
	IssuerID    string    `json:"issuer_id"`
	Status      string    `json:"status"`
	ActivatedAt time.Time `json:"activated_at"`
}

// IssuerSuspendedResponse represents the result of suspending an issuer.
type IssuerSuspendedResponse struct {
	IssuerID    string    `json:"issuer_id"`
	Status      string    `json:"status"`
	SuspendedAt time.Time `json:"suspended_at"`
}

// TemplateCreatedResponse represents the result of creating a template.
type TemplateCreatedResponse struct {
	TemplateID string    `json:"template_id"`
	IssuerID   string    `json:"issuer_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// TemplateUpdatedResponse represents the result of updating a template.
type TemplateUpdatedResponse struct {
	TemplateID string    `json:"template_id"`
	Version    int       `json:"version"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TemplateArchivedResponse represents the result of archiving a template.
type TemplateArchivedResponse struct {
	TemplateID string    `json:"template_id"`
	Status     string    `json:"status"`
	ArchivedAt time.Time `json:"archived_at"`
}
