package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/application/query"
	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Composite Reader
// ============================================================================

// Reader implements all query reader interfaces.
type Reader struct {
	adapter      *postgres.BaseAdapter
	tokenService domain.TokenService // For token validation
}

// Ensure Reader implements all interfaces.
var (
	_ query.UserReader       = (*Reader)(nil)
	_ query.SessionReader    = (*Reader)(nil)
	_ query.ConnectionReader = (*Reader)(nil)
	_ query.APIKeyReader     = (*Reader)(nil)
	_ query.TokenReader      = (*Reader)(nil)
)

// NewReader creates a new composite reader.
func NewReader(adapter *postgres.BaseAdapter, tokenService domain.TokenService) *Reader {
	return &Reader{
		adapter:      adapter,
		tokenService: tokenService,
	}
}

// NewReaderWithoutTokenService creates a reader without token validation.
// Token validation methods will return errors.
func NewReaderWithoutTokenService(adapter *postgres.BaseAdapter) *Reader {
	return &Reader{
		adapter: adapter,
	}
}

// ============================================================================
// User Reader Implementation
// ============================================================================

// GetUser retrieves a user by ID.
func (r *Reader) GetUser(ctx context.Context, userID string) (*query.UserView, error) {
	const op = "postgres.Reader.GetUser"

	row, err := r.scanUserRow(ctx, queryUserByID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, userID)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toUserView(ctx, row)
}

// GetUserByDID retrieves a user by DID.
func (r *Reader) GetUserByDID(ctx context.Context, did string) (*query.UserView, error) {
	const op = "postgres.Reader.GetUserByDID"

	row, err := r.scanUserRow(ctx, queryUserByDID, did)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, did)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toUserView(ctx, row)
}

// GetUserByWallet retrieves a user by wallet address.
func (r *Reader) GetUserByWallet(ctx context.Context, address string, chain domain.Chain) (*query.UserView, error) {
	const op = "postgres.Reader.GetUserByWallet"

	row, err := r.scanUserRow(ctx, queryUserByWallet, address, chain.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, address)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toUserView(ctx, row)
}

// GetUserByEmail retrieves a user by email.
func (r *Reader) GetUserByEmail(ctx context.Context, email string) (*query.UserView, error) {
	const op = "postgres.Reader.GetUserByEmail"

	row, err := r.scanUserRow(ctx, queryUserByEmail, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, email)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toUserView(ctx, row)
}

// ListUsers retrieves a paginated list of users.
func (r *Reader) ListUsers(ctx context.Context, opts query.ListUsersOptions) (*query.UserListView, error) {
	const op = "postgres.Reader.ListUsers"

	sortBy := sanitizeSortColumn(opts.SortBy, "created_at")
	sortOrder := sanitizeSortOrder(opts.SortOrder, "DESC")

	q := fmt.Sprintf(queryUserList, sortBy, sortOrder)

	var statusFilter interface{}
	if opts.Status != nil {
		statusFilter = opts.Status.String()
	}

	rows, err := r.adapter.Select(ctx, q, statusFilter, opts.Limit, opts.Offset)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var users []query.UserSummaryView
	for rows.Next() {
		var row UserRow
		err := rows.Scan(
			&row.ID,
			&row.DID,
			&row.Status,
			&row.DisplayName,
			&row.AvatarURL,
			&row.Bio,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.LastLoginAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		users = append(users, query.UserSummaryView{
			ID:          row.ID,
			PrimaryDID:  row.DID,
			Status:      row.Status,
			CreatedAt:   row.CreatedAt,
			LastLoginAt: nilTimePtr(row.LastLoginAt),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Get total count
	var total int
	countRow := r.adapter.Executor().QueryRow(ctx, queryUserCount, statusFilter)
	if err := countRow.Scan(&total); err != nil {
		return nil, errors.Wrap(err, op)
	}

	return &query.UserListView{
		Users:   users,
		Total:   total,
		Limit:   opts.Limit,
		Offset:  opts.Offset,
		HasMore: opts.Offset+len(users) < total,
	}, nil
}

// GetUserStats retrieves statistics for a user.
func (r *Reader) GetUserStats(ctx context.Context, userID string) (*query.UserStatsView, error) {
	const op = "postgres.Reader.GetUserStats"

	// Wallet count
	var walletCount int
	row := r.adapter.Executor().QueryRow(ctx, "SELECT COUNT(*) FROM user_wallets WHERE user_id = $1", userID)
	if err := row.Scan(&walletCount); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Connection count
	var connectionCount int
	row = r.adapter.Executor().QueryRow(ctx, "SELECT COUNT(*) FROM connections WHERE user_id = $1 AND status = 'active'", userID)
	if err := row.Scan(&connectionCount); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Active sessions
	var activeSessions int
	row = r.adapter.Executor().QueryRow(ctx, querySessionCountActiveByUserID, userID)
	if err := row.Scan(&activeSessions); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Active API keys
	var activeAPIKeys int
	row = r.adapter.Executor().QueryRow(ctx, queryAPIKeyCountActiveByUserID, userID)
	if err := row.Scan(&activeAPIKeys); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Total API key usage
	var totalUsage int64
	row = r.adapter.Executor().QueryRow(ctx, "SELECT COALESCE(SUM(usage_count), 0) FROM api_keys WHERE user_id = $1", userID)
	if err := row.Scan(&totalUsage); err != nil {
		return nil, errors.Wrap(err, op)
	}

	return &query.UserStatsView{
		UserID:           userID,
		WalletCount:      walletCount,
		ConnectionCount:  connectionCount,
		ActiveSessions:   activeSessions,
		ActiveAPIKeys:    activeAPIKeys,
		TotalAPIKeyUsage: totalUsage,
	}, nil
}

// GetPublicProfile retrieves the public profile for a DID.
func (r *Reader) GetPublicProfile(ctx context.Context, did string) (*query.PublicProfileView, error) {
	const op = "postgres.Reader.GetPublicProfile"

	row, err := r.scanUserRow(ctx, queryUserByDID, did)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound(op, did)
		}
		return nil, errors.Wrap(err, op)
	}

	// Load wallets
	wallets, err := r.loadWalletViews(ctx, row.ID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Load connections
	connections, err := r.loadPublicConnections(ctx, row.ID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Convert wallets to public format
	var publicWallets []query.PublicWallet
	for _, w := range wallets {
		publicWallets = append(publicWallets, query.PublicWallet{
			Chain: w.Chain,
			DID:   w.DID,
		})
	}

	return &query.PublicProfileView{
		DID:         row.DID,
		DisplayName: nullStringValue(row.DisplayName),
		AvatarURL:   nullStringValue(row.AvatarURL),
		Wallets:     publicWallets,
		Connections: connections,
		CreatedAt:   row.CreatedAt,
	}, nil
}

// SearchUsers searches for users.
func (r *Reader) SearchUsers(ctx context.Context, searchQuery string, opts query.ListOptions) (*query.UserListView, error) {
	const op = "postgres.Reader.SearchUsers"

	sortBy := sanitizeSortColumn(opts.SortBy, "created_at")
	sortOrder := sanitizeSortOrder(opts.SortOrder, "DESC")

	q := fmt.Sprintf(`
		SELECT id, did, status, display_name, avatar_url, bio, created_at, updated_at, last_login_at
		FROM users
		WHERE did ILIKE $1 OR display_name ILIKE $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortBy, sortOrder)

	searchPattern := "%" + searchQuery + "%"

	rows, err := r.adapter.Select(ctx, q, searchPattern, opts.Limit, opts.Offset)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var users []query.UserSummaryView
	for rows.Next() {
		var row UserRow
		err := rows.Scan(
			&row.ID,
			&row.DID,
			&row.Status,
			&row.DisplayName,
			&row.AvatarURL,
			&row.Bio,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.LastLoginAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		users = append(users, query.UserSummaryView{
			ID:          row.ID,
			PrimaryDID:  row.DID,
			Status:      row.Status,
			CreatedAt:   row.CreatedAt,
			LastLoginAt: nilTimePtr(row.LastLoginAt),
		})
	}

	// Get total count
	var total int
	countRow := r.adapter.Executor().QueryRow(ctx,
		"SELECT COUNT(*) FROM users WHERE did ILIKE $1 OR display_name ILIKE $1",
		searchPattern,
	)
	if err := countRow.Scan(&total); err != nil {
		return nil, errors.Wrap(err, op)
	}

	return &query.UserListView{
		Users:   users,
		Total:   total,
		Limit:   opts.Limit,
		Offset:  opts.Offset,
		HasMore: opts.Offset+len(users) < total,
	}, nil
}

// UserExists checks if a user exists.
func (r *Reader) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// UserExistsByDID checks if a user exists by DID.
func (r *Reader) UserExistsByDID(ctx context.Context, did string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryUserExistsByDID, did)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// UserExistsByWallet checks if a user exists by wallet.
func (r *Reader) UserExistsByWallet(ctx context.Context, address string, chain domain.Chain) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryUserExistsByWallet, address, chain.String())
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// UserExistsByEmail checks if a user exists by email.
func (r *Reader) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryUserExistsByEmail, email)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// ============================================================================
// Session Reader Implementation
// ============================================================================

// GetSession retrieves a session by ID.
func (r *Reader) GetSession(ctx context.Context, sessionID string) (*query.SessionView, error) {
	const op = "postgres.Reader.GetSession"

	row, err := r.scanSessionRow(ctx, querySessionByID, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound(op, sessionID)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toSessionView(row, ""), nil
}

// GetSessionByTokenHash retrieves a session by token hash.
func (r *Reader) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*query.SessionView, error) {
	const op = "postgres.Reader.GetSessionByTokenHash"

	row, err := r.scanSessionRow(ctx, querySessionByTokenHash, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound(op, tokenHash)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toSessionView(row, ""), nil
}

// ListUserSessions retrieves sessions for a user.
func (r *Reader) ListUserSessions(ctx context.Context, userID string, opts query.ListSessionsOptions) (*query.SessionListView, error) {
	const op = "postgres.Reader.ListUserSessions"

	var q string
	if opts.ActiveOnly {
		q = querySessionsActiveByUserID
	} else {
		q = querySessionsByUserID
	}

	rows, err := r.adapter.Select(ctx, q, userID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var sessions []query.SessionSummaryView
	for rows.Next() {
		var row SessionRow
		err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.TokenHash,
			&row.AuthMethod,
			&row.IPAddress,
			&row.UserAgent,
			&row.DeviceID,
			&row.Status,
			&row.CreatedAt,
			&row.ExpiresAt,
			&row.LastActiveAt,
			&row.RevokedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		sessions = append(sessions, query.SessionSummaryView{
			ID:         row.ID,
			Method:     row.AuthMethod,
			Status:     row.Status,
			Device:     nullStringValue(row.DeviceID),
			LastSeenAt: row.LastActiveAt,
			IsCurrent:  row.ID == opts.CurrentSession,
		})
	}

	// Apply pagination in-memory (could optimize with LIMIT/OFFSET in query)
	total := len(sessions)
	start := opts.Offset
	end := opts.Offset + opts.Limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginated := sessions[start:end]

	return &query.SessionListView{
		Sessions: paginated,
		Total:    total,
		Limit:    opts.Limit,
		Offset:   opts.Offset,
		HasMore:  end < total,
	}, nil
}

// CountUserSessions counts sessions for a user.
func (r *Reader) CountUserSessions(ctx context.Context, userID string, activeOnly bool) (int, error) {
	var count int
	var q string
	if activeOnly {
		q = querySessionCountActiveByUserID
	} else {
		q = querySessionCountByUserID
	}
	row := r.adapter.Executor().QueryRow(ctx, q, userID)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// SessionExists checks if a session exists.
func (r *Reader) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1)", sessionID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// ValidateSession validates a session.
func (r *Reader) ValidateSession(ctx context.Context, sessionID string, tokenHash string) (*query.SessionView, error) {
	const op = "postgres.Reader.ValidateSession"

	row, err := r.scanSessionRow(ctx, querySessionByID, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound(op, sessionID)
		}
		return nil, errors.Wrap(err, op)
	}

	// Validate token hash
	if row.TokenHash != tokenHash {
		return nil, domain.ErrTokenInvalid(op, "token mismatch")
	}

	// Validate status
	if row.Status != "active" {
		return nil, domain.ErrSessionRevoked(op, sessionID)
	}

	return r.toSessionView(row, ""), nil
}

// ============================================================================
// Connection Reader Implementation
// ============================================================================

// GetConnection retrieves a connection by ID.
func (r *Reader) GetConnection(ctx context.Context, connectionID string) (*query.ConnectionView, error) {
	const op = "postgres.Reader.GetConnection"

	row, err := r.scanConnectionRow(ctx, queryConnectionByID, connectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConnectionNotFound(op, "", connectionID)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toConnectionView(row), nil
}

// GetConnectionByProvider retrieves a connection by provider.
func (r *Reader) GetConnectionByProvider(ctx context.Context, userID string, provider domain.OAuthProvider) (*query.ConnectionView, error) {
	const op = "postgres.Reader.GetConnectionByProvider"

	row, err := r.scanConnectionRow(ctx, queryConnectionByUserAndProvider, userID, provider.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConnectionNotFound(op, provider.String(), userID)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toConnectionView(row), nil
}

// ListUserConnections retrieves connections for a user.
func (r *Reader) ListUserConnections(ctx context.Context, userID string, opts query.ListConnectionsOptions) (*query.ConnectionListView, error) {
	const op = "postgres.Reader.ListUserConnections"

	var q string
	if opts.ActiveOnly {
		q = queryConnectionsActiveByUserID
	} else {
		q = queryConnectionsByUserID
	}

	rows, err := r.adapter.Select(ctx, q, userID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var connections []query.ConnectionSummaryView
	for rows.Next() {
		row, err := scanConnectionRowFromRows(rows)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		connections = append(connections, query.ConnectionSummaryView{
			ID:          row.ID,
			Provider:    row.Provider,
			Username:    nullStringValue(row.ProviderUsername),
			Status:      row.Status,
			ConnectedAt: row.ConnectedAt,
		})
	}

	total := len(connections)
	start := opts.Offset
	end := opts.Offset + opts.Limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginated := connections[start:end]

	return &query.ConnectionListView{
		Connections: paginated,
		Total:       total,
		Limit:       opts.Limit,
		Offset:      opts.Offset,
		HasMore:     end < total,
	}, nil
}

// CountUserConnections counts connections for a user.
func (r *Reader) CountUserConnections(ctx context.Context, userID string, activeOnly bool) (int, error) {
	var count int
	var q string
	if activeOnly {
		q = "SELECT COUNT(*) FROM connections WHERE user_id = $1 AND status = 'active'"
	} else {
		q = "SELECT COUNT(*) FROM connections WHERE user_id = $1"
	}
	row := r.adapter.Executor().QueryRow(ctx, q, userID)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// ConnectionExists checks if a connection exists.
func (r *Reader) ConnectionExists(ctx context.Context, userID string, provider domain.OAuthProvider) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryConnectionExistsByUserAndProvider, userID, provider.String())
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// ============================================================================
// API Key Reader Implementation
// ============================================================================

// GetAPIKey retrieves an API key by ID.
func (r *Reader) GetAPIKey(ctx context.Context, keyID string) (*query.APIKeyView, error) {
	const op = "postgres.Reader.GetAPIKey"

	row, err := r.scanAPIKeyRow(ctx, queryAPIKeyByID, keyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAPIKeyNotFound(op, keyID)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toAPIKeyView(row), nil
}

// GetAPIKeyByHash retrieves an API key by hash.
func (r *Reader) GetAPIKeyByHash(ctx context.Context, keyHash string) (*query.APIKeyView, error) {
	const op = "postgres.Reader.GetAPIKeyByHash"

	row, err := r.scanAPIKeyRow(ctx, queryAPIKeyByKeyHash, keyHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAPIKeyNotFound(op, keyHash)
		}
		return nil, errors.Wrap(err, op)
	}

	return r.toAPIKeyView(row), nil
}

// ListUserAPIKeys retrieves API keys for a user.
func (r *Reader) ListUserAPIKeys(ctx context.Context, userID string, opts query.ListAPIKeysOptions) (*query.APIKeyListView, error) {
	const op = "postgres.Reader.ListUserAPIKeys"

	var q string
	if opts.ActiveOnly {
		q = queryAPIKeysActiveByUserID
	} else {
		q = queryAPIKeysByUserID
	}

	rows, err := r.adapter.Select(ctx, q, userID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var apiKeys []query.APIKeySummaryView
	for rows.Next() {
		row, err := scanAPIKeyRowFromRows(rows)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		apiKeys = append(apiKeys, query.APIKeySummaryView{
			ID:         row.ID,
			Name:       row.Name,
			Prefix:     row.KeyPrefix,
			Status:     row.Status,
			ExpiresAt:  nilTimePtr(row.ExpiresAt),
			LastUsedAt: nilTimePtr(row.LastUsedAt),
		})
	}

	total := len(apiKeys)
	start := opts.Offset
	end := opts.Offset + opts.Limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginated := apiKeys[start:end]

	return &query.APIKeyListView{
		APIKeys: paginated,
		Total:   total,
		Limit:   opts.Limit,
		Offset:  opts.Offset,
		HasMore: end < total,
	}, nil
}

// CountUserAPIKeys counts API keys for a user.
func (r *Reader) CountUserAPIKeys(ctx context.Context, userID string, activeOnly bool) (int, error) {
	var count int
	var q string
	if activeOnly {
		q = queryAPIKeyCountActiveByUserID
	} else {
		q = queryAPIKeyCountByUserID
	}
	row := r.adapter.Executor().QueryRow(ctx, q, userID)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// ValidateAPIKey validates an API key.
func (r *Reader) ValidateAPIKey(ctx context.Context, keyHash string) (*query.APIKeyValidationView, error) {
	const op = "postgres.Reader.ValidateAPIKey"

	row, err := r.scanAPIKeyRow(ctx, queryAPIKeyByKeyHash, keyHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.APIKeyValidationView{
				Valid: false,
				Error: "api key not found",
			}, nil
		}
		return nil, errors.Wrap(err, op)
	}

	// Check status
	if row.Status != "active" {
		return &query.APIKeyValidationView{
			Valid: false,
			Error: "api key revoked",
		}, nil
	}

	// Check expiration
	if row.ExpiresAt.Valid && row.ExpiresAt.Time.Before(timeNow()) {
		return &query.APIKeyValidationView{
			Valid: false,
			Error: "api key expired",
		}, nil
	}

	// Get user DID
	var did string
	userRow := r.adapter.Executor().QueryRow(ctx, "SELECT did FROM users WHERE id = $1", row.UserID)
	_ = userRow.Scan(&did)

	return &query.APIKeyValidationView{
		Valid:   true,
		UserID:  row.UserID,
		DID:     did,
		KeyID:   row.ID,
		KeyName: row.Name,
		Scopes:  row.Scopes,
	}, nil
}

// ============================================================================
// Token Reader Implementation
// ============================================================================

// ValidateAccessToken validates an access token.
func (r *Reader) ValidateAccessToken(ctx context.Context, token string) (*query.TokenValidationView, error) {
	const op = "postgres.Reader.ValidateAccessToken"

	if r.tokenService == nil {
		return nil, errors.Internal(op, nil).WithMessage("token service not configured")
	}

	claims, err := r.tokenService.ValidateAccessToken(ctx, token)
	if err != nil {
		return &query.TokenValidationView{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &query.TokenValidationView{
		Valid:     true,
		UserID:    claims.UserID,
		DID:       claims.DID,
		SessionID: claims.SessionID,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

// ValidateRefreshToken validates a refresh token.
func (r *Reader) ValidateRefreshToken(ctx context.Context, token string) (*query.TokenValidationView, error) {
	const op = "postgres.Reader.ValidateRefreshToken"

	if r.tokenService == nil {
		return nil, errors.Internal(op, nil).WithMessage("token service not configured")
	}

	claims, err := r.tokenService.ValidateRefreshToken(ctx, token)
	if err != nil {
		return &query.TokenValidationView{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &query.TokenValidationView{
		Valid:     true,
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

// IsTokenRevoked checks if a token is revoked.
func (r *Reader) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	var revoked bool
	row := r.adapter.Executor().QueryRow(ctx,
		"SELECT revoked FROM refresh_tokens WHERE id = $1",
		tokenID,
	)
	if err := row.Scan(&revoked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil // Not found = revoked
		}
		return false, err
	}
	return revoked, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (r *Reader) scanUserRow(ctx context.Context, query string, args ...interface{}) (*UserRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)

	var userRow UserRow
	err := row.Scan(
		&userRow.ID,
		&userRow.DID,
		&userRow.Status,
		&userRow.DisplayName,
		&userRow.AvatarURL,
		&userRow.Bio,
		&userRow.CreatedAt,
		&userRow.UpdatedAt,
		&userRow.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}

	return &userRow, nil
}

func (r *Reader) scanSessionRow(ctx context.Context, query string, args ...interface{}) (*SessionRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)

	var sessionRow SessionRow
	err := row.Scan(
		&sessionRow.ID,
		&sessionRow.UserID,
		&sessionRow.TokenHash,
		&sessionRow.AuthMethod,
		&sessionRow.IPAddress,
		&sessionRow.UserAgent,
		&sessionRow.DeviceID,
		&sessionRow.Status,
		&sessionRow.CreatedAt,
		&sessionRow.ExpiresAt,
		&sessionRow.LastActiveAt,
		&sessionRow.RevokedAt,
	)
	if err != nil {
		return nil, err
	}

	return &sessionRow, nil
}

func (r *Reader) scanConnectionRow(ctx context.Context, query string, args ...interface{}) (*ConnectionRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)
	return scanConnectionRowFromRow(row)
}

func scanConnectionRowFromRow(row interface{ Scan(...interface{}) error }) (*ConnectionRow, error) {
	var connRow ConnectionRow
	err := row.Scan(
		&connRow.ID,
		&connRow.UserID,
		&connRow.Provider,
		&connRow.ProviderUserID,
		&connRow.ProviderUsername,
		&connRow.ProviderEmail,
		&connRow.ProviderAvatar,
		&connRow.ProfileData,
		&connRow.AccessToken,
		&connRow.RefreshToken,
		&connRow.TokenExpiresAt,
		&connRow.Scopes,
		&connRow.Status,
		&connRow.LastSyncedAt,
		&connRow.SyncError,
		&connRow.ConnectedAt,
		&connRow.UpdatedAt,
		&connRow.DisconnectedAt,
	)
	if err != nil {
		return nil, err
	}
	return &connRow, nil
}

func scanConnectionRowFromRows(rows interface{ Scan(...interface{}) error }) (*ConnectionRow, error) {
	return scanConnectionRowFromRow(rows)
}

func (r *Reader) scanAPIKeyRow(ctx context.Context, query string, args ...interface{}) (*APIKeyRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)
	return scanAPIKeyRowFromRow(row)
}

func scanAPIKeyRowFromRow(row interface{ Scan(...interface{}) error }) (*APIKeyRow, error) {
	var apiKeyRow APIKeyRow
	err := row.Scan(
		&apiKeyRow.ID,
		&apiKeyRow.UserID,
		&apiKeyRow.Name,
		&apiKeyRow.KeyHash,
		&apiKeyRow.KeyPrefix,
		&apiKeyRow.Scopes,
		&apiKeyRow.RateLimit,
		&apiKeyRow.Status,
		&apiKeyRow.LastUsedAt,
		&apiKeyRow.UsageCount,
		&apiKeyRow.CreatedAt,
		&apiKeyRow.ExpiresAt,
		&apiKeyRow.RevokedAt,
	)
	if err != nil {
		return nil, err
	}
	return &apiKeyRow, nil
}

func scanAPIKeyRowFromRows(rows interface{ Scan(...interface{}) error }) (*APIKeyRow, error) {
	return scanAPIKeyRowFromRow(rows)
}

func (r *Reader) toUserView(ctx context.Context, row *UserRow) (*query.UserView, error) {
	wallets, err := r.loadWalletViews(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	email, err := r.loadEmailView(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	connections, err := r.loadConnectionViews(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	return &query.UserView{
		ID:          row.ID,
		PrimaryDID:  row.DID,
		Status:      row.Status,
		Wallets:     wallets,
		LinkedEmail: email,
		Connections: connections,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		LastLoginAt: nilTimePtr(row.LastLoginAt),
	}, nil
}

func (r *Reader) toSessionView(row *SessionRow, currentSessionID string) *query.SessionView {
	return &query.SessionView{
		ID:         row.ID,
		UserID:     row.UserID,
		Method:     row.AuthMethod,
		Status:     row.Status,
		UserAgent:  nullStringValue(row.UserAgent),
		IPAddress:  nullStringValue(row.IPAddress),
		Device:     nullStringValue(row.DeviceID),
		CreatedAt:  row.CreatedAt,
		ExpiresAt:  row.ExpiresAt,
		LastSeenAt: row.LastActiveAt,
		IsCurrent:  row.ID == currentSessionID,
	}
}

func (r *Reader) toConnectionView(row *ConnectionRow) *query.ConnectionView {
	return &query.ConnectionView{
		ID:           row.ID,
		UserID:       row.UserID,
		Provider:     row.Provider,
		ProviderName: providerDisplayName(row.Provider),
		Username:     nullStringValue(row.ProviderUsername),
		Email:        nullStringValue(row.ProviderEmail),
		AvatarURL:    nullStringValue(row.ProviderAvatar),
		Status:       row.Status,
		ConnectedAt:  row.ConnectedAt,
		LastSyncedAt: nilTimePtr(row.LastSyncedAt),
	}
}

func (r *Reader) toAPIKeyView(row *APIKeyRow) *query.APIKeyView {
	return &query.APIKeyView{
		ID:         row.ID,
		UserID:     row.UserID,
		Name:       row.Name,
		Prefix:     row.KeyPrefix,
		Scopes:     row.Scopes,
		Status:     row.Status,
		ExpiresAt:  nilTimePtr(row.ExpiresAt),
		LastUsedAt: nilTimePtr(row.LastUsedAt),
		UsageCount: row.UsageCount,
		CreatedAt:  row.CreatedAt,
	}
}

func (r *Reader) loadWalletViews(ctx context.Context, userID string) ([]query.WalletView, error) {
	rows, err := r.adapter.Select(ctx, queryUserWalletsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []query.WalletView
	for rows.Next() {
		var row UserWalletRow
		err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.Address,
			&row.Chain,
			&row.IsPrimary,
			&row.Verified,
			&row.VerifiedAt,
			&row.LinkedAt,
		)
		if err != nil {
			return nil, err
		}

		wallet := domain.NewWalletAddress(row.Address, domain.Chain(row.Chain))
		wallets = append(wallets, query.WalletView{
			Address:  row.Address,
			Chain:    row.Chain,
			DID:      wallet.ToDID(),
			LinkedAt: row.LinkedAt,
		})
	}

	return wallets, rows.Err()
}

func (r *Reader) loadEmailView(ctx context.Context, userID string) (*query.EmailView, error) {
	rows, err := r.adapter.Select(ctx, queryUserEmailsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var row UserEmailRow
		err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.Email,
			&row.IsPrimary,
			&row.Verified,
			&row.VerifiedAt,
			&row.LinkedAt,
		)
		if err != nil {
			return nil, err
		}

		return &query.EmailView{
			Email:      row.Email,
			Verified:   row.Verified,
			VerifiedAt: nilTimePtr(row.VerifiedAt),
			LinkedAt:   row.LinkedAt,
		}, nil
	}

	// No email found
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *Reader) loadConnectionViews(ctx context.Context, userID string) ([]query.ConnectionView, error) {
	rows, err := r.adapter.Select(ctx, queryConnectionsActiveByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []query.ConnectionView
	for rows.Next() {
		row, err := scanConnectionRowFromRows(rows)
		if err != nil {
			return nil, err
		}

		connections = append(connections, *r.toConnectionView(row))
	}

	return connections, rows.Err()
}

func (r *Reader) loadPublicConnections(ctx context.Context, userID string) ([]query.PublicConnection, error) {
	rows, err := r.adapter.Select(ctx, queryConnectionsActiveByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []query.PublicConnection
	for rows.Next() {
		row, err := scanConnectionRowFromRows(rows)
		if err != nil {
			return nil, err
		}

		connections = append(connections, query.PublicConnection{
			Provider:    row.Provider,
			Username:    nullStringValue(row.ProviderUsername),
			ConnectedAt: row.ConnectedAt,
		})
	}

	return connections, rows.Err()
}

// ============================================================================
// Utility Functions
// ============================================================================

func nullStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nilTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func providerDisplayName(provider string) string {
	names := map[string]string{
		"github":   "GitHub",
		"google":   "Google",
		"linkedin": "LinkedIn",
		"twitter":  "Twitter",
		"discord":  "Discord",
	}
	if name, ok := names[provider]; ok {
		return name
	}
	return provider
}

// timeNow is a variable to allow mocking in tests.
var timeNow = time.Now
