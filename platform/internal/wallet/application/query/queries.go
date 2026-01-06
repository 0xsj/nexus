package query

import (
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Query Types
// ============================================================================

const (
	TypeGetWalletByID      = "wallet.get_by_id"
	TypeGetWalletByAddress = "wallet.get_by_address"
	TypeGetWalletByDID     = "wallet.get_by_did"
	TypeListWalletsByUser  = "wallet.list_by_user"
	TypeListWalletsByChain = "wallet.list_by_chain"
	TypeGetPrimaryWallet   = "wallet.get_primary"
	TypeCountWalletsByUser = "wallet.count_by_user"
	TypeCheckWalletExists  = "wallet.check_exists"
)

// ============================================================================
// Get Wallet By ID Query
// ============================================================================

// GetWalletByID retrieves a wallet by its unique ID.
type GetWalletByID struct {
	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id,omitempty"` // Optional: for ownership verification
}

// QueryName returns the query name.
func (q *GetWalletByID) QueryName() string {
	return TypeGetWalletByID
}

// ============================================================================
// Get Wallet By Address Query
// ============================================================================

// GetWalletByAddress retrieves a wallet by address and chain.
type GetWalletByAddress struct {
	Address string         `json:"address"`
	ChainID domain.ChainID `json:"chain_id"`
}

// QueryName returns the query name.
func (q *GetWalletByAddress) QueryName() string {
	return TypeGetWalletByAddress
}

// ============================================================================
// Get Wallet By DID Query
// ============================================================================

// GetWalletByDID retrieves a wallet by its derived DID.
type GetWalletByDID struct {
	DID string `json:"did"`
}

// QueryName returns the query name.
func (q *GetWalletByDID) QueryName() string {
	return TypeGetWalletByDID
}

// ============================================================================
// List Wallets By User Query
// ============================================================================

// ListWalletsByUser retrieves all wallets for a user.
type ListWalletsByUser struct {
	UserID    string               `json:"user_id"`
	Status    *domain.WalletStatus `json:"status,omitempty"`
	Limit     int                  `json:"limit"`
	Offset    int                  `json:"offset"`
	SortBy    string               `json:"sort_by,omitempty"`
	SortOrder string               `json:"sort_order,omitempty"`
}

// QueryName returns the query name.
func (q *ListWalletsByUser) QueryName() string {
	return TypeListWalletsByUser
}

// NewListWalletsByUser creates a new ListWalletsByUser query with defaults.
func NewListWalletsByUser(userID string) *ListWalletsByUser {
	return &ListWalletsByUser{
		UserID:    userID,
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// WithStatus filters by status.
func (q *ListWalletsByUser) WithStatus(status domain.WalletStatus) *ListWalletsByUser {
	q.Status = &status
	return q
}

// WithLimit sets the limit.
func (q *ListWalletsByUser) WithLimit(limit int) *ListWalletsByUser {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *ListWalletsByUser) WithOffset(offset int) *ListWalletsByUser {
	q.Offset = offset
	return q
}

// WithSort sets the sort field and order.
func (q *ListWalletsByUser) WithSort(sortBy, sortOrder string) *ListWalletsByUser {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// ============================================================================
// List Wallets By Chain Query
// ============================================================================

// ListWalletsByChain retrieves wallets by chain.
type ListWalletsByChain struct {
	ChainID   domain.ChainID `json:"chain_id"`
	Limit     int            `json:"limit"`
	Offset    int            `json:"offset"`
	SortBy    string         `json:"sort_by,omitempty"`
	SortOrder string         `json:"sort_order,omitempty"`
}

// QueryName returns the query name.
func (q *ListWalletsByChain) QueryName() string {
	return TypeListWalletsByChain
}

// NewListWalletsByChain creates a new ListWalletsByChain query with defaults.
func NewListWalletsByChain(chainID domain.ChainID) *ListWalletsByChain {
	return &ListWalletsByChain{
		ChainID:   chainID,
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// WithLimit sets the limit.
func (q *ListWalletsByChain) WithLimit(limit int) *ListWalletsByChain {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *ListWalletsByChain) WithOffset(offset int) *ListWalletsByChain {
	q.Offset = offset
	return q
}

// WithSort sets the sort field and order.
func (q *ListWalletsByChain) WithSort(sortBy, sortOrder string) *ListWalletsByChain {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// ============================================================================
// Get Primary Wallet Query
// ============================================================================

// GetPrimaryWallet retrieves the primary wallet for a user.
type GetPrimaryWallet struct {
	UserID string `json:"user_id"`
}

// QueryName returns the query name.
func (q *GetPrimaryWallet) QueryName() string {
	return TypeGetPrimaryWallet
}

// ============================================================================
// Count Wallets By User Query
// ============================================================================

// CountWalletsByUser counts wallets for a user.
type CountWalletsByUser struct {
	UserID     string `json:"user_id"`
	ActiveOnly bool   `json:"active_only"`
}

// QueryName returns the query name.
func (q *CountWalletsByUser) QueryName() string {
	return TypeCountWalletsByUser
}

// NewCountWalletsByUser creates a new CountWalletsByUser query with defaults.
func NewCountWalletsByUser(userID string) *CountWalletsByUser {
	return &CountWalletsByUser{
		UserID:     userID,
		ActiveOnly: false,
	}
}

// WithActiveOnly filters to active wallets only.
func (q *CountWalletsByUser) WithActiveOnly(activeOnly bool) *CountWalletsByUser {
	q.ActiveOnly = activeOnly
	return q
}

// ============================================================================
// Check Wallet Exists Query
// ============================================================================

// CheckWalletExists checks if a wallet exists by various identifiers.
type CheckWalletExists struct {
	Address *AddressIdentifier `json:"address,omitempty"`
	DID     *string            `json:"did,omitempty"`
}

// QueryName returns the query name.
func (q *CheckWalletExists) QueryName() string {
	return TypeCheckWalletExists
}

// AddressIdentifier identifies a wallet by address and chain.
type AddressIdentifier struct {
	Address string         `json:"address"`
	ChainID domain.ChainID `json:"chain_id"`
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Query = (*GetWalletByID)(nil)
	_ cqrs.Query = (*GetWalletByAddress)(nil)
	_ cqrs.Query = (*GetWalletByDID)(nil)
	_ cqrs.Query = (*ListWalletsByUser)(nil)
	_ cqrs.Query = (*ListWalletsByChain)(nil)
	_ cqrs.Query = (*GetPrimaryWallet)(nil)
	_ cqrs.Query = (*CountWalletsByUser)(nil)
	_ cqrs.Query = (*CheckWalletExists)(nil)
)
