package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Wallet Repository
// ============================================================================

// WalletRepository implements domain.WalletRepository using PostgreSQL.
type WalletRepository struct {
	adapter *postgres.BaseAdapter
}

// NewWalletRepository creates a new PostgreSQL wallet repository.
func NewWalletRepository(adapter *postgres.BaseAdapter) *WalletRepository {
	return &WalletRepository{adapter: adapter}
}

// ============================================================================
// Write Operations
// ============================================================================

// Save saves a wallet to the database.
func (r *WalletRepository) Save(ctx context.Context, wallet *domain.Wallet) error {
	const op = "WalletRepository.Save"

	row := walletToRow(wallet)

	// Check if wallet exists
	exists, err := r.existsByID(ctx, wallet.ID())
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	if exists {
		// Update existing wallet
		_, err = r.adapter.Exec(ctx, queryUpdateWallet,
			row.Label,
			row.IsPrimary,
			row.Status,
			row.UpdatedAt,
			row.LastUsedAt,
			row.ID,
		)
	} else {
		// Insert new wallet
		_, err = r.adapter.Exec(ctx, queryInsertWallet,
			row.ID,
			row.UserID,
			row.Address,
			row.AddressNormalized,
			row.ChainID,
			row.ChainFamily,
			row.DID,
			row.Label,
			row.IsPrimary,
			row.Status,
			row.VerifiedAt,
			row.CreatedAt,
			row.UpdatedAt,
			row.LastUsedAt,
		)
	}

	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	return nil
}

// Delete deletes a wallet from the database.
func (r *WalletRepository) Delete(ctx context.Context, id string) error {
	const op = "WalletRepository.Delete"

	result, err := r.adapter.Exec(ctx, queryDeleteWallet, id)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrWalletNotFound(op, id)
	}

	return nil
}

// DeleteByUserID deletes all wallets for a user.
func (r *WalletRepository) DeleteByUserID(ctx context.Context, userID string) error {
	const op = "WalletRepository.DeleteByUserID"

	_, err := r.adapter.Exec(ctx, queryDeleteWalletsByUserID, userID)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// FindByID finds a wallet by ID.
func (r *WalletRepository) FindByID(ctx context.Context, id string) (*domain.Wallet, error) {
	const op = "WalletRepository.FindByID"

	row := r.adapter.Executor().QueryRow(ctx, querySelectWalletByID, id)

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, id)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toDomain()
}

// FindByAddress finds a wallet by address and chain.
func (r *WalletRepository) FindByAddress(ctx context.Context, address domain.Address) (*domain.Wallet, error) {
	const op = "WalletRepository.FindByAddress"

	row := r.adapter.Executor().QueryRow(ctx, querySelectWalletByAddress,
		address.Normalized(),
		address.ChainID().String(),
	)

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, address.Normalized())
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toDomain()
}

// FindByAddressString finds a wallet by address string and chain ID.
func (r *WalletRepository) FindByAddressString(ctx context.Context, address string, chainID domain.ChainID) (*domain.Wallet, error) {
	const op = "WalletRepository.FindByAddressString"

	row := r.adapter.Executor().QueryRow(ctx, querySelectWalletByAddress,
		address,
		chainID.String(),
	)

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, address)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toDomain()
}

// FindByDID finds a wallet by DID.
func (r *WalletRepository) FindByDID(ctx context.Context, didString string) (*domain.Wallet, error) {
	const op = "WalletRepository.FindByDID"

	row := r.adapter.Executor().QueryRow(ctx, querySelectWalletByDID, didString)

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, didString)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toDomain()
}

// FindByUserID finds all wallets for a user.
func (r *WalletRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Wallet, error) {
	const op = "WalletRepository.FindByUserID"

	rows, err := r.adapter.Select(ctx, querySelectWalletsByUserID, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	defer rows.Close()

	return r.scanWallets(rows)
}

// FindPrimaryByUserID finds the primary wallet for a user.
func (r *WalletRepository) FindPrimaryByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	const op = "WalletRepository.FindPrimaryByUserID"

	row := r.adapter.Executor().QueryRow(ctx, querySelectPrimaryWalletByUserID, userID)

	walletRow, err := scanWalletRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, "primary wallet for user "+userID)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return walletRow.toDomain()
}

// ============================================================================
// Existence Checks
// ============================================================================

// ExistsByAddress checks if a wallet exists by address.
func (r *WalletRepository) ExistsByAddress(ctx context.Context, address domain.Address) (bool, error) {
	const op = "WalletRepository.ExistsByAddress"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExistsWalletByAddress,
		address.Normalized(),
		address.ChainID().String(),
	)
	if err := row.Scan(&exists); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}
	return exists, nil
}

// ExistsByAddressString checks if a wallet exists by address string.
func (r *WalletRepository) ExistsByAddressString(ctx context.Context, address string, chainID domain.ChainID) (bool, error) {
	const op = "WalletRepository.ExistsByAddressString"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExistsWalletByAddress,
		address,
		chainID.String(),
	)
	if err := row.Scan(&exists); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}
	return exists, nil
}

func (r *WalletRepository) existsByID(ctx context.Context, id string) (bool, error) {
	const op = "WalletRepository.existsByID"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExistsWalletByID, id)
	if err := row.Scan(&exists); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}
	return exists, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// scanWalletRow scans a single wallet row.
func scanWalletRow(row interface{ Scan(...interface{}) error }) (*walletRow, error) {
	var r walletRow
	err := row.Scan(
		&r.ID,
		&r.UserID,
		&r.Address,
		&r.AddressNormalized,
		&r.ChainID,
		&r.ChainFamily,
		&r.DID,
		&r.Label,
		&r.IsPrimary,
		&r.Status,
		&r.VerifiedAt,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.LastUsedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// scanWallets scans multiple wallet rows.
func (r *WalletRepository) scanWallets(rows interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
}) ([]*domain.Wallet, error) {
	var wallets []*domain.Wallet

	for rows.Next() {
		walletRow, err := scanWalletRow(rows)
		if err != nil {
			return nil, err
		}

		wallet, err := walletRow.toDomain()
		if err != nil {
			return nil, err
		}

		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.WalletRepository = (*WalletRepository)(nil)
