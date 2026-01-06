package command

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Command Types
// ============================================================================

const (
	TypeLinkWallet        = "wallet.link"
	TypeUnlinkWallet      = "wallet.unlink"
	TypeUpdateLabel       = "wallet.update_label"
	TypeSetPrimary        = "wallet.set_primary"
	TypeActivateWallet    = "wallet.activate"
	TypeDeactivateWallet  = "wallet.deactivate"
	TypeSuspendWallet     = "wallet.suspend"
	TypeCreateChallenge   = "wallet.create_challenge"
	TypeVerifyChallenge   = "wallet.verify_challenge"
	TypeRecordWalletUsage = "wallet.record_usage"
)

// ============================================================================
// Link Wallet Command
// ============================================================================

// LinkWallet links a wallet to a user after signature verification.
type LinkWallet struct {
	UserID    string         `json:"user_id"`
	Address   string         `json:"address"`
	ChainID   domain.ChainID `json:"chain_id"`
	Signature string         `json:"signature"`
	Message   string         `json:"message"`
	Nonce     string         `json:"nonce"`
	Label     string         `json:"label,omitempty"`
}

// CommandName returns the command name.
func (c *LinkWallet) CommandName() string {
	return TypeLinkWallet
}

// LinkWalletResult is the result of linking a wallet.
type LinkWalletResult struct {
	WalletID  string    `json:"wallet_id"`
	UserID    string    `json:"user_id"`
	Address   string    `json:"address"`
	ChainID   string    `json:"chain_id"`
	DID       string    `json:"did"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// Unlink Wallet Command
// ============================================================================

// UnlinkWallet unlinks a wallet from a user.
type UnlinkWallet struct {
	UserID   string `json:"user_id"`
	WalletID string `json:"wallet_id"`
	Reason   string `json:"reason,omitempty"`
}

// CommandName returns the command name.
func (c *UnlinkWallet) CommandName() string {
	return TypeUnlinkWallet
}

// UnlinkWalletResult is the result of unlinking a wallet.
type UnlinkWalletResult struct {
	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id"`
	Address  string `json:"address"`
	ChainID  string `json:"chain_id"`
}

// ============================================================================
// Update Label Command
// ============================================================================

// UpdateLabel updates a wallet's display label.
type UpdateLabel struct {
	UserID   string `json:"user_id"`
	WalletID string `json:"wallet_id"`
	Label    string `json:"label"`
}

// CommandName returns the command name.
func (c *UpdateLabel) CommandName() string {
	return TypeUpdateLabel
}

// UpdateLabelResult is the result of updating a wallet label.
type UpdateLabelResult struct {
	WalletID string `json:"wallet_id"`
	OldLabel string `json:"old_label"`
	NewLabel string `json:"new_label"`
}

// ============================================================================
// Set Primary Command
// ============================================================================

// SetPrimary sets a wallet as the user's primary wallet.
type SetPrimary struct {
	UserID   string `json:"user_id"`
	WalletID string `json:"wallet_id"`
}

// CommandName returns the command name.
func (c *SetPrimary) CommandName() string {
	return TypeSetPrimary
}

// SetPrimaryResult is the result of setting a primary wallet.
type SetPrimaryResult struct {
	WalletID          string `json:"wallet_id"`
	PreviousPrimaryID string `json:"previous_primary_id,omitempty"`
}

// ============================================================================
// Activate Wallet Command
// ============================================================================

// ActivateWallet activates a deactivated wallet.
type ActivateWallet struct {
	UserID   string `json:"user_id"`
	WalletID string `json:"wallet_id"`
}

// CommandName returns the command name.
func (c *ActivateWallet) CommandName() string {
	return TypeActivateWallet
}

// ActivateWalletResult is the result of activating a wallet.
type ActivateWalletResult struct {
	WalletID       string `json:"wallet_id"`
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
}

// ============================================================================
// Deactivate Wallet Command
// ============================================================================

// DeactivateWallet deactivates a wallet.
type DeactivateWallet struct {
	UserID   string `json:"user_id"`
	WalletID string `json:"wallet_id"`
	Reason   string `json:"reason,omitempty"`
}

// CommandName returns the command name.
func (c *DeactivateWallet) CommandName() string {
	return TypeDeactivateWallet
}

// DeactivateWalletResult is the result of deactivating a wallet.
type DeactivateWalletResult struct {
	WalletID       string `json:"wallet_id"`
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
}

// ============================================================================
// Suspend Wallet Command
// ============================================================================

// SuspendWallet suspends a wallet (admin action).
type SuspendWallet struct {
	WalletID string `json:"wallet_id"`
	Reason   string `json:"reason"`
	AdminID  string `json:"admin_id"`
}

// CommandName returns the command name.
func (c *SuspendWallet) CommandName() string {
	return TypeSuspendWallet
}

// SuspendWalletResult is the result of suspending a wallet.
type SuspendWalletResult struct {
	WalletID       string `json:"wallet_id"`
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
	Reason         string `json:"reason"`
}

// ============================================================================
// Create Challenge Command
// ============================================================================

// CreateChallenge creates a SIWE challenge for wallet verification.
type CreateChallenge struct {
	Address   string         `json:"address"`
	ChainID   domain.ChainID `json:"chain_id"`
	Domain    string         `json:"domain"`
	URI       string         `json:"uri"`
	Statement string         `json:"statement,omitempty"`
	Resources []string       `json:"resources,omitempty"`
}

// CommandName returns the command name.
func (c *CreateChallenge) CommandName() string {
	return TypeCreateChallenge
}

// CreateChallengeResult is the result of creating a challenge.
type CreateChallengeResult struct {
	Nonce     string    `json:"nonce"`
	Message   string    `json:"message"`
	Domain    string    `json:"domain"`
	URI       string    `json:"uri"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ============================================================================
// Verify Challenge Command
// ============================================================================

// VerifyChallenge verifies a signed challenge.
type VerifyChallenge struct {
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

// CommandName returns the command name.
func (c *VerifyChallenge) CommandName() string {
	return TypeVerifyChallenge
}

// VerifyChallengeResult is the result of verifying a challenge.
type VerifyChallengeResult struct {
	Valid    bool   `json:"valid"`
	Address  string `json:"address"`
	ChainID  string `json:"chain_id"`
	Nonce    string `json:"nonce"`
	WalletID string `json:"wallet_id,omitempty"`
}

// ============================================================================
// Record Wallet Usage Command
// ============================================================================

// RecordWalletUsage records that a wallet was used.
type RecordWalletUsage struct {
	WalletID  string `json:"wallet_id"`
	Action    string `json:"action"` // "login", "sign", "verify"
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// CommandName returns the command name.
func (c *RecordWalletUsage) CommandName() string {
	return TypeRecordWalletUsage
}

// RecordWalletUsageResult is the result of recording wallet usage.
type RecordWalletUsageResult struct {
	WalletID   string    `json:"wallet_id"`
	Action     string    `json:"action"`
	RecordedAt time.Time `json:"recorded_at"`
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Command = (*LinkWallet)(nil)
	_ cqrs.Command = (*UnlinkWallet)(nil)
	_ cqrs.Command = (*UpdateLabel)(nil)
	_ cqrs.Command = (*SetPrimary)(nil)
	_ cqrs.Command = (*ActivateWallet)(nil)
	_ cqrs.Command = (*DeactivateWallet)(nil)
	_ cqrs.Command = (*SuspendWallet)(nil)
	_ cqrs.Command = (*CreateChallenge)(nil)
	_ cqrs.Command = (*VerifyChallenge)(nil)
	_ cqrs.Command = (*RecordWalletUsage)(nil)
)
