package command

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandLinkWallet        = "wallet.LinkWallet"
	CommandVerifyWallet      = "wallet.VerifyWallet"
	CommandUnlinkWallet      = "wallet.UnlinkWallet"
	CommandSetPrimaryWallet  = "wallet.SetPrimaryWallet"
	CommandUpdateWalletLabel = "wallet.UpdateWalletLabel"
)

// ============================================================================
// LinkWallet
// ============================================================================

// LinkWallet links a new wallet to a user.
type LinkWallet struct {
	UserID  string `json:"user_id" validate:"required"`
	Address string `json:"address" validate:"required"`
	ChainID int    `json:"chain_id" validate:"required"`
	Label   string `json:"label" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c LinkWallet) CommandName() string {
	return CommandLinkWallet
}

// Validate implements cqrs.Validatable.
func (c LinkWallet) Validate() error {
	if c.UserID == "" {
		return cqrs.ErrCommandValidation("LinkWallet.Validate", "user_id is required")
	}
	if c.Address == "" {
		return cqrs.ErrCommandValidation("LinkWallet.Validate", "address is required")
	}
	if c.ChainID <= 0 {
		return cqrs.ErrCommandValidation("LinkWallet.Validate", "chain_id must be positive")
	}
	return nil
}

// LinkWalletResult is the result data for LinkWallet.
type LinkWalletResult struct {
	WalletID string `json:"wallet_id"`
	Address  string `json:"address"`
	Status   string `json:"status"`
}

// ============================================================================
// VerifyWallet
// ============================================================================

// VerifyWallet verifies wallet ownership via SIWE signature.
type VerifyWallet struct {
	WalletID  types.ID `json:"wallet_id" validate:"required"`
	Signature string   `json:"signature" validate:"required"`
	Message   string   `json:"message" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c VerifyWallet) CommandName() string {
	return CommandVerifyWallet
}

// Validate implements cqrs.Validatable.
func (c VerifyWallet) Validate() error {
	if c.WalletID.IsZero() {
		return cqrs.ErrCommandValidation("VerifyWallet.Validate", "wallet_id is required")
	}
	if c.Signature == "" {
		return cqrs.ErrCommandValidation("VerifyWallet.Validate", "signature is required")
	}
	if c.Message == "" {
		return cqrs.ErrCommandValidation("VerifyWallet.Validate", "message is required")
	}
	return nil
}

// VerifyWalletResult is the result data for VerifyWallet.
type VerifyWalletResult struct {
	WalletID string `json:"wallet_id"`
	Status   string `json:"status"`
}

// ============================================================================
// UnlinkWallet
// ============================================================================

// UnlinkWallet unlinks a wallet from a user.
type UnlinkWallet struct {
	WalletID types.ID `json:"wallet_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c UnlinkWallet) CommandName() string {
	return CommandUnlinkWallet
}

// Validate implements cqrs.Validatable.
func (c UnlinkWallet) Validate() error {
	if c.WalletID.IsZero() {
		return cqrs.ErrCommandValidation("UnlinkWallet.Validate", "wallet_id is required")
	}
	return nil
}

// UnlinkWalletResult is the result data for UnlinkWallet.
type UnlinkWalletResult struct {
	WalletID string `json:"wallet_id"`
	Status   string `json:"status"`
}

// ============================================================================
// SetPrimaryWallet
// ============================================================================

// SetPrimaryWallet designates a wallet as the user's primary wallet.
type SetPrimaryWallet struct {
	WalletID types.ID `json:"wallet_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c SetPrimaryWallet) CommandName() string {
	return CommandSetPrimaryWallet
}

// Validate implements cqrs.Validatable.
func (c SetPrimaryWallet) Validate() error {
	if c.WalletID.IsZero() {
		return cqrs.ErrCommandValidation("SetPrimaryWallet.Validate", "wallet_id is required")
	}
	return nil
}

// SetPrimaryWalletResult is the result data for SetPrimaryWallet.
type SetPrimaryWalletResult struct {
	WalletID string `json:"wallet_id"`
	Status   string `json:"status"`
}

// ============================================================================
// UpdateWalletLabel
// ============================================================================

// UpdateWalletLabel updates the label on a wallet.
type UpdateWalletLabel struct {
	WalletID types.ID `json:"wallet_id" validate:"required"`
	Label    string   `json:"label" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c UpdateWalletLabel) CommandName() string {
	return CommandUpdateWalletLabel
}

// Validate implements cqrs.Validatable.
func (c UpdateWalletLabel) Validate() error {
	if c.WalletID.IsZero() {
		return cqrs.ErrCommandValidation("UpdateWalletLabel.Validate", "wallet_id is required")
	}
	if c.Label == "" {
		return cqrs.ErrCommandValidation("UpdateWalletLabel.Validate", "label is required")
	}
	return nil
}

// UpdateWalletLabelResult is the result data for UpdateWalletLabel.
type UpdateWalletLabelResult struct {
	WalletID string `json:"wallet_id"`
	Label    string `json:"label"`
}
