package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetWallet         = "wallet.GetWallet"
	QueryListWalletsByUser = "wallet.ListWalletsByUser"
	QueryGetPrimaryWallet  = "wallet.GetPrimaryWallet"
)

// ============================================================================
// GetWallet
// ============================================================================

// GetWallet retrieves a single wallet by ID.
type GetWallet struct {
	WalletID types.ID `json:"wallet_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetWallet) QueryName() string {
	return QueryGetWallet
}

// Validate implements cqrs.Validatable.
func (q GetWallet) Validate() error {
	if q.WalletID.IsZero() {
		return cqrs.ErrQueryValidation("GetWallet.Validate", "wallet_id is required")
	}
	return nil
}

// ============================================================================
// ListWalletsByUser
// ============================================================================

// ListWalletsByUser lists wallets for a user ID.
type ListWalletsByUser struct {
	UserID string `json:"user_id" validate:"required"`
	Limit  int    `json:"limit" validate:"omitempty"`
	Offset int    `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListWalletsByUser) QueryName() string {
	return QueryListWalletsByUser
}

// Validate implements cqrs.Validatable.
func (q ListWalletsByUser) Validate() error {
	if q.UserID == "" {
		return cqrs.ErrQueryValidation("ListWalletsByUser.Validate", "user_id is required")
	}
	return nil
}

// ============================================================================
// GetPrimaryWallet
// ============================================================================

// GetPrimaryWallet retrieves the primary wallet for a user.
type GetPrimaryWallet struct {
	UserID string `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetPrimaryWallet) QueryName() string {
	return QueryGetPrimaryWallet
}

// Validate implements cqrs.Validatable.
func (q GetPrimaryWallet) Validate() error {
	if q.UserID == "" {
		return cqrs.ErrQueryValidation("GetPrimaryWallet.Validate", "user_id is required")
	}
	return nil
}
