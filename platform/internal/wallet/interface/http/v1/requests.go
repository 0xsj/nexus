package v1

// ============================================================================
// Wallet Requests
// ============================================================================

// LinkWalletRequest represents a request to link a new wallet.
type LinkWalletRequest struct {
	Address string `json:"address" validate:"required"`
	ChainID int    `json:"chain_id" validate:"required"`
	Label   string `json:"label,omitempty"`
}

// VerifyWalletRequest represents a request to verify wallet ownership.
type VerifyWalletRequest struct {
	Signature string `json:"signature" validate:"required"`
	Message   string `json:"message" validate:"required"`
}

// UpdateWalletLabelRequest represents a request to update a wallet's label.
type UpdateWalletLabelRequest struct {
	Label string `json:"label" validate:"required"`
}
