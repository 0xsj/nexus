package v1

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/application/query"
)

// ============================================================================
// Challenge Responses
// ============================================================================

// ChallengeResponse is the response containing a SIWE challenge.
type ChallengeResponse struct {
	Nonce     string    `json:"nonce"`
	Message   string    `json:"message"`
	Address   string    `json:"address"`
	ChainID   string    `json:"chain_id"`
	Domain    string    `json:"domain"`
	URI       string    `json:"uri"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// VerifyChallengeResponse is the response after verifying a challenge.
type VerifyChallengeResponse struct {
	Valid   bool   `json:"valid"`
	Address string `json:"address"`
	ChainID string `json:"chain_id"`
	DID     string `json:"did,omitempty"`
}

// ============================================================================
// Wallet Responses
// ============================================================================

// WalletResponse is the response containing wallet details.
type WalletResponse struct {
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

// FromWalletView converts a query.WalletView to WalletResponse.
func FromWalletView(v *query.WalletView) *WalletResponse {
	if v == nil {
		return nil
	}
	return &WalletResponse{
		ID:         v.ID,
		UserID:     v.UserID,
		Address:    v.Address,
		ChainID:    v.ChainID,
		ChainName:  v.ChainName,
		Family:     v.Family,
		DID:        v.DID,
		Label:      v.Label,
		IsPrimary:  v.IsPrimary,
		Status:     v.Status,
		VerifiedAt: v.VerifiedAt,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
		LastUsedAt: v.LastUsedAt,
	}
}

// WalletSummaryResponse is a summary of a wallet.
type WalletSummaryResponse struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	ChainID   string    `json:"chain_id"`
	ChainName string    `json:"chain_name"`
	Label     string    `json:"label,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// FromWalletSummaryView converts a query.WalletSummaryView to WalletSummaryResponse.
func FromWalletSummaryView(v query.WalletSummaryView) WalletSummaryResponse {
	return WalletSummaryResponse{
		ID:        v.ID,
		Address:   v.Address,
		ChainID:   v.ChainID,
		ChainName: v.ChainName,
		Label:     v.Label,
		IsPrimary: v.IsPrimary,
		Status:    v.Status,
		CreatedAt: v.CreatedAt,
	}
}

// WalletListResponse is the response containing a list of wallets.
type WalletListResponse struct {
	Wallets []WalletSummaryResponse `json:"wallets"`
	Total   int                     `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasMore bool                    `json:"has_more"`
}

// FromWalletListView converts a query.WalletListView to WalletListResponse.
func FromWalletListView(v *query.WalletListView) *WalletListResponse {
	if v == nil {
		return nil
	}

	resp := &WalletListResponse{
		Total:   v.Total,
		Limit:   v.Limit,
		Offset:  v.Offset,
		HasMore: v.HasMore,
	}

	for _, w := range v.Wallets {
		resp.Wallets = append(resp.Wallets, FromWalletSummaryView(w))
	}

	// Ensure empty array instead of null
	if resp.Wallets == nil {
		resp.Wallets = []WalletSummaryResponse{}
	}

	return resp
}

// ============================================================================
// Link Responses
// ============================================================================

// LinkWalletResponse is the response after linking a wallet.
type LinkWalletResponse struct {
	WalletID  string    `json:"wallet_id"`
	UserID    string    `json:"user_id"`
	Address   string    `json:"address"`
	ChainID   string    `json:"chain_id"`
	ChainName string    `json:"chain_name"`
	DID       string    `json:"did"`
	IsPrimary bool      `json:"is_primary"`
	LinkedAt  time.Time `json:"linked_at"`
}

// UnlinkWalletResponse is the response after unlinking a wallet.
type UnlinkWalletResponse struct {
	WalletID   string    `json:"wallet_id"`
	Address    string    `json:"address"`
	ChainID    string    `json:"chain_id"`
	UnlinkedAt time.Time `json:"unlinked_at"`
}

// ReverifyWalletResponse is the response after re-verifying a wallet.
type ReverifyWalletResponse struct {
	WalletID   string    `json:"wallet_id"`
	Address    string    `json:"address"`
	Verified   bool      `json:"verified"`
	VerifiedAt time.Time `json:"verified_at"`
}

// ============================================================================
// Update Responses
// ============================================================================

// UpdateWalletResponse is the response after updating a wallet.
type UpdateWalletResponse struct {
	WalletID  string    `json:"wallet_id"`
	Label     string    `json:"label,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetPrimaryWalletResponse is the response after setting a primary wallet.
type SetPrimaryWalletResponse struct {
	WalletID        string `json:"wallet_id"`
	IsPrimary       bool   `json:"is_primary"`
	PreviousPrimary string `json:"previous_primary,omitempty"`
}

// ============================================================================
// Status Responses
// ============================================================================

// ActivateWalletResponse is the response after activating a wallet.
type ActivateWalletResponse struct {
	WalletID    string    `json:"wallet_id"`
	Status      string    `json:"status"`
	ActivatedAt time.Time `json:"activated_at"`
}

// DeactivateWalletResponse is the response after deactivating a wallet.
type DeactivateWalletResponse struct {
	WalletID      string    `json:"wallet_id"`
	Status        string    `json:"status"`
	DeactivatedAt time.Time `json:"deactivated_at"`
}

// ============================================================================
// Statistics Responses
// ============================================================================

// UserWalletStatsResponse is the response containing user wallet statistics.
type UserWalletStatsResponse struct {
	UserID        string               `json:"user_id"`
	TotalCount    int                  `json:"total_count"`
	ActiveCount   int                  `json:"active_count"`
	PrimaryWallet *WalletResponse      `json:"primary_wallet,omitempty"`
	ByChain       []ChainCountResponse `json:"by_chain"`
}

// FromUserWalletStatsView converts a query.UserWalletStatsView to UserWalletStatsResponse.
func FromUserWalletStatsView(v *query.UserWalletStatsView) *UserWalletStatsResponse {
	if v == nil {
		return nil
	}

	resp := &UserWalletStatsResponse{
		UserID:        v.UserID,
		TotalCount:    v.TotalCount,
		ActiveCount:   v.ActiveCount,
		PrimaryWallet: FromWalletView(v.PrimaryWallet),
	}

	for _, c := range v.ByChain {
		resp.ByChain = append(resp.ByChain, ChainCountResponse{
			ChainID:   c.ChainID,
			ChainName: c.ChainName,
			Count:     c.Count,
		})
	}

	// Ensure empty array instead of null
	if resp.ByChain == nil {
		resp.ByChain = []ChainCountResponse{}
	}

	return resp
}

// ChainCountResponse is a count of wallets for a chain.
type ChainCountResponse struct {
	ChainID   string `json:"chain_id"`
	ChainName string `json:"chain_name"`
	Count     int    `json:"count"`
}

// ChainStatsResponse is the response containing chain statistics.
type ChainStatsResponse struct {
	ChainID     string `json:"chain_id"`
	ChainName   string `json:"chain_name"`
	WalletCount int    `json:"wallet_count"`
	ActiveCount int    `json:"active_count"`
}

// FromChainStatsView converts a query.ChainStatsView to ChainStatsResponse.
func FromChainStatsView(v *query.ChainStatsView) *ChainStatsResponse {
	if v == nil {
		return nil
	}
	return &ChainStatsResponse{
		ChainID:     v.ChainID,
		ChainName:   v.ChainName,
		WalletCount: v.WalletCount,
		ActiveCount: v.ActiveCount,
	}
}

// ============================================================================
// Supported Chains Response
// ============================================================================

// SupportedChainResponse describes a supported blockchain.
type SupportedChainResponse struct {
	ChainID     string `json:"chain_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Family      string `json:"family"`
	IsTestnet   bool   `json:"is_testnet"`
}

// SupportedChainsResponse is the response containing supported chains.
type SupportedChainsResponse struct {
	Chains []SupportedChainResponse `json:"chains"`
}
