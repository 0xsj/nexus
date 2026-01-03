package postgres

// ============================================================================
// User Queries
// ============================================================================

const (
	// Insert
	queryUserInsert = `
		INSERT INTO users (id, did, status, display_name, avatar_url, bio, created_at, updated_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Update
	queryUserUpdate = `
		UPDATE users
		SET did = $2, status = $3, display_name = $4, avatar_url = $5, bio = $6, updated_at = $7, last_login_at = $8
		WHERE id = $1`

	// Select
	queryUserByID = `
		SELECT id, did, status, display_name, avatar_url, bio, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1`

	queryUserByDID = `
		SELECT id, did, status, display_name, avatar_url, bio, created_at, updated_at, last_login_at
		FROM users
		WHERE did = $1`

	queryUserByEmail = `
		SELECT u.id, u.did, u.status, u.display_name, u.avatar_url, u.bio, u.created_at, u.updated_at, u.last_login_at
		FROM users u
		INNER JOIN user_emails e ON u.id = e.user_id
		WHERE e.email = $1`

	queryUserByWallet = `
		SELECT u.id, u.did, u.status, u.display_name, u.avatar_url, u.bio, u.created_at, u.updated_at, u.last_login_at
		FROM users u
		INNER JOIN user_wallets w ON u.id = w.user_id
		WHERE w.address = $1 AND w.chain = $2`

	// Exists
	queryUserExistsByDID = `
		SELECT EXISTS(SELECT 1 FROM users WHERE did = $1)`

	queryUserExistsByEmail = `
		SELECT EXISTS(SELECT 1 FROM user_emails WHERE email = $1)`

	queryUserExistsByWallet = `
		SELECT EXISTS(SELECT 1 FROM user_wallets WHERE address = $1 AND chain = $2)`

	// Delete
	queryUserDelete = `
		DELETE FROM users WHERE id = $1`

	// List
	queryUserList = `
		SELECT id, did, status, display_name, avatar_url, bio, created_at, updated_at, last_login_at
		FROM users
		WHERE ($1::text IS NULL OR status = $1)
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`

	queryUserCount = `
		SELECT COUNT(*) FROM users
		WHERE ($1::text IS NULL OR status = $1)`
)

// ============================================================================
// User Wallet Queries
// ============================================================================

const (
	queryUserWalletInsert = `
		INSERT INTO user_wallets (id, user_id, address, chain, is_primary, verified, verified_at, linked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	queryUserWalletsByUserID = `
		SELECT id, user_id, address, chain, is_primary, verified, verified_at, linked_at
		FROM user_wallets
		WHERE user_id = $1
		ORDER BY is_primary DESC, linked_at ASC`

	queryUserWalletByAddressChain = `
		SELECT id, user_id, address, chain, is_primary, verified, verified_at, linked_at
		FROM user_wallets
		WHERE address = $1 AND chain = $2`

	queryUserWalletDelete = `
		DELETE FROM user_wallets WHERE user_id = $1 AND address = $2 AND chain = $3`

	queryUserWalletDeleteByUserID = `
		DELETE FROM user_wallets WHERE user_id = $1`
)

// ============================================================================
// User Email Queries
// ============================================================================

const (
	queryUserEmailInsert = `
		INSERT INTO user_emails (id, user_id, email, is_primary, verified, verified_at, linked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	queryUserEmailsByUserID = `
		SELECT id, user_id, email, is_primary, verified, verified_at, linked_at
		FROM user_emails
		WHERE user_id = $1
		ORDER BY is_primary DESC, linked_at ASC`

	queryUserEmailByEmail = `
		SELECT id, user_id, email, is_primary, verified, verified_at, linked_at
		FROM user_emails
		WHERE email = $1`

	queryUserEmailUpdate = `
		UPDATE user_emails
		SET verified = $2, verified_at = $3
		WHERE id = $1`

	queryUserEmailDelete = `
		DELETE FROM user_emails WHERE user_id = $1 AND email = $2`

	queryUserEmailDeleteByUserID = `
		DELETE FROM user_emails WHERE user_id = $1`
)

// ============================================================================
// Session Queries
// ============================================================================

const (
	querySessionInsert = `
		INSERT INTO sessions (id, user_id, token_hash, auth_method, ip_address, user_agent, device_id, status, created_at, expires_at, last_active_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	querySessionUpdate = `
		UPDATE sessions
		SET status = $2, last_active_at = $3, revoked_at = $4
		WHERE id = $1`

	querySessionByID = `
		SELECT id, user_id, token_hash, auth_method, ip_address, user_agent, device_id, status, created_at, expires_at, last_active_at, revoked_at
		FROM sessions
		WHERE id = $1`

	querySessionByTokenHash = `
		SELECT id, user_id, token_hash, auth_method, ip_address, user_agent, device_id, status, created_at, expires_at, last_active_at, revoked_at
		FROM sessions
		WHERE token_hash = $1`

	querySessionsByUserID = `
		SELECT id, user_id, token_hash, auth_method, ip_address, user_agent, device_id, status, created_at, expires_at, last_active_at, revoked_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC`

	querySessionsActiveByUserID = `
		SELECT id, user_id, token_hash, auth_method, ip_address, user_agent, device_id, status, created_at, expires_at, last_active_at, revoked_at
		FROM sessions
		WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY created_at DESC`

	querySessionDelete = `
		DELETE FROM sessions WHERE id = $1`

	querySessionDeleteByUserID = `
		DELETE FROM sessions WHERE user_id = $1`

	querySessionDeleteExpired = `
		DELETE FROM sessions WHERE expires_at < NOW() AND status = 'active'`

	querySessionCountByUserID = `
		SELECT COUNT(*) FROM sessions WHERE user_id = $1`

	querySessionCountActiveByUserID = `
		SELECT COUNT(*) FROM sessions WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()`
)

// ============================================================================
// Connection Queries
// ============================================================================

const (
	queryConnectionInsert = `
		INSERT INTO connections (
			id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar,
			profile_data, access_token, refresh_token, token_expires_at, scopes, status, last_synced_at,
			connected_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	queryConnectionUpdate = `
		UPDATE connections
		SET provider_username = $2, provider_email = $3, provider_avatar = $4, profile_data = $5,
			access_token = $6, refresh_token = $7, token_expires_at = $8, scopes = $9, status = $10,
			last_synced_at = $11, sync_error = $12, updated_at = $13, disconnected_at = $14
		WHERE id = $1`

	queryConnectionByID = `
		SELECT id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar,
			profile_data, access_token, refresh_token, token_expires_at, scopes, status, last_synced_at,
			sync_error, connected_at, updated_at, disconnected_at
		FROM connections
		WHERE id = $1`

	queryConnectionByUserAndProvider = `
		SELECT id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar,
			profile_data, access_token, refresh_token, token_expires_at, scopes, status, last_synced_at,
			sync_error, connected_at, updated_at, disconnected_at
		FROM connections
		WHERE user_id = $1 AND provider = $2`

	queryConnectionByProviderUserID = `
		SELECT id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar,
			profile_data, access_token, refresh_token, token_expires_at, scopes, status, last_synced_at,
			sync_error, connected_at, updated_at, disconnected_at
		FROM connections
		WHERE provider = $1 AND provider_user_id = $2`

	queryConnectionsByUserID = `
		SELECT id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar,
			profile_data, access_token, refresh_token, token_expires_at, scopes, status, last_synced_at,
			sync_error, connected_at, updated_at, disconnected_at
		FROM connections
		WHERE user_id = $1
		ORDER BY connected_at DESC`

	queryConnectionsActiveByUserID = `
		SELECT id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar,
			profile_data, access_token, refresh_token, token_expires_at, scopes, status, last_synced_at,
			sync_error, connected_at, updated_at, disconnected_at
		FROM connections
		WHERE user_id = $1 AND status = 'active'
		ORDER BY connected_at DESC`

	queryConnectionDelete = `
		DELETE FROM connections WHERE id = $1`

	queryConnectionDeleteByUserID = `
		DELETE FROM connections WHERE user_id = $1`

	queryConnectionExistsByUserAndProvider = `
		SELECT EXISTS(SELECT 1 FROM connections WHERE user_id = $1 AND provider = $2)`

	queryConnectionsStale = `
		SELECT id, user_id, provider, provider_user_id, provider_username, provider_email, provider_avatar,
			profile_data, access_token, refresh_token, token_expires_at, scopes, status, last_synced_at,
			sync_error, connected_at, updated_at, disconnected_at
		FROM connections
		WHERE status = 'active' AND (last_synced_at IS NULL OR last_synced_at < $1)
		ORDER BY last_synced_at ASC NULLS FIRST
		LIMIT $2`
)

// ============================================================================
// API Key Queries
// ============================================================================

const (
	queryAPIKeyInsert = `
		INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, scopes, rate_limit, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	queryAPIKeyUpdate = `
		UPDATE api_keys
		SET name = $2, scopes = $3, rate_limit = $4, status = $5, last_used_at = $6, usage_count = $7, revoked_at = $8
		WHERE id = $1`

	queryAPIKeyByID = `
		SELECT id, user_id, name, key_hash, key_prefix, scopes, rate_limit, status, last_used_at, usage_count, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE id = $1`

	queryAPIKeyByKeyHash = `
		SELECT id, user_id, name, key_hash, key_prefix, scopes, rate_limit, status, last_used_at, usage_count, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE key_hash = $1`

	queryAPIKeysByKeyPrefix = `
		SELECT id, user_id, name, key_hash, key_prefix, scopes, rate_limit, status, last_used_at, usage_count, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE key_prefix = $1
		ORDER BY created_at DESC`

	queryAPIKeysByUserID = `
		SELECT id, user_id, name, key_hash, key_prefix, scopes, rate_limit, status, last_used_at, usage_count, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC`

	queryAPIKeysActiveByUserID = `
		SELECT id, user_id, name, key_hash, key_prefix, scopes, rate_limit, status, last_used_at, usage_count, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE user_id = $1 AND status = 'active' AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC`

	queryAPIKeyDelete = `
		DELETE FROM api_keys WHERE id = $1`

	queryAPIKeyDeleteByUserID = `
		DELETE FROM api_keys WHERE user_id = $1`

	queryAPIKeyCountByUserID = `
		SELECT COUNT(*) FROM api_keys WHERE user_id = $1`

	queryAPIKeyCountActiveByUserID = `
		SELECT COUNT(*) FROM api_keys WHERE user_id = $1 AND status = 'active' AND (expires_at IS NULL OR expires_at > NOW())`
)

// ============================================================================
// Refresh Token Queries
// ============================================================================

const (
	queryRefreshTokenInsert = `
		INSERT INTO refresh_tokens (id, user_id, session_id, token_hash, revoked, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	queryRefreshTokenUpdate = `
		UPDATE refresh_tokens
		SET revoked = $2, revoked_at = $3
		WHERE id = $1`

	queryRefreshTokenByID = `
		SELECT id, user_id, session_id, token_hash, revoked, created_at, expires_at, revoked_at
		FROM refresh_tokens
		WHERE id = $1`

	queryRefreshTokenByTokenHash = `
		SELECT id, user_id, session_id, token_hash, revoked, created_at, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1`

	queryRefreshTokensByUserID = `
		SELECT id, user_id, session_id, token_hash, revoked, created_at, expires_at, revoked_at
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC`

	queryRefreshTokenDelete = `
		DELETE FROM refresh_tokens WHERE id = $1`

	queryRefreshTokenDeleteByTokenHash = `
		DELETE FROM refresh_tokens WHERE token_hash = $1`

	queryRefreshTokenDeleteByUserID = `
		DELETE FROM refresh_tokens WHERE user_id = $1`

	queryRefreshTokenDeleteExpired = `
		DELETE FROM refresh_tokens WHERE expires_at < NOW()`
)
