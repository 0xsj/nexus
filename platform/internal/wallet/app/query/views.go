package query

import "time"

// ============================================================================
// Wallet Views
// ============================================================================

// WalletView is the full wallet read model.
type WalletView struct {
	WalletID   string     `json:"wallet_id"`
	UserID     string     `json:"user_id"`
	Address    string     `json:"address"`
	Chain      string     `json:"chain"`
	Label      string     `json:"label,omitempty"`
	Status     string     `json:"status"`
	IsPrimary  bool       `json:"is_primary"`
	DID        string     `json:"did,omitempty"`
	LinkedAt   time.Time  `json:"linked_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// WalletSummaryView is a lightweight wallet representation for lists.
type WalletSummaryView struct {
	WalletID  string `json:"wallet_id"`
	Address   string `json:"address"`
	Chain     string `json:"chain"`
	Label     string `json:"label,omitempty"`
	Status    string `json:"status"`
	IsPrimary bool   `json:"is_primary"`
}

// WalletListView is a paginated list of wallets.
type WalletListView struct {
	Wallets    []WalletSummaryView `json:"wallets"`
	TotalCount int                 `json:"total_count"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	HasMore    bool                `json:"has_more"`
}
