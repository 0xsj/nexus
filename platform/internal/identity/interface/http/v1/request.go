package v1

import (
	"github.com/0xsj/nexus/platform/internal/identity/domain"
)

// ============================================================================
// Auth Requests
// ============================================================================

// ChallengeRequest is the request to generate a SIWE challenge.
type ChallengeRequest struct {
	Address string `json:"address" validate:"required"`
	Chain   string `json:"chain" validate:"required,oneof=ethereum polygon arbitrum optimism base"`
}

// ToChain converts the chain string to domain.Chain.
func (r *ChallengeRequest) ToChain() domain.Chain {
	return domain.Chain(r.Chain)
}

// RegisterWalletRequest is the request to register with a wallet.
type RegisterWalletRequest struct {
	Address   string `json:"address" validate:"required"`
	Chain     string `json:"chain" validate:"required,oneof=ethereum polygon arbitrum optimism base"`
	Signature string `json:"signature" validate:"required"`
	Message   string `json:"message" validate:"required"`
	Nonce     string `json:"nonce" validate:"required"`
}

// ToChain converts the chain string to domain.Chain.
func (r *RegisterWalletRequest) ToChain() domain.Chain {
	return domain.Chain(r.Chain)
}

// LoginWalletRequest is the request to login with a wallet.
type LoginWalletRequest struct {
	Address   string `json:"address" validate:"required"`
	Chain     string `json:"chain" validate:"required,oneof=ethereum polygon arbitrum optimism base"`
	Signature string `json:"signature" validate:"required"`
	Message   string `json:"message" validate:"required"`
	Nonce     string `json:"nonce" validate:"required"`
}

// ToChain converts the chain string to domain.Chain.
func (r *LoginWalletRequest) ToChain() domain.Chain {
	return domain.Chain(r.Chain)
}

// RefreshTokenRequest is the request to refresh an access token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ============================================================================
// Session Requests
// ============================================================================

// RevokeSessionRequest is the request to revoke a session.
type RevokeSessionRequest struct {
	Reason string `json:"reason,omitempty"`
}

// RevokeAllSessionsRequest is the request to revoke all sessions.
type RevokeAllSessionsRequest struct {
	ExceptCurrent bool   `json:"except_current,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

// ============================================================================
// API Key Requests
// ============================================================================

// CreateAPIKeyRequest is the request to create an API key.
type CreateAPIKeyRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Scopes      []string `json:"scopes" validate:"required,min=1"`
	ExpiresIn   *int64   `json:"expires_in,omitempty"` // Seconds, nil = default
	Description string   `json:"description,omitempty" validate:"max=500"`
}

// ToScopes converts string scopes to domain.APIKeyScopes.
func (r *CreateAPIKeyRequest) ToScopes() domain.APIKeyScopes {
	return domain.ParseAPIKeyScopes(r.Scopes)
}

// RevokeAPIKeyRequest is the request to revoke an API key.
type RevokeAPIKeyRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ============================================================================
// User Requests
// ============================================================================

// LinkWalletRequest is the request to link a wallet.
type LinkWalletRequest struct {
	Address   string `json:"address" validate:"required"`
	Chain     string `json:"chain" validate:"required,oneof=ethereum polygon arbitrum optimism base"`
	Signature string `json:"signature" validate:"required"`
	Message   string `json:"message" validate:"required"`
	Nonce     string `json:"nonce" validate:"required"`
}

// ToChain converts the chain string to domain.Chain.
func (r *LinkWalletRequest) ToChain() domain.Chain {
	return domain.Chain(r.Chain)
}

// UnlinkWalletRequest is the request to unlink a wallet.
type UnlinkWalletRequest struct {
	Address string `json:"address" validate:"required"`
	Chain   string `json:"chain" validate:"required,oneof=ethereum polygon arbitrum optimism base"`
}

// ToChain converts the chain string to domain.Chain.
func (r *UnlinkWalletRequest) ToChain() domain.Chain {
	return domain.Chain(r.Chain)
}

// ============================================================================
// Query Requests (URL Parameters)
// ============================================================================

// ListParams contains common list query parameters.
type ListParams struct {
	Limit     int    `query:"limit" validate:"min=1,max=100"`
	Offset    int    `query:"offset" validate:"min=0"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// WithDefaults applies default values.
func (p *ListParams) WithDefaults() {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.SortOrder == "" {
		p.SortOrder = "desc"
	}
}

// SessionListParams contains session list query parameters.
type SessionListParams struct {
	ListParams
	ActiveOnly bool `query:"active_only"`
}

// APIKeyListParams contains API key list query parameters.
type APIKeyListParams struct {
	ListParams
	ActiveOnly bool `query:"active_only"`
}
