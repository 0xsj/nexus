package v1

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/credential/application/query"
	"github.com/0xsj/nexus/platform/internal/credential/domain"
)

// ============================================================================
// Credential Response
// ============================================================================

// CredentialResponse is the response for a single credential.
type CredentialResponse struct {
	ID             string         `json:"id"`
	CredentialType string         `json:"credential_type"`
	SchemaID       string         `json:"schema_id,omitempty"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	Status         string         `json:"status"`
	Claims         map[string]any `json:"claims,omitempty"`
	HasSignedVC    bool           `json:"has_signed_vc"`
	IssuedAt       *time.Time     `json:"issued_at,omitempty"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	RevokedAt      *time.Time     `json:"revoked_at,omitempty"`
	SuspendedAt    *time.Time     `json:"suspended_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Version        int            `json:"version"`
}

// FromCredentialView converts a query.CredentialView to a CredentialResponse.
func FromCredentialView(view *query.CredentialView) *CredentialResponse {
	if view == nil {
		return nil
	}

	return &CredentialResponse{
		ID:             view.ID,
		CredentialType: view.CredentialType,
		SchemaID:       view.SchemaID,
		HolderDID:      view.HolderDID,
		IssuerDID:      view.IssuerDID,
		Status:         view.Status,
		Claims:         view.Claims,
		HasSignedVC:    view.HasSignedVC(),
		IssuedAt:       view.IssuedAt,
		ExpiresAt:      view.ExpiresAt,
		RevokedAt:      view.RevokedAt,
		SuspendedAt:    view.SuspendedAt,
		CreatedAt:      view.CreatedAt,
		UpdatedAt:      view.UpdatedAt,
		Version:        view.Version,
	}
}

// ============================================================================
// Credential List Response
// ============================================================================

// CredentialListResponse is the response for a list of credentials.
type CredentialListResponse struct {
	Credentials []*CredentialResponse `json:"credentials"`
	Total       int                   `json:"total"`
	Limit       int                   `json:"limit"`
	Offset      int                   `json:"offset"`
	HasMore     bool                  `json:"has_more"`
}

// FromCredentialListResult converts a query.CredentialListResult to a CredentialListResponse.
func FromCredentialListResult(result *query.CredentialListResult) *CredentialListResponse {
	if result == nil {
		return &CredentialListResponse{
			Credentials: []*CredentialResponse{},
		}
	}

	credentials := make([]*CredentialResponse, len(result.Credentials))
	for i, view := range result.Credentials {
		credentials[i] = FromCredentialView(view)
	}

	return &CredentialListResponse{
		Credentials: credentials,
		Total:       result.Total,
		Limit:       result.Limit,
		Offset:      result.Offset,
		HasMore:     result.HasMore,
	}
}

// ============================================================================
// Verifiable Credential Response
// ============================================================================

// VerifiableCredentialResponse is the response for a signed JWT-VC.
type VerifiableCredentialResponse struct {
	ID       string `json:"id"`
	JWT      string `json:"jwt"`
	Format   string `json:"format"`
	IssuedAt string `json:"issued_at,omitempty"`
}

// NewVerifiableCredentialResponse creates a new VerifiableCredentialResponse.
func NewVerifiableCredentialResponse(id string, jwt string, issuedAt *time.Time) *VerifiableCredentialResponse {
	resp := &VerifiableCredentialResponse{
		ID:     id,
		JWT:    jwt,
		Format: "jwt_vc",
	}

	if issuedAt != nil {
		resp.IssuedAt = issuedAt.Format(time.RFC3339)
	}

	return resp
}

// ============================================================================
// Verification Response
// ============================================================================

// VerificationResponse is the response for credential verification.
type VerificationResponse struct {
	Valid           bool                       `json:"valid"`
	Issuer          string                     `json:"issuer,omitempty"`
	Holder          string                     `json:"holder,omitempty"`
	CredentialID    string                     `json:"credential_id,omitempty"`
	CredentialTypes []string                   `json:"credential_types,omitempty"`
	Claims          map[string]any             `json:"claims,omitempty"`
	IssuedAt        *time.Time                 `json:"issued_at,omitempty"`
	ExpiresAt       *time.Time                 `json:"expires_at,omitempty"`
	Checks          VerificationChecksResponse `json:"checks"`
	Error           string                     `json:"error,omitempty"`
}

// VerificationChecksResponse contains the result of individual verification checks.
type VerificationChecksResponse struct {
	Signature  string `json:"signature"`
	Expiration string `json:"expiration"`
	NotBefore  string `json:"not_before"`
	IssuerDID  string `json:"issuer_did"`
	HolderDID  string `json:"holder_did"`
}

// FromVerificationResult converts a domain.VerificationResult to a VerificationResponse.
func FromVerificationResult(result *domain.VerificationResult) *VerificationResponse {
	if result == nil {
		return nil
	}

	return &VerificationResponse{
		Valid:           result.Valid,
		Issuer:          result.Issuer,
		Holder:          result.Holder,
		CredentialID:    result.CredentialID,
		CredentialTypes: result.CredentialTypes,
		Claims:          result.Claims,
		IssuedAt:        result.IssuedAt,
		ExpiresAt:       result.ExpiresAt,
		Checks: VerificationChecksResponse{
			Signature:  result.Checks.Signature.String(),
			Expiration: result.Checks.Expiration.String(),
			NotBefore:  result.Checks.NotBefore.String(),
			IssuerDID:  result.Checks.IssuerDID.String(),
			HolderDID:  result.Checks.HolderDID.String(),
		},
		Error: result.Error,
	}
}

// ============================================================================
// Command Result Responses
// ============================================================================

// CredentialCreatedResponse is returned when a credential is created.
type CredentialCreatedResponse struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Message string `json:"message"`
}

// NewCredentialCreatedResponse creates a new CredentialCreatedResponse.
func NewCredentialCreatedResponse(id string, version int) *CredentialCreatedResponse {
	return &CredentialCreatedResponse{
		ID:      id,
		Version: version,
		Message: "Credential created successfully",
	}
}

// CredentialIssuedResponse is returned when a credential is issued.
type CredentialIssuedResponse struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Message string `json:"message"`
}

// NewCredentialIssuedResponse creates a new CredentialIssuedResponse.
func NewCredentialIssuedResponse(id string, version int) *CredentialIssuedResponse {
	return &CredentialIssuedResponse{
		ID:      id,
		Version: version,
		Message: "Credential issued successfully",
	}
}

// CredentialRevokedResponse is returned when a credential is revoked.
type CredentialRevokedResponse struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Message string `json:"message"`
}

// NewCredentialRevokedResponse creates a new CredentialRevokedResponse.
func NewCredentialRevokedResponse(id string, version int) *CredentialRevokedResponse {
	return &CredentialRevokedResponse{
		ID:      id,
		Version: version,
		Message: "Credential revoked successfully",
	}
}

// CredentialSuspendedResponse is returned when a credential is suspended.
type CredentialSuspendedResponse struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Message string `json:"message"`
}

// NewCredentialSuspendedResponse creates a new CredentialSuspendedResponse.
func NewCredentialSuspendedResponse(id string, version int) *CredentialSuspendedResponse {
	return &CredentialSuspendedResponse{
		ID:      id,
		Version: version,
		Message: "Credential suspended successfully",
	}
}

// CredentialReinstatedResponse is returned when a credential is reinstated.
type CredentialReinstatedResponse struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Message string `json:"message"`
}

// NewCredentialReinstatedResponse creates a new CredentialReinstatedResponse.
func NewCredentialReinstatedResponse(id string, version int) *CredentialReinstatedResponse {
	return &CredentialReinstatedResponse{
		ID:      id,
		Version: version,
		Message: "Credential reinstated successfully",
	}
}
