package v1

// ============================================================================
// Issuer Requests
// ============================================================================

// RegisterIssuerRequest represents a request to register a new issuer.
type RegisterIssuerRequest struct {
	OrganizationID string `json:"organization_id" validate:"required"`
	Name           string `json:"name" validate:"required"`
	Description    string `json:"description,omitempty"`
	WebhookURL     string `json:"webhook_url,omitempty"`
}

// ActivateIssuerRequest represents a request to activate an issuer.
type ActivateIssuerRequest struct {
	DID               string `json:"did" validate:"required"`
	APIKey            string `json:"api_key" validate:"required"`
	LogoURL           string `json:"logo_url,omitempty"`
	PrimaryColor      string `json:"primary_color,omitempty"`
	SecondaryColor    string `json:"secondary_color,omitempty"`
	CertificateDesign string `json:"certificate_design,omitempty"`
}

// SuspendIssuerRequest represents a request to suspend an issuer.
type SuspendIssuerRequest struct {
	Reason string `json:"reason" validate:"required,max=500"`
}

// ============================================================================
// Template Requests
// ============================================================================

// CreateTemplateRequest represents a request to create a new template.
type CreateTemplateRequest struct {
	Name           string            `json:"name" validate:"required"`
	Description    string            `json:"description,omitempty"`
	SchemaType     string            `json:"schema_type" validate:"required"`
	ClaimMappings  map[string]string `json:"claim_mappings,omitempty"`
	DefaultValues  map[string]any    `json:"default_values,omitempty"`
	ExpirationDays int               `json:"expiration_days,omitempty"`
	AutoApprove    bool              `json:"auto_approve,omitempty"`
}

// UpdateTemplateRequest represents a request to update a template.
type UpdateTemplateRequest struct {
	Name           string            `json:"name" validate:"required"`
	Description    string            `json:"description,omitempty"`
	ClaimMappings  map[string]string `json:"claim_mappings,omitempty"`
	DefaultValues  map[string]any    `json:"default_values,omitempty"`
	ExpirationDays int               `json:"expiration_days,omitempty"`
	AutoApprove    bool              `json:"auto_approve,omitempty"`
}
