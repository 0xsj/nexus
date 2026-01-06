package postgres

// ============================================================================
// Wallet Queries
// ============================================================================

const (
	// Insert wallet
	queryInsertWallet = `
		INSERT INTO wallets (
			id, user_id, address, address_normalized, chain_id, chain_family,
			did, label, is_primary, status, verified_at, created_at, updated_at, last_used_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	// Update wallet
	queryUpdateWallet = `
		UPDATE wallets SET
			label = $1,
			is_primary = $2,
			status = $3,
			updated_at = $4,
			last_used_at = $5
		WHERE id = $6`

	// Delete wallet
	queryDeleteWallet = `
		DELETE FROM wallets WHERE id = $1`

	// Add to queries.go

	// Delete wallets by user ID
	queryDeleteWalletsByUserID = `
	DELETE FROM wallets WHERE user_id = $1`

	// Select wallet by ID
	querySelectWalletByID = `
		SELECT
			id, user_id, address, address_normalized, chain_id, chain_family,
			did, label, is_primary, status, verified_at, created_at, updated_at, last_used_at
		FROM wallets
		WHERE id = $1`

	// Select wallet by address and chain
	querySelectWalletByAddress = `
		SELECT
			id, user_id, address, address_normalized, chain_id, chain_family,
			did, label, is_primary, status, verified_at, created_at, updated_at, last_used_at
		FROM wallets
		WHERE address_normalized = $1 AND chain_id = $2`

	// Select wallet by DID
	querySelectWalletByDID = `
		SELECT
			id, user_id, address, address_normalized, chain_id, chain_family,
			did, label, is_primary, status, verified_at, created_at, updated_at, last_used_at
		FROM wallets
		WHERE did = $1`

	// Select wallets by user ID
	querySelectWalletsByUserID = `
		SELECT
			id, user_id, address, address_normalized, chain_id, chain_family,
			did, label, is_primary, status, verified_at, created_at, updated_at, last_used_at
		FROM wallets
		WHERE user_id = $1
		ORDER BY is_primary DESC, created_at ASC`

	// Select wallets by user ID with status filter
	querySelectWalletsByUserIDWithStatus = `
		SELECT
			id, user_id, address, address_normalized, chain_id, chain_family,
			did, label, is_primary, status, verified_at, created_at, updated_at, last_used_at
		FROM wallets
		WHERE user_id = $1 AND status = $2
		ORDER BY is_primary DESC, created_at ASC`

	// Select wallets by chain ID
	querySelectWalletsByChainID = `
		SELECT
			id, user_id, address, address_normalized, chain_id, chain_family,
			did, label, is_primary, status, verified_at, created_at, updated_at, last_used_at
		FROM wallets
		WHERE chain_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	// Select primary wallet by user ID
	querySelectPrimaryWalletByUserID = `
		SELECT
			id, user_id, address, address_normalized, chain_id, chain_family,
			did, label, is_primary, status, verified_at, created_at, updated_at, last_used_at
		FROM wallets
		WHERE user_id = $1 AND is_primary = true
		LIMIT 1`

	// Check wallet exists by ID
	queryExistsWalletByID = `
		SELECT EXISTS(SELECT 1 FROM wallets WHERE id = $1)`

	// Check wallet exists by address and chain
	queryExistsWalletByAddress = `
		SELECT EXISTS(SELECT 1 FROM wallets WHERE address_normalized = $1 AND chain_id = $2)`

	// Check wallet exists by DID
	queryExistsWalletByDID = `
		SELECT EXISTS(SELECT 1 FROM wallets WHERE did = $1)`

	// Count wallets by user ID
	queryCountWalletsByUserID = `
		SELECT COUNT(*) FROM wallets WHERE user_id = $1`

	// Count active wallets by user ID
	queryCountActiveWalletsByUserID = `
		SELECT COUNT(*) FROM wallets WHERE user_id = $1 AND status = 'active'`

	// Count wallets by chain ID
	queryCountWalletsByChainID = `
		SELECT COUNT(*) FROM wallets WHERE chain_id = $1`

	// Unset primary for user (used before setting new primary)
	queryUnsetPrimaryForUser = `
		UPDATE wallets SET is_primary = false, updated_at = $1 WHERE user_id = $2 AND is_primary = true`
)

// ============================================================================
// Challenge Queries
// ============================================================================

const (
	// Insert challenge
	queryInsertChallenge = `
		INSERT INTO wallet_challenges (
			nonce, message, address, address_normalized, chain_id, domain, uri,
			issued_at, expires_at, used, used_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	// Update challenge (mark as used)
	queryUpdateChallenge = `
		UPDATE wallet_challenges SET
			used = $1,
			used_at = $2
		WHERE nonce = $3`

	// Select challenge by nonce
	querySelectChallengeByNonce = `
		SELECT
			nonce, message, address, address_normalized, chain_id, domain, uri,
			issued_at, expires_at, used, used_at
		FROM wallet_challenges
		WHERE nonce = $1`

	// Delete challenge by nonce
	queryDeleteChallenge = `
		DELETE FROM wallet_challenges WHERE nonce = $1`

	// Delete expired challenges
	queryDeleteExpiredChallenges = `
		DELETE FROM wallet_challenges WHERE expires_at < $1`

	// Check challenge exists
	queryExistsChallengeByNonce = `
		SELECT EXISTS(SELECT 1 FROM wallet_challenges WHERE nonce = $1)`

	// Check challenge is valid (exists, not used, not expired)
	queryIsChallengeValid = `
		SELECT EXISTS(
			SELECT 1 FROM wallet_challenges 
			WHERE nonce = $1 AND used = false AND expires_at > $2
		)`
)

// ============================================================================
// Nonce Queries
// ============================================================================

const (
	// Insert nonce
	queryInsertNonce = `
		INSERT INTO wallet_nonces (
			nonce, created_at, expires_at, used, used_at
		) VALUES (
			$1, $2, $3, $4, $5
		)`

	// Update nonce (mark as used)
	queryUpdateNonce = `
		UPDATE wallet_nonces SET
			used = $1,
			used_at = $2
		WHERE nonce = $3`

	// Select nonce
	querySelectNonce = `
		SELECT nonce, created_at, expires_at, used, used_at
		FROM wallet_nonces
		WHERE nonce = $1`

	// Delete nonce
	queryDeleteNonce = `
		DELETE FROM wallet_nonces WHERE nonce = $1`

	// Delete expired nonces
	queryDeleteExpiredNonces = `
		DELETE FROM wallet_nonces WHERE expires_at < $1`

	// Check nonce exists
	queryExistsNonce = `
		SELECT EXISTS(SELECT 1 FROM wallet_nonces WHERE nonce = $1)`

	// Check nonce is valid (exists, not used, not expired)
	queryIsNonceValid = `
		SELECT EXISTS(
			SELECT 1 FROM wallet_nonces 
			WHERE nonce = $1 AND used = false AND expires_at > $2
		)`

	// Check nonce is used
	queryIsNonceUsed = `
		SELECT used FROM wallet_nonces WHERE nonce = $1`
)

// ============================================================================
// Wallet Activity Queries
// ============================================================================

const (
	// Insert wallet activity
	queryInsertWalletActivity = `
		INSERT INTO wallet_activities (
			id, wallet_id, action, ip_address, user_agent, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`

	// Select wallet activities
	querySelectWalletActivities = `
		SELECT id, wallet_id, action, ip_address, user_agent, created_at
		FROM wallet_activities
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	// Select user wallet activities
	querySelectUserWalletActivities = `
		SELECT wa.id, wa.wallet_id, wa.action, wa.ip_address, wa.user_agent, wa.created_at
		FROM wallet_activities wa
		JOIN wallets w ON wa.wallet_id = w.id
		WHERE w.user_id = $1
		ORDER BY wa.created_at DESC
		LIMIT $2 OFFSET $3`

	// Count wallet activities
	queryCountWalletActivities = `
		SELECT COUNT(*) FROM wallet_activities WHERE wallet_id = $1`
)

// ============================================================================
// Wallet Statistics Queries
// ============================================================================

const (
	// Get user wallet stats
	queryUserWalletStats = `
		SELECT
			COUNT(*) as total_count,
			COUNT(*) FILTER (WHERE status = 'active') as active_count
		FROM wallets
		WHERE user_id = $1`

	// Get chain stats
	queryChainStats = `
		SELECT
			COUNT(*) as total_count,
			COUNT(*) FILTER (WHERE status = 'active') as active_count
		FROM wallets
		WHERE chain_id = $1`

	// Get wallet count by chain for user
	queryWalletCountByChainForUser = `
		SELECT chain_id, COUNT(*) as count
		FROM wallets
		WHERE user_id = $1
		GROUP BY chain_id`
)
