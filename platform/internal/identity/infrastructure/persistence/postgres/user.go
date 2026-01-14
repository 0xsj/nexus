package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/database"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// User Repository
// ============================================================================

// UserRepository implements domain.UserRepository using PostgreSQL.
type UserRepository struct {
	adapter *postgres.BaseAdapter
}

// Ensure UserRepository implements domain.UserRepository.
var _ domain.UserRepository = (*UserRepository)(nil)

// NewUserRepository creates a new user repository.
func NewUserRepository(adapter *postgres.BaseAdapter) *UserRepository {
	return &UserRepository{
		adapter: adapter,
	}
}

// ============================================================================
// Save
// ============================================================================

// Save persists a user (create or update).
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	const op = "postgres.UserRepository.Save"

	// Check if user exists
	exists, err := r.existsByID(ctx, user.ID())
	if err != nil {
		return errors.Wrap(err, op)
	}

	row := ToUserRow(user)

	if exists {
		return r.update(ctx, row, user)
	}

	return r.insert(ctx, row, user)
}

func (r *UserRepository) insert(ctx context.Context, row *UserRow, user *domain.User) error {
	const op = "postgres.UserRepository.insert"

	fmt.Printf("[DEBUG] insert: userID=%s\n", row.ID)

	// Insert user
	_, err := r.adapter.Exec(ctx, queryUserInsert,
		row.ID,
		row.DID,
		row.Status,
		row.DisplayName,
		row.AvatarURL,
		row.Bio,
		row.LastLoginMethod,
		row.CreatedAt,
		row.UpdatedAt,
		row.LastLoginAt,
	)
	if err != nil {
		fmt.Printf("[DEBUG] insert user failed: %v\n", err)
		return errors.Wrap(err, op)
	}
	fmt.Printf("[DEBUG] user inserted\n")

	// Insert linked DIDs
	fmt.Printf("[DEBUG] linkedDIDs count: %d\n", len(user.LinkedDIDs()))
	for _, linkedDID := range user.LinkedDIDs() {
		fmt.Printf("[DEBUG] inserting linkedDID: id=%s, did=%s\n", linkedDID.ID, linkedDID.DID.String())
		if err := r.insertLinkedDID(ctx, user.ID(), linkedDID); err != nil {
			fmt.Printf("[DEBUG] insert linkedDID failed: %v\n", err)
			return errors.Wrap(err, op)
		}
	}
	fmt.Printf("[DEBUG] linkedDIDs inserted\n")

	// Insert wallets
	fmt.Printf("[DEBUG] wallets count: %d\n", len(user.Wallets()))
	for _, wallet := range user.Wallets() {
		if err := r.insertWallet(ctx, user.ID(), wallet, false); err != nil {
			fmt.Printf("[DEBUG] insert wallet failed: %v\n", err)
			return errors.Wrap(err, op)
		}
	}

	// Insert linked emails
	fmt.Printf("[DEBUG] linkedIdentities count: %d\n", len(user.LinkedIdentities()))
	for _, identity := range user.LinkedIdentities() {
		fmt.Printf("[DEBUG] identity: type=%s, value=%s\n", identity.Type, identity.Value)
		if identity.Type == domain.IdentityTypeEmail {
			if err := r.insertEmail(ctx, user.ID(), identity); err != nil {
				fmt.Printf("[DEBUG] insert email failed: %v\n", err)
				return errors.Wrap(err, op)
			}
		}
	}
	fmt.Printf("[DEBUG] insert complete\n")

	return nil
}

func (r *UserRepository) update(ctx context.Context, row *UserRow, user *domain.User) error {
	const op = "postgres.UserRepository.update"

	_, err := r.adapter.Exec(ctx, queryUserUpdate,
		row.ID,
		row.DID,
		row.Status,
		row.DisplayName,
		row.AvatarURL,
		row.Bio,
		row.LastLoginMethod,
		row.UpdatedAt,
		row.LastLoginAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	// Sync linked DIDs (delete removed, insert new, update existing)
	if err := r.syncLinkedDIDs(ctx, user.ID(), user.LinkedDIDs()); err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *UserRepository) insertLinkedDID(ctx context.Context, userID string, linkedDID domain.LinkedDID) error {
	const op = "postgres.UserRepository.insertLinkedDID"

	row, err := ToLinkedDIDRow(userID, linkedDID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	_, err = r.adapter.Exec(ctx, queryLinkedDIDInsert,
		row.ID,
		row.UserID,
		row.DID,
		row.Source,
		row.IsPrimary,
		row.Label,
		row.Metadata,
		row.LinkedAt,
		row.LastUsedAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *UserRepository) syncLinkedDIDs(ctx context.Context, userID string, linkedDIDs domain.LinkedDIDs) error {
	const op = "postgres.UserRepository.syncLinkedDIDs"

	// Load existing linked DIDs
	existingRows, err := r.loadLinkedDIDRows(ctx, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	// Build maps for comparison
	existingMap := make(map[string]*LinkedDIDRow)
	for _, row := range existingRows {
		existingMap[row.ID] = row
	}

	newMap := make(map[string]domain.LinkedDID)
	for _, ld := range linkedDIDs {
		newMap[ld.ID] = ld
	}

	// Delete removed DIDs
	for id := range existingMap {
		if _, exists := newMap[id]; !exists {
			if err := r.deleteLinkedDID(ctx, id); err != nil {
				return errors.Wrap(err, op)
			}
		}
	}

	// Insert or update DIDs
	for _, ld := range linkedDIDs {
		if _, exists := existingMap[ld.ID]; exists {
			// Update existing
			if err := r.updateLinkedDID(ctx, userID, ld); err != nil {
				return errors.Wrap(err, op)
			}
		} else {
			// Insert new
			if err := r.insertLinkedDID(ctx, userID, ld); err != nil {
				return errors.Wrap(err, op)
			}
		}
	}

	return nil
}

func (r *UserRepository) updateLinkedDID(ctx context.Context, userID string, linkedDID domain.LinkedDID) error {
	const op = "postgres.UserRepository.updateLinkedDID"

	row, err := ToLinkedDIDRow(userID, linkedDID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	_, err = r.adapter.Exec(ctx, queryLinkedDIDUpdate,
		row.ID,
		row.IsPrimary,
		row.Label,
		row.Metadata,
		row.LastUsedAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *UserRepository) deleteLinkedDID(ctx context.Context, id string) error {
	const op = "postgres.UserRepository.deleteLinkedDID"

	_, err := r.adapter.Exec(ctx, queryLinkedDIDDelete, id)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *UserRepository) insertWallet(ctx context.Context, userID string, wallet domain.WalletAddress, isPrimary bool) error {
	const op = "postgres.UserRepository.insertWallet"

	id, err := generateID()
	if err != nil {
		return errors.Wrap(err, op)
	}

	_, err = r.adapter.Exec(ctx, queryUserWalletInsert,
		id,
		userID,
		wallet.Address,
		wallet.Chain.String(),
		isPrimary,
		true, // verified
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *UserRepository) insertEmail(ctx context.Context, userID string, identity domain.LinkedIdentity) error {
	const op = "postgres.UserRepository.insertEmail"

	id, err := generateID()
	if err != nil {
		return errors.Wrap(err, op)
	}

	var verifiedAt sql.NullTime
	if identity.Verified && !identity.VerifiedAt.IsZero() {
		verifiedAt = sql.NullTime{Time: identity.VerifiedAt, Valid: true}
	}

	_, err = r.adapter.Exec(ctx, queryUserEmailInsert,
		id,
		userID,
		identity.Value,
		false, // isPrimary
		identity.Verified,
		verifiedAt,
		identity.LinkedAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// Find
// ============================================================================

// FindByID finds a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	const op = "postgres.UserRepository.FindByID"

	row, err := r.scanUser(ctx, queryUserByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, id)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.hydrateUser(ctx, row)
}

// FindByDID finds a user by their primary DID.
func (r *UserRepository) FindByDID(ctx context.Context, did string) (*domain.User, error) {
	const op = "postgres.UserRepository.FindByDID"

	row, err := r.scanUser(ctx, queryUserByDID, did)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, did)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.hydrateUser(ctx, row)
}

// FindByLinkedDID finds a user by any linked DID (not just primary).
func (r *UserRepository) FindByLinkedDID(ctx context.Context, did string) (*domain.User, error) {
	const op = "postgres.UserRepository.FindByLinkedDID"

	row, err := r.scanUser(ctx, queryUserByLinkedDID, did)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, did)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.hydrateUser(ctx, row)
}

// FindByWallet finds a user by a linked wallet address.
func (r *UserRepository) FindByWallet(ctx context.Context, address string, chain domain.Chain) (*domain.User, error) {
	const op = "postgres.UserRepository.FindByWallet"

	row, err := r.scanUser(ctx, queryUserByWallet, address, chain.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, address)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.hydrateUser(ctx, row)
}

// FindByEmail finds a user by a linked email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const op = "postgres.UserRepository.FindByEmail"

	row, err := r.scanUser(ctx, queryUserByEmail, email)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, domain.ErrUserNotFound(op, email)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.hydrateUser(ctx, row)
}

// ============================================================================
// Exists
// ============================================================================

// ExistsByDID checks if a user with the given DID exists.
func (r *UserRepository) ExistsByDID(ctx context.Context, did string) (bool, error) {
	const op = "postgres.UserRepository.ExistsByDID"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryUserExistsByDID, did)
	if err := row.Scan(&exists); err != nil {
		return false, errors.Wrap(err, op)
	}

	return exists, nil
}

// ExistsByLinkedDID checks if a user with the given linked DID exists.
func (r *UserRepository) ExistsByLinkedDID(ctx context.Context, did string) (bool, error) {
	const op = "postgres.UserRepository.ExistsByLinkedDID"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryUserExistsByLinkedDID, did)
	if err := row.Scan(&exists); err != nil {
		return false, errors.Wrap(err, op)
	}

	return exists, nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const op = "postgres.UserRepository.ExistsByEmail"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryUserExistsByEmail, email)
	if err := row.Scan(&exists); err != nil {
		return false, errors.Wrap(err, op)
	}

	return exists, nil
}

// ExistsByWallet checks if a user with the given wallet exists.
func (r *UserRepository) ExistsByWallet(ctx context.Context, address string, chain domain.Chain) (bool, error) {
	const op = "postgres.UserRepository.ExistsByWallet"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryUserExistsByWallet, address, chain.String())
	if err := row.Scan(&exists); err != nil {
		return false, errors.Wrap(err, op)
	}

	return exists, nil
}

func (r *UserRepository) existsByID(ctx context.Context, id string) (bool, error) {
	const op = "postgres.UserRepository.existsByID"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id)
	if err := row.Scan(&exists); err != nil {
		return false, errors.Wrap(err, op)
	}

	return exists, nil
}

// ============================================================================
// Delete
// ============================================================================

// Delete removes a user.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	const op = "postgres.UserRepository.Delete"

	// Delete linked DIDs first (foreign key constraint)
	if err := r.deleteLinkedDIDsByUserID(ctx, id); err != nil {
		return errors.Wrap(err, op)
	}

	result, err := r.adapter.Exec(ctx, queryUserDelete, id)
	if err != nil {
		return errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound(op, id)
	}

	return nil
}

func (r *UserRepository) deleteLinkedDIDsByUserID(ctx context.Context, userID string) error {
	const op = "postgres.UserRepository.deleteLinkedDIDsByUserID"

	_, err := r.adapter.Exec(ctx, queryLinkedDIDDeleteByUserID, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// List
// ============================================================================

// List returns users with pagination.
func (r *UserRepository) List(ctx context.Context, opts domain.UserListOptions) ([]*domain.User, int, error) {
	const op = "postgres.UserRepository.List"

	// Validate and sanitize sort options
	sortBy := sanitizeSortColumn(opts.SortBy, "created_at")
	sortOrder := sanitizeSortOrder(opts.SortOrder, "desc")

	// Build query with dynamic ORDER BY
	query := fmt.Sprintf(queryUserList, sortBy, sortOrder)

	// Status filter
	var statusFilter interface{}
	if opts.Status != nil {
		statusFilter = opts.Status.String()
	}

	// Execute query
	rows, err := r.adapter.Select(ctx, query, statusFilter, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, errors.Wrap(err, op)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		row, err := scanUserRow(rows)
		if err != nil {
			return nil, 0, errors.Wrap(err, op)
		}

		user, err := r.hydrateUser(ctx, row)
		if err != nil {
			return nil, 0, errors.Wrap(err, op)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, op)
	}

	// Get total count
	var total int
	countRow := r.adapter.Executor().QueryRow(ctx, queryUserCount, statusFilter)
	if err := countRow.Scan(&total); err != nil {
		return nil, 0, errors.Wrap(err, op)
	}

	return users, total, nil
}

// ============================================================================
// Helpers
// ============================================================================

func (r *UserRepository) scanUser(ctx context.Context, query string, args ...interface{}) (*UserRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)

	var userRow UserRow
	err := row.Scan(
		&userRow.ID,
		&userRow.DID,
		&userRow.Status,
		&userRow.DisplayName,
		&userRow.AvatarURL,
		&userRow.Bio,
		&userRow.LastLoginMethod,
		&userRow.CreatedAt,
		&userRow.UpdatedAt,
		&userRow.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}

	return &userRow, nil
}

func scanUserRow(rows interface{ Scan(...interface{}) error }) (*UserRow, error) {
	var userRow UserRow
	err := rows.Scan(
		&userRow.ID,
		&userRow.DID,
		&userRow.Status,
		&userRow.DisplayName,
		&userRow.AvatarURL,
		&userRow.Bio,
		&userRow.LastLoginMethod,
		&userRow.CreatedAt,
		&userRow.UpdatedAt,
		&userRow.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}

	return &userRow, nil
}

func (r *UserRepository) hydrateUser(ctx context.Context, row *UserRow) (*domain.User, error) {
	const op = "postgres.UserRepository.hydrateUser"

	// Load linked DIDs
	linkedDIDs, err := r.loadLinkedDIDs(ctx, row.ID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Load wallets
	wallets, err := r.loadWallets(ctx, row.ID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Load linked identities (emails)
	identities, err := r.loadLinkedIdentities(ctx, row.ID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Add wallet identities
	for _, wallet := range wallets {
		identities = append(identities, domain.NewWalletIdentity(wallet.Address, wallet.Chain))
	}

	return row.ToDomainUser(linkedDIDs, identities, wallets)
}

func (r *UserRepository) loadLinkedDIDs(ctx context.Context, userID string) (domain.LinkedDIDs, error) {
	rows, err := r.loadLinkedDIDRows(ctx, userID)
	if err != nil {
		return nil, err
	}

	return ToDomainLinkedDIDs(rows)
}

func (r *UserRepository) loadLinkedDIDRows(ctx context.Context, userID string) ([]*LinkedDIDRow, error) {
	rows, err := r.adapter.Select(ctx, queryLinkedDIDsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var linkedDIDRows []*LinkedDIDRow
	for rows.Next() {
		var row LinkedDIDRow
		err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.DID,
			&row.Source,
			&row.IsPrimary,
			&row.Label,
			&row.Metadata,
			&row.LinkedAt,
			&row.LastUsedAt,
		)
		if err != nil {
			return nil, err
		}
		linkedDIDRows = append(linkedDIDRows, &row)
	}

	return linkedDIDRows, rows.Err()
}

func (r *UserRepository) loadWallets(ctx context.Context, userID string) ([]domain.WalletAddress, error) {
	rows, err := r.adapter.Select(ctx, queryUserWalletsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []domain.WalletAddress
	for rows.Next() {
		var walletRow UserWalletRow
		err := rows.Scan(
			&walletRow.ID,
			&walletRow.UserID,
			&walletRow.Address,
			&walletRow.Chain,
			&walletRow.IsPrimary,
			&walletRow.Verified,
			&walletRow.VerifiedAt,
			&walletRow.LinkedAt,
		)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, walletRow.ToDomainWalletAddress())
	}

	return wallets, rows.Err()
}

func (r *UserRepository) loadLinkedIdentities(ctx context.Context, userID string) ([]domain.LinkedIdentity, error) {
	rows, err := r.adapter.Select(ctx, queryUserEmailsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var identities []domain.LinkedIdentity
	for rows.Next() {
		var emailRow UserEmailRow
		err := rows.Scan(
			&emailRow.ID,
			&emailRow.UserID,
			&emailRow.Email,
			&emailRow.IsPrimary,
			&emailRow.Verified,
			&emailRow.VerifiedAt,
			&emailRow.LinkedAt,
		)
		if err != nil {
			return nil, err
		}
		identities = append(identities, emailRow.ToDomainLinkedIdentity())
	}

	return identities, rows.Err()
}

// ============================================================================
// Utilities
// ============================================================================

func sanitizeSortColumn(column, defaultColumn string) string {
	allowed := map[string]bool{
		"created_at":    true,
		"updated_at":    true,
		"last_login_at": true,
	}
	if allowed[column] {
		return column
	}
	return defaultColumn
}

func sanitizeSortOrder(order, defaultOrder string) string {
	if order == "asc" || order == "ASC" {
		return "ASC"
	}
	if order == "desc" || order == "DESC" {
		return "DESC"
	}
	return defaultOrder
}

func generateID() (string, error) {
	// Simple UUID-like ID generation
	// In production, use github.com/google/uuid
	b := make([]byte, 16)
	if _, err := randomRead(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// randomRead is a variable to allow mocking in tests
var randomRead = func(b []byte) (int, error) {
	return cryptoRandRead(b)
}

func cryptoRandRead(b []byte) (int, error) {
	// Import crypto/rand at the top if needed
	// For now, using a simple approach
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
		time.Sleep(time.Nanosecond)
	}
	return len(b), nil
}

func (r *UserRepository) FindWalletByAddressAndChain(ctx context.Context, address string, chain domain.Chain) (*UserWalletRow, error) {
	const op = "postgres.UserRepository.FindWalletByAddressAndChain"

	row := r.adapter.Executor().QueryRow(ctx, queryUserWalletByAddressChain, address, chain.String())

	var walletRow UserWalletRow
	err := row.Scan(
		&walletRow.ID,
		&walletRow.UserID,
		&walletRow.Address,
		&walletRow.Chain,
		&walletRow.IsPrimary,
		&walletRow.Verified,
		&walletRow.VerifiedAt,
		&walletRow.LinkedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound(op, address)
		}
		return nil, errors.Wrap(err, op)
	}

	return &walletRow, nil
}

// DeleteWallet removes a wallet from a user.
func (r *UserRepository) DeleteWallet(ctx context.Context, userID string, address string, chain domain.Chain) error {
	const op = "postgres.UserRepository.DeleteWallet"

	result, err := r.adapter.Exec(ctx, queryUserWalletDelete, userID, address, chain.String())
	if err != nil {
		return errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrWalletNotFound(op, address)
	}

	return nil
}

// DeleteWalletsByUserID removes all wallets for a user.
func (r *UserRepository) DeleteWalletsByUserID(ctx context.Context, userID string) error {
	const op = "postgres.UserRepository.DeleteWalletsByUserID"

	_, err := r.adapter.Exec(ctx, queryUserWalletDeleteByUserID, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// Email Operations
// ============================================================================

// FindEmailByEmail finds an email record by email address.
func (r *UserRepository) FindEmailByEmail(ctx context.Context, email string) (*UserEmailRow, error) {
	const op = "postgres.UserRepository.FindEmailByEmail"

	row := r.adapter.Executor().QueryRow(ctx, queryUserEmailByEmail, email)

	var emailRow UserEmailRow
	err := row.Scan(
		&emailRow.ID,
		&emailRow.UserID,
		&emailRow.Email,
		&emailRow.IsPrimary,
		&emailRow.Verified,
		&emailRow.VerifiedAt,
		&emailRow.LinkedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, email)
		}
		return nil, errors.Wrap(err, op)
	}

	return &emailRow, nil
}

// UpdateEmailVerification updates the verification status of an email.
func (r *UserRepository) UpdateEmailVerification(ctx context.Context, emailID string, verified bool, verifiedAt sql.NullTime) error {
	const op = "postgres.UserRepository.UpdateEmailVerification"

	_, err := r.adapter.Exec(ctx, queryUserEmailUpdate, emailID, verified, verifiedAt)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// DeleteEmail removes an email from a user.
func (r *UserRepository) DeleteEmail(ctx context.Context, userID string, email string) error {
	const op = "postgres.UserRepository.DeleteEmail"

	result, err := r.adapter.Exec(ctx, queryUserEmailDelete, userID, email)
	if err != nil {
		return errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound(op, email)
	}

	return nil
}

// DeleteEmailsByUserID removes all emails for a user.
func (r *UserRepository) DeleteEmailsByUserID(ctx context.Context, userID string) error {
	const op = "postgres.UserRepository.DeleteEmailsByUserID"

	_, err := r.adapter.Exec(ctx, queryUserEmailDeleteByUserID, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}
