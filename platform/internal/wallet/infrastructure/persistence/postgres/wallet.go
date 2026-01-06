package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
)

// ============================================================================
// Wallet Repository
// ============================================================================

// WalletRepository implements domain.WalletRepository using PostgreSQL.
type WalletRepository struct {
	db postgres.DB
}

// NewWalletRepository creates a new PostgreSQL wallet repository.
func NewWalletRepository(db postgres.DB) *WalletRepository {
	return &WalletRepository{db: db}
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
		return err
	}

	if exists {
		// Update existing wallet
		_, err = r.db.ExecContext(ctx, queryUpdateWallet,
			row.Label,
			row.IsPrimary,
			row.Status,
			row.UpdatedAt,
			row.LastUsedAt,
			row.ID,
		)
	} else {
		// Insert new wallet
		_, err = r.db.ExecContext(ctx, queryInsertWallet,
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
		return domain.ErrWalletNotFound(op, wallet.ID())
	}

	return nil
}

// Delete deletes a wallet from the database.
func (r *WalletRepository) Delete(ctx context.Context, id string) error {
	const op = "WalletRepository.Delete"

	result, err := r.db.ExecContext(ctx, queryDeleteWallet, id)
	if err != nil {
		return domain.ErrWalletNotFound(op, id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrWalletNotFound(op, id)
	}

	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// FindByID finds a wallet by ID.
func (r *WalletRepository) FindByID(ctx context.Context, id string) (*domain.Wallet, error) {
	const op = "WalletRepository.FindByID"

	var row walletRow
	err := r.db.QueryRowContext(ctx, querySelectWalletByID, id).Scan(
		&row.ID,
		&row.UserID,
		&row.Address,
		&row.AddressNormalized,
		&row.ChainID,
		&row.ChainFamily,
		&row.DID,
		&row.Label,
		&row.IsPrimary,
		&row.Status,
		&row.VerifiedAt,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.LastUsedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, id)
		}
		return nil, err
	}

	return row.toDomain()
}

// FindByAddress finds a wallet by address and chain.
func (r *WalletRepository) FindByAddress(ctx context.Context, address domain.Address) (*domain.Wallet, error) {
	const op = "WalletRepository.FindByAddress"

	var row walletRow
	err := r.db.QueryRowContext(ctx, querySelectWalletByAddress,
		address.Normalized(),
		address.ChainID().String(),
	).Scan(
		&row.ID,
		&row.UserID,
		&row.Address,
		&row.AddressNormalized,
		&row.ChainID,
		&row.ChainFamily,
		&row.DID,
		&row.Label,
		&row.IsPrimary,
		&row.Status,
		&row.VerifiedAt,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.LastUsedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, address.Normalized())
		}
		return nil, err
	}

	return row.toDomain()
}

// FindByDID finds a wallet by DID.
func (r *WalletRepository) FindByDID(ctx context.Context, didString string) (*domain.Wallet, error) {
	const op = "WalletRepository.FindByDID"

	var row walletRow
	err := r.db.QueryRowContext(ctx, querySelectWalletByDID, didString).Scan(
		&row.ID,
		&row.UserID,
		&row.Address,
		&row.AddressNormalized,
		&row.ChainID,
		&row.ChainFamily,
		&row.DID,
		&row.Label,
		&row.IsPrimary,
		&row.Status,
		&row.VerifiedAt,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.LastUsedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, didString)
		}
		return nil, err
	}

	return row.toDomain()
}

// FindByUserID finds all wallets for a user.
func (r *WalletRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Wallet, error) {
	rows, err := r.db.QueryContext(ctx, querySelectWalletsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanWallets(rows)
}

// FindByUserIDWithStatus finds wallets for a user with a specific status.
func (r *WalletRepository) FindByUserIDWithStatus(ctx context.Context, userID string, status domain.WalletStatus) ([]*domain.Wallet, error) {
	rows, err := r.db.QueryContext(ctx, querySelectWalletsByUserIDWithStatus, userID, status.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanWallets(rows)
}

// FindPrimaryByUserID finds the primary wallet for a user.
func (r *WalletRepository) FindPrimaryByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	const op = "WalletRepository.FindPrimaryByUserID"

	var row walletRow
	err := r.db.QueryRowContext(ctx, querySelectPrimaryWalletByUserID, userID).Scan(
		&row.ID,
		&row.UserID,
		&row.Address,
		&row.AddressNormalized,
		&row.ChainID,
		&row.ChainFamily,
		&row.DID,
		&row.Label,
		&row.IsPrimary,
		&row.Status,
		&row.VerifiedAt,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.LastUsedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, "primary wallet for user "+userID)
		}
		return nil, err
	}

	return row.toDomain()
}

// ============================================================================
// Existence Checks
// ============================================================================

// ExistsByID checks if a wallet exists by ID.
func (r *WalletRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	return r.existsByID(ctx, id)
}

func (r *WalletRepository) existsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, queryExistsWalletByID, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ExistsByAddress checks if a wallet exists by address.
func (r *WalletRepository) ExistsByAddress(ctx context.Context, address domain.Address) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, queryExistsWalletByAddress,
		address.Normalized(),
		address.ChainID().String(),
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ExistsByDID checks if a wallet exists by DID.
func (r *WalletRepository) ExistsByDID(ctx context.Context, didString string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, queryExistsWalletByDID, didString).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ============================================================================
// Count Operations
// ============================================================================

// CountByUserID counts wallets for a user.
func (r *WalletRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, queryCountWalletsByUserID, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountActiveByUserID counts active wallets for a user.
func (r *WalletRepository) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, queryCountActiveWalletsByUserID, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ============================================================================
// Primary Wallet Operations
// ============================================================================

// UnsetPrimaryForUser unsets the primary wallet for a user.
func (r *WalletRepository) UnsetPrimaryForUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, queryUnsetPrimaryForUser, userID)
	return err
}

// ============================================================================
// Helper Methods
// ============================================================================

// scanWallets scans multiple wallet rows.
func (r *WalletRepository) scanWallets(rows *sql.Rows) ([]*domain.Wallet, error) {
	var wallets []*domain.Wallet

	for rows.Next() {
		var row walletRow
		err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.Address,
			&row.AddressNormalized,
			&row.ChainID,
			&row.ChainFamily,
			&row.DID,
			&row.Label,
			&row.IsPrimary,
			&row.Status,
			&row.VerifiedAt,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.LastUsedAt,
		)
		if err != nil {
			return nil, err
		}

		wallet, err := row.toDomain()
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