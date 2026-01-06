package v1

import (
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
)

// ============================================================================
// Challenge Requests
// ============================================================================

// CreateChallengeRequest is the request to generate a SIWE challenge.
type CreateChallengeRequest struct {
	Address string `json:"address" validate:"required"`
	ChainID string `json:"chain_id" validate:"required"`
}

// ToChainID converts the chain ID string to domain.ChainID.
func (r *CreateChallengeRequest) ToChainID() domain.ChainID {
	return domain.ChainID(r.ChainID)
}

// VerifyChallengeRequest is the request to verify a signed challenge.
type VerifyChallengeRequest struct {
	Nonce     string `json:"nonce" validate:"required"`
	Signature string `json:"signature" validate:"required"`
	Message   string `json:"message" validate:"required"`
}

// ============================================================================
// Wallet Link Requests
// ============================================================================

// LinkWalletRequest is the request to link a wallet to a user.
type LinkWalletRequest struct {
	Address   string  `json:"address" validate:"required"`
	ChainID   string  `json:"chain_id" validate:"required"`
	Signature string  `json:"signature" validate:"required"`
	Message   string  `json:"message" validate:"required"`
	Nonce     string  `json:"nonce" validate:"required"`
	Label     *string `json:"label,omitempty" validate:"omitempty,max=128"`
	IsPrimary bool    `json:"is_primary,omitempty"`
}

// ToChainID converts the chain ID string to domain.ChainID.
func (r *LinkWalletRequest) ToChainID() domain.ChainID {
	return domain.ChainID(r.ChainID)
}

// GetLabel returns the label or empty string.
func (r *LinkWalletRequest) GetLabel() string {
	if r.Label != nil {
		return *r.Label
	}
	return ""
}

// ReverifyWalletRequest is the request to re-verify wallet ownership.
type ReverifyWalletRequest struct {
	Signature string `json:"signature" validate:"required"`
	Message   string `json:"message" validate:"required"`
	Nonce     string `json:"nonce" validate:"required"`
}

// ============================================================================
// Wallet Update Requests
// ============================================================================

// UpdateWalletRequest is the request to update wallet properties.
type UpdateWalletRequest struct {
	Label     *string `json:"label,omitempty" validate:"omitempty,max=128"`
	IsPrimary *bool   `json:"is_primary,omitempty"`
}

// HasUpdates returns true if any field is set.
func (r *UpdateWalletRequest) HasUpdates() bool {
	return r.Label != nil || r.IsPrimary != nil
}

// SetPrimaryWalletRequest is the request to set a wallet as primary.
type SetPrimaryWalletRequest struct {
	// Empty body - wallet ID comes from URL
}

// ============================================================================
// Wallet Status Requests
// ============================================================================

// ActivateWalletRequest is the request to activate a wallet.
type ActivateWalletRequest struct {
	// Empty body - wallet ID comes from URL
}

// DeactivateWalletRequest is the request to deactivate a wallet.
type DeactivateWalletRequest struct {
	Reason string `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// ============================================================================
// Query Parameters
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

// WalletListParams contains wallet list query parameters.
type WalletListParams struct {
	ListParams
	ChainID    string `query:"chain_id"`
	Status     string `query:"status" validate:"omitempty,oneof=active inactive suspended"`
	ActiveOnly bool   `query:"active_only"`
}

// ToChainID converts the chain ID string to domain.ChainID.
func (p *WalletListParams) ToChainID() domain.ChainID {
	if p.ChainID == "" {
		return ""
	}
	return domain.ChainID(p.ChainID)
}

// ToStatus converts the status string to domain.WalletStatus.
func (p *WalletListParams) ToStatus() *domain.WalletStatus {
	if p.Status == "" {
		return nil
	}
	status := domain.WalletStatus(p.Status)
	return &status
}

// ============================================================================
// Request Validation
// ============================================================================

// ValidateCreateChallengeRequest validates a create challenge request.
func ValidateCreateChallengeRequest(req *CreateChallengeRequest) error {
	if req.Address == "" {
		return validationError("address is required")
	}
	if req.ChainID == "" {
		return validationError("chain_id is required")
	}
	return nil
}

// ValidateVerifyChallengeRequest validates a verify challenge request.
func ValidateVerifyChallengeRequest(req *VerifyChallengeRequest) error {
	if req.Nonce == "" {
		return validationError("nonce is required")
	}
	if req.Signature == "" {
		return validationError("signature is required")
	}
	if req.Message == "" {
		return validationError("message is required")
	}
	return nil
}

// ValidateLinkWalletRequest validates a link wallet request.
func ValidateLinkWalletRequest(req *LinkWalletRequest) error {
	if req.Address == "" {
		return validationError("address is required")
	}
	if req.ChainID == "" {
		return validationError("chain_id is required")
	}
	if req.Signature == "" {
		return validationError("signature is required")
	}
	if req.Message == "" {
		return validationError("message is required")
	}
	if req.Nonce == "" {
		return validationError("nonce is required")
	}
	if req.Label != nil && len(*req.Label) > 128 {
		return validationError("label must be 128 characters or less")
	}
	return nil
}

// ValidateReverifyWalletRequest validates a reverify wallet request.
func ValidateReverifyWalletRequest(req *ReverifyWalletRequest) error {
	if req.Signature == "" {
		return validationError("signature is required")
	}
	if req.Message == "" {
		return validationError("message is required")
	}
	if req.Nonce == "" {
		return validationError("nonce is required")
	}
	return nil
}

// ValidateUpdateWalletRequest validates an update wallet request.
func ValidateUpdateWalletRequest(req *UpdateWalletRequest) error {
	if !req.HasUpdates() {
		return validationError("at least one field must be provided")
	}
	if req.Label != nil && len(*req.Label) > 128 {
		return validationError("label must be 128 characters or less")
	}
	return nil
}

// ValidateDeactivateWalletRequest validates a deactivate wallet request.
func ValidateDeactivateWalletRequest(req *DeactivateWalletRequest) error {
	if len(req.Reason) > 256 {
		return validationError("reason must be 256 characters or less")
	}
	return nil
}
