package domain

// ============================================================================
// Wallet Status
// ============================================================================

// WalletStatus represents the status of a wallet.
type WalletStatus string

const (
	// WalletStatusActive indicates the wallet is active.
	WalletStatusActive WalletStatus = "active"

	// WalletStatusInactive indicates the wallet is inactive.
	WalletStatusInactive WalletStatus = "inactive"

	// WalletStatusSuspended indicates the wallet is suspended.
	WalletStatusSuspended WalletStatus = "suspended"
)

// String returns the string representation of the status.
func (s WalletStatus) String() string {
	return string(s)
}

// IsValid returns true if the status is valid.
func (s WalletStatus) IsValid() bool {
	switch s {
	case WalletStatusActive, WalletStatusInactive, WalletStatusSuspended:
		return true
	default:
		return false
	}
}

// IsActive returns true if the wallet is active.
func (s WalletStatus) IsActive() bool {
	return s == WalletStatusActive
}
