package v1

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
)

// ============================================================================
// Verification Response
// ============================================================================

// VerificationResponse is the response for a single verification.
type VerificationResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Provider       string     `json:"provider"`
	ProviderName   string     `json:"provider_name"`
	CredentialType string     `json:"credential_type"`
	Status         string     `json:"status"`
	ProviderUserID string     `json:"provider_user_id,omitempty"`
	Username       string     `json:"username,omitempty"`
	CredentialID   string     `json:"credential_id,omitempty"`
	FailureReason  string     `json:"failure_reason,omitempty"`
	FailureCode    string     `json:"failure_code,omitempty"`
	InitiatedAt    time.Time  `json:"initiated_at"`
	AuthorizedAt   *time.Time `json:"authorized_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	FailedAt       *time.Time `json:"failed_at,omitempty"`
	ExpiresAt      time.Time  `json:"expires_at"`
}

// FromVerificationView converts a domain.VerificationView to a VerificationResponse.
func FromVerificationView(view *domain.VerificationView) *VerificationResponse {
	if view == nil {
		return nil
	}

	return &VerificationResponse{
		ID:             view.ID,
		UserID:         view.UserID,
		Provider:       view.Provider,
		ProviderName:   view.ProviderName,
		CredentialType: view.CredentialType,
		Status:         view.Status,
		ProviderUserID: view.ProviderUserID,
		Username:       view.Username,
		CredentialID:   view.CredentialID,
		FailureReason:  view.FailureReason,
		FailureCode:    view.FailureCode,
		InitiatedAt:    view.InitiatedAt,
		AuthorizedAt:   view.AuthorizedAt,
		CompletedAt:    view.CompletedAt,
		FailedAt:       view.FailedAt,
		ExpiresAt:      view.ExpiresAt,
	}
}

// ============================================================================
// Verification Summary Response
// ============================================================================

// VerificationSummaryResponse is a lightweight response for lists.
type VerificationSummaryResponse struct {
	ID           string     `json:"id"`
	Provider     string     `json:"provider"`
	ProviderName string     `json:"provider_name"`
	Status       string     `json:"status"`
	InitiatedAt  time.Time  `json:"initiated_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// FromVerificationSummary converts a domain.VerificationSummary to a VerificationSummaryResponse.
func FromVerificationSummary(summary *domain.VerificationSummary) *VerificationSummaryResponse {
	if summary == nil {
		return nil
	}

	return &VerificationSummaryResponse{
		ID:           summary.ID,
		Provider:     summary.Provider,
		ProviderName: summary.ProviderName,
		Status:       summary.Status,
		InitiatedAt:  summary.InitiatedAt,
		CompletedAt:  summary.CompletedAt,
	}
}

// ============================================================================
// Verification List Response
// ============================================================================

// VerificationListResponse is the response for a list of verifications.
type VerificationListResponse struct {
	Verifications []*VerificationSummaryResponse `json:"verifications"`
	Total         int                            `json:"total"`
	Limit         int                            `json:"limit"`
	Offset        int                            `json:"offset"`
	HasMore       bool                           `json:"has_more"`
}

// FromVerificationSummaries converts a slice of domain.VerificationSummary to a VerificationListResponse.
func FromVerificationSummaries(summaries []*domain.VerificationSummary, total, limit, offset int) *VerificationListResponse {
	verifications := make([]*VerificationSummaryResponse, len(summaries))
	for i, summary := range summaries {
		verifications[i] = FromVerificationSummary(summary)
	}

	return &VerificationListResponse{
		Verifications: verifications,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
		HasMore:       offset+len(summaries) < total,
	}
}

// ============================================================================
// Initiate Verification Response
// ============================================================================

// InitiateVerificationResponse is returned when a verification is initiated.
type InitiateVerificationResponse struct {
	VerificationID   string `json:"verification_id"`
	AuthorizationURL string `json:"authorization_url"`
	ExpiresIn        int64  `json:"expires_in"`
	Message          string `json:"message"`
}

// NewInitiateVerificationResponse creates a new InitiateVerificationResponse.
func NewInitiateVerificationResponse(verificationID, authURL string, expiresIn int64) *InitiateVerificationResponse {
	return &InitiateVerificationResponse{
		VerificationID:   verificationID,
		AuthorizationURL: authURL,
		ExpiresIn:        expiresIn,
		Message:          "Verification initiated. Redirect user to authorization URL.",
	}
}

// ============================================================================
// OAuth Callback Response
// ============================================================================

// OAuthCallbackResponse is returned after processing an OAuth callback.
type OAuthCallbackResponse struct {
	VerificationID string `json:"verification_id"`
	Status         string `json:"status"`
	RedirectURL    string `json:"redirect_url,omitempty"`
	Message        string `json:"message"`
}

// NewOAuthCallbackSuccessResponse creates a success response for OAuth callback.
func NewOAuthCallbackSuccessResponse(verificationID, redirectURL string) *OAuthCallbackResponse {
	return &OAuthCallbackResponse{
		VerificationID: verificationID,
		Status:         "authorized",
		RedirectURL:    redirectURL,
		Message:        "Authorization successful. Processing verification.",
	}
}

// NewOAuthCallbackErrorResponse creates an error response for OAuth callback.
func NewOAuthCallbackErrorResponse(verificationID, errorMsg string) *OAuthCallbackResponse {
	return &OAuthCallbackResponse{
		VerificationID: verificationID,
		Status:         "failed",
		Message:        errorMsg,
	}
}

// ============================================================================
// Verification Complete Response
// ============================================================================

// VerificationCompleteResponse is returned when a verification is completed.
type VerificationCompleteResponse struct {
	VerificationID string `json:"verification_id"`
	CredentialID   string `json:"credential_id"`
	CredentialType string `json:"credential_type"`
	Provider       string `json:"provider"`
	Username       string `json:"username,omitempty"`
	Message        string `json:"message"`
}

// NewVerificationCompleteResponse creates a new VerificationCompleteResponse.
func NewVerificationCompleteResponse(verificationID, credentialID, credentialType, provider, username string) *VerificationCompleteResponse {
	return &VerificationCompleteResponse{
		VerificationID: verificationID,
		CredentialID:   credentialID,
		CredentialType: credentialType,
		Provider:       provider,
		Username:       username,
		Message:        "Verification completed successfully. Credential issued.",
	}
}

// ============================================================================
// Verification Cancelled Response
// ============================================================================

// VerificationCancelledResponse is returned when a verification is cancelled.
type VerificationCancelledResponse struct {
	VerificationID string `json:"verification_id"`
	Cancelled      bool   `json:"cancelled"`
	Message        string `json:"message"`
}

// NewVerificationCancelledResponse creates a new VerificationCancelledResponse.
func NewVerificationCancelledResponse(verificationID string) *VerificationCancelledResponse {
	return &VerificationCancelledResponse{
		VerificationID: verificationID,
		Cancelled:      true,
		Message:        "Verification cancelled successfully.",
	}
}

// ============================================================================
// Provider Connection Response
// ============================================================================

// ProviderConnectionResponse represents a user's connection status with a provider.
type ProviderConnectionResponse struct {
	Provider       string     `json:"provider"`
	ProviderName   string     `json:"provider_name"`
	Connected      bool       `json:"connected"`
	ProviderUserID string     `json:"provider_user_id,omitempty"`
	Username       string     `json:"username,omitempty"`
	CredentialID   string     `json:"credential_id,omitempty"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
}

// FromProviderConnectionView converts a domain.ProviderConnectionView to a ProviderConnectionResponse.
func FromProviderConnectionView(view *domain.ProviderConnectionView) *ProviderConnectionResponse {
	if view == nil {
		return nil
	}

	return &ProviderConnectionResponse{
		Provider:       view.Provider,
		ProviderName:   view.ProviderName,
		Connected:      view.Connected,
		ProviderUserID: view.ProviderUserID,
		Username:       view.Username,
		CredentialID:   view.CredentialID,
		VerifiedAt:     view.VerifiedAt,
	}
}

// ============================================================================
// Provider Connections Response
// ============================================================================

// ProviderConnectionsResponse is the response for all provider connections.
type ProviderConnectionsResponse struct {
	Connections []*ProviderConnectionResponse `json:"connections"`
}

// FromProviderConnectionViews converts a slice of domain.ProviderConnectionView to a ProviderConnectionsResponse.
func FromProviderConnectionViews(views []*domain.ProviderConnectionView) *ProviderConnectionsResponse {
	connections := make([]*ProviderConnectionResponse, len(views))
	for i, view := range views {
		connections[i] = FromProviderConnectionView(view)
	}

	return &ProviderConnectionsResponse{
		Connections: connections,
	}
}

// ============================================================================
// Check Provider Connected Response
// ============================================================================

// CheckProviderConnectedResponse is the response for checking if a provider is connected.
type CheckProviderConnectedResponse struct {
	Provider     string `json:"provider"`
	Connected    bool   `json:"connected"`
	CredentialID string `json:"credential_id,omitempty"`
	Username     string `json:"username,omitempty"`
}

// NewCheckProviderConnectedResponse creates a new CheckProviderConnectedResponse.
func NewCheckProviderConnectedResponse(provider string, connected bool, credentialID, username *string) *CheckProviderConnectedResponse {
	resp := &CheckProviderConnectedResponse{
		Provider:  provider,
		Connected: connected,
	}

	if credentialID != nil {
		resp.CredentialID = *credentialID
	}
	if username != nil {
		resp.Username = *username
	}

	return resp
}
