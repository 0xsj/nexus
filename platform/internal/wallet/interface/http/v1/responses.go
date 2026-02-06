package v1

import "time"

// ============================================================================
// Wallet Responses
// ============================================================================

// WalletResponse represents a full wallet in API responses.
type WalletResponse struct {
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

// WalletSummaryResponse represents a lightweight wallet in list responses.
type WalletSummaryResponse struct {
	WalletID  string `json:"wallet_id"`
	Address   string `json:"address"`
	Chain     string `json:"chain"`
	Label     string `json:"label,omitempty"`
	Status    string `json:"status"`
	IsPrimary bool   `json:"is_primary"`
}

// WalletListResponse represents a paginated list of wallets.
type WalletListResponse struct {
	Wallets    []WalletSummaryResponse `json:"wallets"`
	TotalCount int                     `json:"total_count"`
	Limit      int                     `json:"limit"`
	Offset     int                     `json:"offset"`
	HasMore    bool                    `json:"has_more"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// WalletLinkedResponse represents the result of linking a wallet.
type WalletLinkedResponse struct {
	WalletID string    `json:"wallet_id"`
	Address  string    `json:"address"`
	Status   string    `json:"status"`
	LinkedAt time.Time `json:"linked_at"`
}

// WalletVerifiedResponse represents the result of verifying a wallet.
type WalletVerifiedResponse struct {
	WalletID   string    `json:"wallet_id"`
	Status     string    `json:"status"`
	VerifiedAt time.Time `json:"verified_at"`
}

// WalletUnlinkedResponse represents the result of unlinking a wallet.
type WalletUnlinkedResponse struct {
	WalletID   string    `json:"wallet_id"`
	Status     string    `json:"status"`
	UnlinkedAt time.Time `json:"unlinked_at"`
}

// WalletPrimarySetResponse represents the result of setting a primary wallet.
type WalletPrimarySetResponse struct {
	WalletID string `json:"wallet_id"`
	Status   string `json:"status"`
}

// WalletLabelUpdatedResponse represents the result of updating a wallet label.
type WalletLabelUpdatedResponse struct {
	WalletID string `json:"wallet_id"`
	Label    string `json:"label"`
}
