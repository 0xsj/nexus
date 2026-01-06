package domain

import (
	"context"
	"time"
)

// ============================================================================
// Wallet Repository
// ============================================================================

// WalletRepository manages wallet persistence.
type WalletRepository interface {
	// Save persists a wallet (create or update).
	Save(ctx context.Context, wallet *Wallet) error

	// FindByID finds a wallet by its unique ID.
	FindByID(ctx context.Context, id string) (*Wallet, error)

	// FindByAddress finds a wallet by address and chain.
	FindByAddress(ctx context.Context, address Address) (*Wallet, error)

	// FindByAddressString finds a wallet by address string and chain ID.
	FindByAddressString(ctx context.Context, address string, chainID ChainID) (*Wallet, error)

	// FindByUserID finds all wallets for a user.
	FindByUserID(ctx context.Context, userID string) ([]*Wallet, error)

	// FindByDID finds a wallet by its derived DID.
	FindByDID(ctx context.Context, did string) (*Wallet, error)

	// FindPrimaryByUserID finds the primary wallet for a user.
	FindPrimaryByUserID(ctx context.Context, userID string) (*Wallet, error)

	// ExistsByAddress checks if a wallet with the address exists.
	ExistsByAddress(ctx context.Context, address Address) (bool, error)

	// ExistsByAddressString checks if a wallet with the address string exists.
	ExistsByAddressString(ctx context.Context, address string, chainID ChainID) (bool, error)

	// Delete removes a wallet.
	Delete(ctx context.Context, id string) error

	// DeleteByUserID removes all wallets for a user.
	DeleteByUserID(ctx context.Context, userID string) error
}

// ============================================================================
// Challenge Repository
// ============================================================================

// ChallengeRepository manages challenge persistence.
type ChallengeRepository interface {
	// Save persists a challenge.
	Save(ctx context.Context, challenge *Challenge) error

	// FindByNonce finds a challenge by nonce.
	FindByNonce(ctx context.Context, nonce string) (*Challenge, error)

	// MarkUsed marks a challenge as used.
	MarkUsed(ctx context.Context, nonce string) error

	// Delete removes a challenge.
	Delete(ctx context.Context, nonce string) error

	// DeleteExpired removes all expired challenges.
	DeleteExpired(ctx context.Context) (int64, error)

	// ExistsByNonce checks if a challenge with the nonce exists.
	ExistsByNonce(ctx context.Context, nonce string) (bool, error)
}

// ============================================================================
// Nonce Repository
// ============================================================================

// NonceRepository manages nonce persistence for replay protection.
type NonceRepository interface {
	// Save persists a nonce.
	Save(ctx context.Context, nonce *Nonce) error

	// FindByValue finds a nonce by its value.
	FindByValue(ctx context.Context, value string) (*Nonce, error)

	// MarkUsed marks a nonce as used.
	MarkUsed(ctx context.Context, value string) error

	// IsUsed checks if a nonce has been used.
	IsUsed(ctx context.Context, value string) (bool, error)

	// Delete removes a nonce.
	Delete(ctx context.Context, value string) error

	// DeleteExpired removes all expired nonces.
	DeleteExpired(ctx context.Context) (int64, error)
}

// Nonce represents a nonce for replay protection.
type Nonce struct {
	// Value is the nonce value.
	Value string

	// Used indicates if the nonce has been consumed.
	Used bool

	// UsedAt is when the nonce was consumed.
	UsedAt *time.Time

	// CreatedAt is when the nonce was created.
	CreatedAt time.Time

	// ExpiresAt is when the nonce expires.
	ExpiresAt time.Time
}

// IsExpired returns true if the nonce has expired.
func (n Nonce) IsExpired() bool {
	return time.Now().After(n.ExpiresAt)
}

// IsValid returns true if the nonce is valid (not expired, not used).
func (n Nonce) IsValid() bool {
	return !n.IsExpired() && !n.Used
}

// ============================================================================
// Read Models
// ============================================================================

// WalletView is a read model for wallet queries.
type WalletView struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Address    string     `json:"address"`
	ChainID    string     `json:"chain_id"`
	ChainName  string     `json:"chain_name"`
	Family     string     `json:"family"`
	DID        string     `json:"did"`
	Label      string     `json:"label,omitempty"`
	IsPrimary  bool       `json:"is_primary"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// WalletSummary is a lightweight read model for wallet lists.
type WalletSummary struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	ChainID   string    `json:"chain_id"`
	ChainName string    `json:"chain_name"`
	Label     string    `json:"label,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// Query Repository
// ============================================================================

// WalletQueryRepository provides read-only wallet queries.
type WalletQueryRepository interface {
	// GetByID returns a wallet view by ID.
	GetByID(ctx context.Context, id string) (*WalletView, error)

	// GetByAddress returns a wallet view by address.
	GetByAddress(ctx context.Context, address string, chainID ChainID) (*WalletView, error)

	// ListByUserID returns wallet summaries for a user.
	ListByUserID(ctx context.Context, userID string) ([]WalletSummary, error)

	// ListByChain returns wallet summaries for a chain.
	ListByChain(ctx context.Context, chainID ChainID, opts ListOptions) ([]WalletSummary, int, error)

	// CountByUserID returns the number of wallets for a user.
	CountByUserID(ctx context.Context, userID string) (int, error)

	// CountByChain returns the number of wallets for a chain.
	CountByChain(ctx context.Context, chainID ChainID) (int, error)
}

// ============================================================================
// List Options
// ============================================================================

// ListOptions contains options for listing wallets.
type ListOptions struct {
	// Limit is the maximum number of results.
	Limit int

	// Offset is the number of results to skip.
	Offset int

	// SortBy is the field to sort by.
	SortBy string

	// SortOrder is the sort direction ("asc" or "desc").
	SortOrder string
}

// DefaultListOptions returns default list options.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// WithLimit sets the limit.
func (o ListOptions) WithLimit(limit int) ListOptions {
	o.Limit = limit
	return o
}

// WithOffset sets the offset.
func (o ListOptions) WithOffset(offset int) ListOptions {
	o.Offset = offset
	return o
}

// WithSort sets the sort field and order.
func (o ListOptions) WithSort(field, order string) ListOptions {
	o.SortBy = field
	o.SortOrder = order
	return o
}
