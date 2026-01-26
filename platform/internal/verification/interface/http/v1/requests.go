package v1

import (
	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/http/request"
)

// ============================================================================
// Initiate Verification Request
// ============================================================================

// InitiateVerificationRequest is the request body for initiating a verification.
type InitiateVerificationRequest struct {
	Provider       string `json:"provider"`
	CredentialType string `json:"credential_type"`
	RedirectURL    string `json:"redirect_url,omitempty"`
}

// Validate validates the request.
func (r *InitiateVerificationRequest) Validate() error {
	v := request.NewValidator()

	v.Required("provider", r.Provider)
	v.OneOf("provider", r.Provider, validProviders())

	v.Required("credential_type", r.CredentialType)
	v.OneOf("credential_type", r.CredentialType, validCredentialTypes())

	if r.RedirectURL != "" {
		v.URL("redirect_url", r.RedirectURL)
	}

	return v.Error()
}

// ToProvider converts the provider string to domain.Provider.
func (r *InitiateVerificationRequest) ToProvider() domain.Provider {
	return domain.Provider(r.Provider)
}

// ToCredentialType converts the credential type string to domain.CredentialType.
func (r *InitiateVerificationRequest) ToCredentialType() domain.CredentialType {
	return domain.CredentialType(r.CredentialType)
}

// ============================================================================
// OAuth Callback Request
// ============================================================================

// OAuthCallbackRequest represents the OAuth callback query parameters.
type OAuthCallbackRequest struct {
	State string `json:"state"`
	Code  string `json:"code"`
	Error string `json:"error,omitempty"`
}

// Validate validates the request.
func (r *OAuthCallbackRequest) Validate() error {
	v := request.NewValidator()

	v.Required("state", r.State)

	// Either code or error must be present
	if r.Code == "" && r.Error == "" {
		v.AddError("code", "either code or error must be present")
	}

	return v.Error()
}

// ============================================================================
// Cancel Verification Request
// ============================================================================

// CancelVerificationRequest is the request body for cancelling a verification.
type CancelVerificationRequest struct {
	Reason string `json:"reason,omitempty"`
}

// Validate validates the request.
func (r *CancelVerificationRequest) Validate() error {
	v := request.NewValidator()

	if r.Reason != "" {
		v.MaxLength("reason", r.Reason, 500)
	}

	return v.Error()
}

// ============================================================================
// List Verifications Query Params
// ============================================================================

// ListVerificationsParams holds query parameters for listing verifications.
type ListVerificationsParams struct {
	Provider string
	Status   string
	Limit    int
	Offset   int
	SortBy   string
	SortDesc bool
}

// DefaultListVerificationsParams returns default list parameters.
func DefaultListVerificationsParams() ListVerificationsParams {
	return ListVerificationsParams{
		Limit:    20,
		Offset:   0,
		SortBy:   "initiated_at",
		SortDesc: true,
	}
}

// Validate validates the parameters.
func (p *ListVerificationsParams) Validate() error {
	v := request.NewValidator()

	if p.Provider != "" {
		v.OneOf("provider", p.Provider, validProviders())
	}

	if p.Status != "" {
		v.OneOf("status", p.Status, validStatuses())
	}

	v.Between("limit", p.Limit, 1, 100)
	v.NonNegative("offset", p.Offset)

	if p.SortBy != "" {
		v.OneOf("sort_by", p.SortBy, []string{"initiated_at", "completed_at", "provider", "status"})
	}

	return v.Error()
}

// ToProviderFilter converts provider string to domain.Provider pointer.
func (p *ListVerificationsParams) ToProviderFilter() *domain.Provider {
	if p.Provider == "" {
		return nil
	}
	provider := domain.Provider(p.Provider)
	return &provider
}

// ToStatusFilter converts status string to domain.VerificationStatus pointer.
func (p *ListVerificationsParams) ToStatusFilter() *domain.VerificationStatus {
	if p.Status == "" {
		return nil
	}
	status := domain.VerificationStatus(p.Status)
	return &status
}

// ============================================================================
// Check Provider Connection Request
// ============================================================================

// CheckProviderConnectionParams holds query parameters for checking provider connection.
type CheckProviderConnectionParams struct {
	Provider string
}

// Validate validates the parameters.
func (p *CheckProviderConnectionParams) Validate() error {
	v := request.NewValidator()

	v.Required("provider", p.Provider)
	v.OneOf("provider", p.Provider, validProviders())

	return v.Error()
}

// ============================================================================
// Validation Helpers
// ============================================================================

// validProviders returns valid provider values.
func validProviders() []string {
	return []string{
		string(domain.ProviderGitHub),
		string(domain.ProviderLinkedIn),
		string(domain.ProviderGoogle),
		string(domain.ProviderTwitter),
		string(domain.ProviderDiscord),
	}
}

// validCredentialTypes returns valid credential type values.
func validCredentialTypes() []string {
	return []string{
		string(domain.CredentialTypeGitHubContributor),
		string(domain.CredentialTypeLinkedInEmployment),
	}
}

// validStatuses returns valid verification status values.
func validStatuses() []string {
	return []string{
		string(domain.StatusPending),
		string(domain.StatusAuthorized),
		string(domain.StatusFetching),
		string(domain.StatusCompleted),
		string(domain.StatusFailed),
		string(domain.StatusExpired),
	}
}
