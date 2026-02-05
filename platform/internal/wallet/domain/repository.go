package domain

import (
	"context"
)

// ============================================================================
// Wallet Repository
// ============================================================================

// WalletRepository defines the persistence operations for Wallet aggregates.
type WalletRepository interface {
	// Save persists a wallet aggregate (appends new events).
	Save(ctx context.Context, wallet *Wallet) error

	// Get retrieves a wallet by ID (replays events to rebuild state).
	Get(ctx context.Context, id WalletID) (*Wallet, error)

	// Exists checks if a wallet with the given ID exists.
	Exists(ctx context.Context, id WalletID) (bool, error)
}
