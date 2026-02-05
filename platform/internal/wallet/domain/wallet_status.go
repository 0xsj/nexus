package domain

import (
	"fmt"
)

// WalletStatus represents the current status of a linked wallet.
type WalletStatus int

const (
	// WalletStatusUnverified indicates the wallet has been linked but not yet verified.
	WalletStatusUnverified WalletStatus = 1

	// WalletStatusActive indicates the wallet has been verified and is active.
	WalletStatusActive WalletStatus = 2

	// WalletStatusRevoked indicates the wallet has been revoked/unlinked.
	WalletStatusRevoked WalletStatus = 3
)

// String returns the string representation of the wallet status.
func (s WalletStatus) String() string {
	switch s {
	case WalletStatusUnverified:
		return "unverified"
	case WalletStatusActive:
		return "active"
	case WalletStatusRevoked:
		return "revoked"
	default:
		return "unknown"
	}
}

// ParseWalletStatus parses a string into a WalletStatus.
func ParseWalletStatus(str string) (WalletStatus, error) {
	switch str {
	case "unverified":
		return WalletStatusUnverified, nil
	case "active":
		return WalletStatusActive, nil
	case "revoked":
		return WalletStatusRevoked, nil
	default:
		return 0, fmt.Errorf("invalid wallet status: %s", str)
	}
}

// IsUnverified returns true if the wallet status is unverified.
func (s WalletStatus) IsUnverified() bool {
	return s == WalletStatusUnverified
}

// IsActive returns true if the wallet status is active.
func (s WalletStatus) IsActive() bool {
	return s == WalletStatusActive
}

// IsRevoked returns true if the wallet status is revoked.
func (s WalletStatus) IsRevoked() bool {
	return s == WalletStatusRevoked
}

// IsTerminal returns true if the wallet status is a terminal state.
func (s WalletStatus) IsTerminal() bool {
	return s == WalletStatusRevoked
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s WalletStatus) CanTransitionTo(target WalletStatus) bool {
	switch s {
	case WalletStatusUnverified:
		// Unverified can transition to active or revoked
		return target == WalletStatusActive || target == WalletStatusRevoked
	case WalletStatusActive:
		// Active can transition to revoked
		return target == WalletStatusRevoked
	case WalletStatusRevoked:
		// Revoked is terminal
		return false
	default:
		return false
	}
}
