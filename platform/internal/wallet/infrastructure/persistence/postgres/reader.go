package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/wallet/application/query"
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Wallet Reader
// ============================================================================

// WalletReader implements query.WalletReader using PostgreSQL.
type WalletReader struct {
	adapter *postgres.BaseAdapter
}

// NewWalletReader creates a new PostgreSQL wallet reader.
func NewWalletReader(adapter *postgres.BaseAdapter) *WalletReader {
	return &WalletReader{adapter: adapter}
}

// ============================================================================
// Single Wallet Queries
// ============================================================================

// GetWallet retrieves a wallet by ID.
func (r *WalletReader) GetWallet(ctx context.Context, walletID string) (*query.WalletView, error) {
	const op = "WalletReader.GetWallet"

	row := r.adapter.Executor().QueryRow(ctx, querySelectWalletByID, walletID)

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, walletID)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toView(), nil
}

// GetWalletByAddress retrieves a wallet by address and chain.
func (r *WalletReader) GetWalletByAddress(ctx context.Context, address string, chainID domain.ChainID) (*query.WalletView, error) {
	const op = "WalletReader.GetWalletByAddress"

	row := r.adapter.Executor().QueryRow(ctx, querySelectWalletByAddress, address, chainID.String())

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, address)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toView(), nil
}

// GetWalletByDID retrieves a wallet by its derived DID.
func (r *WalletReader) GetWalletByDID(ctx context.Context, did string) (*query.WalletView, error) {
	const op = "WalletReader.GetWalletByDID"

	row := r.adapter.Executor().QueryRow(ctx, querySelectWalletByDID, did)

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, did)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toView(), nil
}

// GetPrimaryWallet retrieves the primary wallet for a user.
func (r *WalletReader) GetPrimaryWallet(ctx context.Context, userID string) (*query.WalletView, error) {
	const op = "WalletReader.GetPrimaryWallet"

	row := r.adapter.Executor().QueryRow(ctx, querySelectPrimaryWalletByUserID, userID)

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, "primary wallet for user "+userID)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toView(), nil
}

// ============================================================================
// List Queries
// ============================================================================

// ListWalletsByUser retrieves wallets for a user.
func (r *WalletReader) ListWalletsByUser(ctx context.Context, userID string, opts query.ListWalletsOptions) (*query.WalletListView, error) {
	const op = "WalletReader.ListWalletsByUser"

	// Build query based on status filter
	var rows rowScanner
	var err error

	if opts.Status != nil {
		rows, err = r.adapter.Select(ctx, querySelectWalletsByUserIDWithStatus, userID, opts.Status.String())
	} else {
		rows, err = r.adapter.Select(ctx, querySelectWalletsByUserID, userID)
	}

	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	defer rows.Close()

	wallets, err := r.scanWalletSummaries(rows)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Apply pagination in memory (for simplicity)
	// In production, use LIMIT/OFFSET in query
	total := len(wallets)
	start := opts.Offset
	end := opts.Offset + opts.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedWallets := wallets[start:end]

	return &query.WalletListView{
		Wallets: paginatedWallets,
		Total:   total,
		Limit:   opts.Limit,
		Offset:  opts.Offset,
		HasMore: end < total,
	}, nil
}

// ListWalletsByChain retrieves wallets by chain.
func (r *WalletReader) ListWalletsByChain(ctx context.Context, chainID domain.ChainID, opts query.ListOptions) (*query.WalletListView, error) {
	const op = "WalletReader.ListWalletsByChain"

	rows, err := r.adapter.Select(ctx, querySelectWalletsByChainID, chainID.String(), opts.Limit, opts.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	defer rows.Close()

	wallets, err := r.scanWalletSummaries(rows)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Get total count
	total, err := r.CountWalletsByChain(ctx, chainID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &query.WalletListView{
		Wallets: wallets,
		Total:   total,
		Limit:   opts.Limit,
		Offset:  opts.Offset,
		HasMore: opts.Offset+len(wallets) < total,
	}, nil
}

// ============================================================================
// Count Queries
// ============================================================================

// CountWalletsByUser counts wallets for a user.
func (r *WalletReader) CountWalletsByUser(ctx context.Context, userID string, activeOnly bool) (int, error) {
	const op = "WalletReader.CountWalletsByUser"

	var count int
	var row interface{ Scan(...interface{}) error }

	if activeOnly {
		row = r.adapter.Executor().QueryRow(ctx, queryCountActiveWalletsByUserID, userID)
	} else {
		row = r.adapter.Executor().QueryRow(ctx, queryCountWalletsByUserID, userID)
	}

	if err := row.Scan(&count); err != nil {
		return 0, pkgerrors.Wrap(err, op)
	}

	return count, nil
}

// CountWalletsByChain counts wallets for a chain.
func (r *WalletReader) CountWalletsByChain(ctx context.Context, chainID domain.ChainID) (int, error) {
	const op = "WalletReader.CountWalletsByChain"

	var count int
	row := r.adapter.Executor().QueryRow(ctx, queryCountWalletsByChainID, chainID.String())
	if err := row.Scan(&count); err != nil {
		return 0, pkgerrors.Wrap(err, op)
	}

	return count, nil
}

// ============================================================================
// Existence Checks
// ============================================================================

// WalletExists checks if a wallet exists.
func (r *WalletReader) WalletExists(ctx context.Context, walletID string) (bool, error) {
	const op = "WalletReader.WalletExists"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExistsWalletByID, walletID)
	if err := row.Scan(&exists); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}

	return exists, nil
}

// WalletExistsByAddress checks if a wallet exists by address.
func (r *WalletReader) WalletExistsByAddress(ctx context.Context, address string, chainID domain.ChainID) (bool, error) {
	const op = "WalletReader.WalletExistsByAddress"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExistsWalletByAddress, address, chainID.String())
	if err := row.Scan(&exists); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}

	return exists, nil
}

// WalletExistsByDID checks if a wallet exists by DID.
func (r *WalletReader) WalletExistsByDID(ctx context.Context, did string) (bool, error) {
	const op = "WalletReader.WalletExistsByDID"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExistsWalletByDID, did)
	if err := row.Scan(&exists); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}

	return exists, nil
}

// ============================================================================
// Statistics
// ============================================================================

// GetUserWalletStats retrieves wallet statistics for a user.
func (r *WalletReader) GetUserWalletStats(ctx context.Context, userID string) (*query.UserWalletStatsView, error) {
	const op = "WalletReader.GetUserWalletStats"

	// Get counts
	var stats walletStatsRow
	row := r.adapter.Executor().QueryRow(ctx, queryUserWalletStats, userID)
	if err := row.Scan(&stats.TotalCount, &stats.ActiveCount); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Get primary wallet
	var primaryWallet *query.WalletView
	primary, err := r.GetPrimaryWallet(ctx, userID)
	if err == nil {
		primaryWallet = primary
	}

	// Get counts by chain
	chainCounts, err := r.getWalletCountsByChain(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &query.UserWalletStatsView{
		UserID:        userID,
		TotalCount:    stats.TotalCount,
		ActiveCount:   stats.ActiveCount,
		PrimaryWallet: primaryWallet,
		ByChain:       chainCounts,
	}, nil
}

// GetChainStats retrieves wallet statistics for a chain.
func (r *WalletReader) GetChainStats(ctx context.Context, chainID domain.ChainID) (*query.ChainStatsView, error) {
	const op = "WalletReader.GetChainStats"

	var stats walletStatsRow
	row := r.adapter.Executor().QueryRow(ctx, queryChainStats, chainID.String())
	if err := row.Scan(&stats.TotalCount, &stats.ActiveCount); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	chainInfo := domain.GetChainInfo(chainID)
	chainName := chainID.String()
	if !chainInfo.IsZero() {
		chainName = chainInfo.DisplayName
	}

	return &query.ChainStatsView{
		ChainID:     chainID.String(),
		ChainName:   chainName,
		WalletCount: stats.TotalCount,
		ActiveCount: stats.ActiveCount,
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// rowScanner interface for rows that can be scanned.
type rowScanner interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
	Close() error
}

// scanWalletSummaries scans multiple wallet rows into summaries.
func (r *WalletReader) scanWalletSummaries(rows rowScanner) ([]query.WalletSummaryView, error) {
	var wallets []query.WalletSummaryView

	for rows.Next() {
		walletRow, err := scanWalletRow(rows)
		if err != nil {
			return nil, err
		}

		wallets = append(wallets, walletRow.toSummaryView())
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}

// getWalletCountsByChain gets wallet counts grouped by chain for a user.
func (r *WalletReader) getWalletCountsByChain(ctx context.Context, userID string) ([]query.ChainCountView, error) {
	const op = "WalletReader.getWalletCountsByChain"

	rows, err := r.adapter.Select(ctx, queryWalletCountByChainForUser, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	defer rows.Close()

	var counts []chainCountRow
	for rows.Next() {
		var row chainCountRow
		if err := rows.Scan(&row.ChainID, &row.Count); err != nil {
			return nil, err
		}
		counts = append(counts, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chainCountsToView(counts), nil
}

// ============================================================================
// Challenge Reader
// ============================================================================

// ChallengeReader implements query.ChallengeReader using PostgreSQL.
type ChallengeReader struct {
	adapter *postgres.BaseAdapter
}

// NewChallengeReader creates a new PostgreSQL challenge reader.
func NewChallengeReader(adapter *postgres.BaseAdapter) *ChallengeReader {
	return &ChallengeReader{adapter: adapter}
}

// GetChallenge retrieves a challenge by nonce.
func (r *ChallengeReader) GetChallenge(ctx context.Context, nonce string) (*query.ChallengeView, error) {
	const op = "ChallengeReader.GetChallenge"

	row := r.adapter.Executor().QueryRow(ctx, querySelectChallengeByNonce, nonce)

	challengeRow, err := scanChallengeRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidNonce(op, nonce)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return challengeRow.toView(), nil
}

// ChallengeExists checks if a challenge exists and is valid.
func (r *ChallengeReader) ChallengeExists(ctx context.Context, nonce string) (bool, error) {
	const op = "ChallengeReader.ChallengeExists"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExistsChallengeByNonce, nonce)
	if err := row.Scan(&exists); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}

	return exists, nil
}

// ChallengeIsValid checks if a challenge is valid (not expired, not used).
func (r *ChallengeReader) ChallengeIsValid(ctx context.Context, nonce string) (bool, error) {
	const op = "ChallengeReader.ChallengeIsValid"

	var valid bool
	row := r.adapter.Executor().QueryRow(ctx, queryIsChallengeValid, nonce, fmt.Sprintf("%v", "now()"))
	if err := row.Scan(&valid); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}

	return valid, nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ query.WalletReader = (*WalletReader)(nil)
var _ query.ChallengeReader = (*ChallengeReader)(nil)