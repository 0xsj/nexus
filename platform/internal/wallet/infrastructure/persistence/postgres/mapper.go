package postgres

import (
	"database/sql"
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/application/query"
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/did"
)

// ============================================================================
// Wallet Row
// ============================================================================

// walletRow represents a wallet row in the database.
type walletRow struct {
	ID                string         `db:"id"`
	UserID            string         `db:"user_id"`
	Address           string         `db:"address"`
	AddressNormalized string         `db:"address_normalized"`
	ChainID           string         `db:"chain_id"`
	ChainFamily       string         `db:"chain_family"`
	DID               string         `db:"did"`
	Label             sql.NullString `db:"label"`
	IsPrimary         bool           `db:"is_primary"`
	Status            string         `db:"status"`
	VerifiedAt        time.Time      `db:"verified_at"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	LastUsedAt        sql.NullTime   `db:"last_used_at"`
}

// toDomain converts a wallet row to a domain wallet.
func (r *walletRow) toDomain() (*domain.Wallet, error) {
	// Reconstruct address
	address := domain.NewAddressUnchecked(
		r.Address,
		r.AddressNormalized,
		domain.ChainID(r.ChainID),
		domain.ChainFamily(r.ChainFamily),
	)

	// Parse DID
	walletDID, err := did.Parse(r.DID)
	if err != nil {
		// If DID parsing fails, create empty DID
		walletDID = did.DID{}
	}

	// Parse status
	status := domain.WalletStatus(r.Status)

	// Rehydrate wallet from persisted state
	wallet := domain.RehydrateWallet(
		r.ID,
		r.UserID,
		address,
		walletDID,
		nullStringToString(r.Label),
		r.IsPrimary,
		status,
		r.VerifiedAt,
		nullTimeToPtr(r.LastUsedAt),
		r.CreatedAt,
		r.UpdatedAt,
	)

	return wallet, nil
}

// toView converts a wallet row to a query view.
func (r *walletRow) toView() *query.WalletView {
	chainInfo := domain.GetChainInfo(domain.ChainID(r.ChainID))
	chainName := r.ChainID
	if !chainInfo.IsZero() {
		chainName = chainInfo.DisplayName
	}

	return &query.WalletView{
		ID:         r.ID,
		UserID:     r.UserID,
		Address:    r.AddressNormalized,
		ChainID:    r.ChainID,
		ChainName:  chainName,
		Family:     r.ChainFamily,
		DID:        r.DID,
		Label:      nullStringToString(r.Label),
		IsPrimary:  r.IsPrimary,
		Status:     r.Status,
		VerifiedAt: r.VerifiedAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
		LastUsedAt: nullTimeToPtr(r.LastUsedAt),
	}
}

// toSummaryView converts a wallet row to a summary view.
func (r *walletRow) toSummaryView() query.WalletSummaryView {
	chainInfo := domain.GetChainInfo(domain.ChainID(r.ChainID))
	chainName := r.ChainID
	if !chainInfo.IsZero() {
		chainName = chainInfo.DisplayName
	}

	return query.WalletSummaryView{
		ID:        r.ID,
		Address:   r.AddressNormalized,
		ChainID:   r.ChainID,
		ChainName: chainName,
		Label:     nullStringToString(r.Label),
		IsPrimary: r.IsPrimary,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
	}
}

// walletToRow converts a domain wallet to a database row.
func walletToRow(w *domain.Wallet) *walletRow {
	return &walletRow{
		ID:                w.ID(),
		UserID:            w.UserID(),
		Address:           w.Address().Raw(),
		AddressNormalized: w.Address().Normalized(),
		ChainID:           w.ChainID().String(),
		ChainFamily:       w.Family().String(),
		DID:               w.DIDString(),
		Label:             stringToNullString(w.Label()),
		IsPrimary:         w.IsPrimary(),
		Status:            w.Status().String(),
		VerifiedAt:        w.VerifiedAt(),
		CreatedAt:         w.CreatedAt(),
		UpdatedAt:         w.UpdatedAt(),
		LastUsedAt:        ptrToNullTime(w.LastUsedAt()),
	}
}

// ============================================================================
// Challenge Row
// ============================================================================

// challengeRow represents a challenge row in the database.
type challengeRow struct {
	Nonce             string       `db:"nonce"`
	Message           string       `db:"message"`
	Address           string       `db:"address"`
	AddressNormalized string       `db:"address_normalized"`
	ChainID           string       `db:"chain_id"`
	Domain            string       `db:"domain"`
	URI               string       `db:"uri"`
	IssuedAt          time.Time    `db:"issued_at"`
	ExpiresAt         time.Time    `db:"expires_at"`
	Used              bool         `db:"used"`
	UsedAt            sql.NullTime `db:"used_at"`
}

// toDomain converts a challenge row to a domain challenge.
func (r *challengeRow) toDomain() *domain.Challenge {
	// Reconstruct address
	chainInfo := domain.GetChainInfo(domain.ChainID(r.ChainID))
	family := domain.ChainFamilyEVM
	if !chainInfo.IsZero() {
		family = chainInfo.Family
	}

	address := domain.NewAddressUnchecked(
		r.Address,
		r.AddressNormalized,
		domain.ChainID(r.ChainID),
		family,
	)

	return &domain.Challenge{
		Nonce:     r.Nonce,
		Message:   r.Message,
		Address:   address,
		Domain:    r.Domain,
		URI:       r.URI,
		IssuedAt:  r.IssuedAt,
		ExpiresAt: r.ExpiresAt,
		Used:      r.Used,
		UsedAt:    nullTimeToPtr(r.UsedAt),
	}
}

// toView converts a challenge row to a query view.
func (r *challengeRow) toView() *query.ChallengeView {
	return &query.ChallengeView{
		Nonce:     r.Nonce,
		Message:   r.Message,
		Address:   r.AddressNormalized,
		ChainID:   r.ChainID,
		Domain:    r.Domain,
		URI:       r.URI,
		IssuedAt:  r.IssuedAt,
		ExpiresAt: r.ExpiresAt,
		Used:      r.Used,
		UsedAt:    nullTimeToPtr(r.UsedAt),
	}
}

// challengeToRow converts a domain challenge to a database row.
func challengeToRow(c *domain.Challenge) *challengeRow {
	return &challengeRow{
		Nonce:             c.Nonce,
		Message:           c.Message,
		Address:           c.Address.Raw(),
		AddressNormalized: c.Address.Normalized(),
		ChainID:           c.Address.ChainID().String(),
		Domain:            c.Domain,
		URI:               c.URI,
		IssuedAt:          c.IssuedAt,
		ExpiresAt:         c.ExpiresAt,
		Used:              c.Used,
		UsedAt:            ptrToNullTime(c.UsedAt),
	}
}

// ============================================================================
// Nonce Row
// ============================================================================

// nonceRow represents a nonce row in the database.
type nonceRow struct {
	Nonce     string       `db:"nonce"`
	CreatedAt time.Time    `db:"created_at"`
	ExpiresAt time.Time    `db:"expires_at"`
	Used      bool         `db:"used"`
	UsedAt    sql.NullTime `db:"used_at"`
}

// ============================================================================
// Activity Row
// ============================================================================

// activityRow represents a wallet activity row in the database.
type activityRow struct {
	ID        string         `db:"id"`
	WalletID  string         `db:"wallet_id"`
	Action    string         `db:"action"`
	IPAddress sql.NullString `db:"ip_address"`
	UserAgent sql.NullString `db:"user_agent"`
	CreatedAt time.Time      `db:"created_at"`
}

// toView converts an activity row to a query view.
func (r *activityRow) toView() query.WalletActivityView {
	return query.WalletActivityView{
		WalletID:  r.WalletID,
		Action:    r.Action,
		IPAddress: nullStringToString(r.IPAddress),
		UserAgent: nullStringToString(r.UserAgent),
		Timestamp: r.CreatedAt,
	}
}

// ============================================================================
// Statistics Rows
// ============================================================================

// walletStatsRow represents wallet statistics.
type walletStatsRow struct {
	TotalCount  int `db:"total_count"`
	ActiveCount int `db:"active_count"`
}

// chainCountRow represents wallet count by chain.
type chainCountRow struct {
	ChainID string `db:"chain_id"`
	Count   int    `db:"count"`
}

// toView converts chain count rows to view.
func chainCountsToView(rows []chainCountRow) []query.ChainCountView {
	result := make([]query.ChainCountView, len(rows))
	for i, r := range rows {
		chainInfo := domain.GetChainInfo(domain.ChainID(r.ChainID))
		chainName := r.ChainID
		if !chainInfo.IsZero() {
			chainName = chainInfo.DisplayName
		}

		result[i] = query.ChainCountView{
			ChainID:   r.ChainID,
			ChainName: chainName,
			Count:     r.Count,
		}
	}
	return result
}

// ============================================================================
// Null Type Helpers
// ============================================================================

// nullTimeToPtr converts sql.NullTime to *time.Time.
func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// ptrToNullTime converts *time.Time to sql.NullTime.
func ptrToNullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{Valid: false}
}

// stringToNullString converts string to sql.NullString.
func stringToNullString(s string) sql.NullString {
	if s != "" {
		return sql.NullString{String: s, Valid: true}
	}
	return sql.NullString{Valid: false}
}

// nullStringToString converts sql.NullString to string.
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}