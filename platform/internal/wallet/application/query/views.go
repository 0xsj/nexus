package query

import (
	"time"
)

// ============================================================================
// Wallet Views
// ============================================================================

// WalletView is the read model for a wallet.
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
	VerifiedAt time.Time  `json:"verified_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// WalletSummaryView is a minimal wallet view for lists.
type WalletSummaryView struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	ChainID   string    `json:"chain_id"`
	ChainName string    `json:"chain_name"`
	Label     string    `json:"label,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// List Views
// ============================================================================

// WalletListView is a paginated list of wallets.
type WalletListView struct {
	Wallets []WalletSummaryView `json:"wallets"`
	Total   int                 `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	HasMore bool                `json:"has_more"`
}

// ============================================================================
// Count Views
// ============================================================================

// WalletCountView is the result of counting wallets.
type WalletCountView struct {
	UserID string `json:"user_id"`
	Count  int    `json:"count"`
}

// ============================================================================
// Chain Stats Views
// ============================================================================

// ChainStatsView contains statistics about wallets on a chain.
type ChainStatsView struct {
	ChainID     string `json:"chain_id"`
	ChainName   string `json:"chain_name"`
	WalletCount int    `json:"wallet_count"`
	ActiveCount int    `json:"active_count"`
}

// ============================================================================
// User Wallet Stats Views
// ============================================================================

// UserWalletStatsView contains statistics about a user's wallets.
type UserWalletStatsView struct {
	UserID        string           `json:"user_id"`
	TotalCount    int              `json:"total_count"`
	ActiveCount   int              `json:"active_count"`
	PrimaryWallet *WalletView      `json:"primary_wallet,omitempty"`
	ByChain       []ChainCountView `json:"by_chain,omitempty"`
}

// ChainCountView contains wallet count for a specific chain.
type ChainCountView struct {
	ChainID   string `json:"chain_id"`
	ChainName string `json:"chain_name"`
	Count     int    `json:"count"`
}

// ============================================================================
// Public Wallet View
// ============================================================================

// PublicWalletView is a public-facing wallet view.
// Address is intentionally omitted for privacy.
type PublicWalletView struct {
	ChainID   string    `json:"chain_id"`
	ChainName string    `json:"chain_name"`
	DID       string    `json:"did"`
	IsPrimary bool      `json:"is_primary"`
	LinkedAt  time.Time `json:"linked_at"`
}

// ============================================================================
// Challenge Views
// ============================================================================

// ChallengeView is the read model for a challenge.
type ChallengeView struct {
	Nonce     string     `json:"nonce"`
	Message   string     `json:"message"`
	Address   string     `json:"address"`
	ChainID   string     `json:"chain_id"`
	Domain    string     `json:"domain"`
	URI       string     `json:"uri"`
	IssuedAt  time.Time  `json:"issued_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	Used      bool       `json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// ============================================================================
// Verification Views
// ============================================================================

// WalletVerificationView is returned after verifying wallet ownership.
type WalletVerificationView struct {
	Valid    bool   `json:"valid"`
	Address  string `json:"address,omitempty"`
	ChainID  string `json:"chain_id,omitempty"`
	DID      string `json:"did,omitempty"`
	WalletID string `json:"wallet_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ============================================================================
// Activity Views
// ============================================================================

// WalletActivityView represents a wallet activity entry.
type WalletActivityView struct {
	WalletID  string    `json:"wallet_id"`
	Action    string    `json:"action"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// WalletActivityListView is a paginated list of wallet activities.
type WalletActivityListView struct {
	Activities []WalletActivityView `json:"activities"`
	Total      int                  `json:"total"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	HasMore    bool                 `json:"has_more"`
}
