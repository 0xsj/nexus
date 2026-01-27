package postgres

// ============================================================================
// Verification Queries
// ============================================================================

const (
	// Insert verification
	queryInsertVerification = `
		INSERT INTO verifications (
			id, user_id, provider, credential_type, status, oauth_state,
			redirect_url, expires_at, initiated_at, version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
		)`

	// Update verification
	queryUpdateVerification = `
		UPDATE verifications SET
			status = $2,
			oauth_state = $3,
			provider_user_id = $4,
			username = $5,
			email = $6,
			display_name = $7,
			avatar_url = $8,
			profile_url = $9,
			raw_data = $10,
			credential_id = $11,
			failure_reason = $12,
			failure_code = $13,
			authorized_at = $14,
			completed_at = $15,
			failed_at = $16,
			expires_at = $17,
			version = $18,
			updated_at = NOW()
		WHERE id = $1 AND version = $18 - 1`

	// Select verification by ID
	querySelectVerificationByID = `
		SELECT 
			id, user_id, provider, credential_type, status, oauth_state,
			provider_user_id, username, email, display_name, avatar_url, profile_url,
			raw_data, credential_id, failure_reason, failure_code, redirect_url,
			initiated_at, authorized_at, completed_at, failed_at, expires_at,
			version, created_at, updated_at
		FROM verifications
		WHERE id = $1`

	// Select verification by OAuth state
	querySelectVerificationByOAuthState = `
		SELECT 
			id, user_id, provider, credential_type, status, oauth_state,
			provider_user_id, username, email, display_name, avatar_url, profile_url,
			raw_data, credential_id, failure_reason, failure_code, redirect_url,
			initiated_at, authorized_at, completed_at, failed_at, expires_at,
			version, created_at, updated_at
		FROM verifications
		WHERE oauth_state = $1`

	// Select verifications by user and provider
	querySelectVerificationsByUserAndProvider = `
		SELECT 
			id, user_id, provider, credential_type, status, oauth_state,
			provider_user_id, username, email, display_name, avatar_url, profile_url,
			raw_data, credential_id, failure_reason, failure_code, redirect_url,
			initiated_at, authorized_at, completed_at, failed_at, expires_at,
			version, created_at, updated_at
		FROM verifications
		WHERE user_id = $1 AND provider = $2
		ORDER BY initiated_at DESC`

	// Select active verifications by user
	querySelectActiveVerificationsByUser = `
		SELECT 
			id, user_id, provider, credential_type, status, oauth_state,
			provider_user_id, username, email, display_name, avatar_url, profile_url,
			raw_data, credential_id, failure_reason, failure_code, redirect_url,
			initiated_at, authorized_at, completed_at, failed_at, expires_at,
			version, created_at, updated_at
		FROM verifications
		WHERE user_id = $1 AND status IN ('pending', 'authorized', 'fetching')
		ORDER BY initiated_at DESC`

	// Select pending expired verifications
	querySelectPendingExpired = `
		SELECT 
			id, user_id, provider, credential_type, status, oauth_state,
			provider_user_id, username, email, display_name, avatar_url, profile_url,
			raw_data, credential_id, failure_reason, failure_code, redirect_url,
			initiated_at, authorized_at, completed_at, failed_at, expires_at,
			version, created_at, updated_at
		FROM verifications
		WHERE status IN ('pending', 'authorized', 'fetching') AND expires_at < $1
		ORDER BY expires_at ASC
		LIMIT 100`

	// Check verification exists
	queryVerificationExists = `
		SELECT EXISTS(SELECT 1 FROM verifications WHERE id = $1)`

	// Delete verification
	queryDeleteVerification = `
		DELETE FROM verifications WHERE id = $1`
)

// ============================================================================
// OAuth State Queries
// ============================================================================

const (
	// Insert OAuth state
	queryInsertOAuthState = `
		INSERT INTO oauth_states (
			state, verification_id, user_id, provider,
			code_verifier, nonce, redirect_url, created_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, NOW(), $8
		)`

	// Select OAuth state by value
	querySelectOAuthStateByValue = `
		SELECT 
			state, verification_id, user_id, provider,
			code_verifier, nonce, redirect_url, created_at, expires_at, used_at
		FROM oauth_states
		WHERE state = $1`

	// Mark OAuth state as used
	queryMarkOAuthStateUsed = `
		UPDATE oauth_states SET used_at = NOW() WHERE state = $1`

	// Delete OAuth state
	queryDeleteOAuthState = `
		DELETE FROM oauth_states WHERE state = $1`

	// Delete expired OAuth states
	queryDeleteExpiredOAuthStates = `
		DELETE FROM oauth_states WHERE expires_at < NOW()`

	// Check OAuth state exists
	queryOAuthStateExists = `
		SELECT EXISTS(SELECT 1 FROM oauth_states WHERE state = $1 AND used_at IS NULL)`
)

// ============================================================================
// Provider Token Queries
// ============================================================================

const (
	// Upsert provider tokens
	queryUpsertProviderTokens = `
		INSERT INTO provider_tokens (
			user_id, provider, access_token, refresh_token, token_type,
			scopes, expires_at, provider_user_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
		)
		ON CONFLICT (user_id, provider) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = COALESCE(EXCLUDED.refresh_token, provider_tokens.refresh_token),
			token_type = EXCLUDED.token_type,
			scopes = EXCLUDED.scopes,
			expires_at = EXCLUDED.expires_at,
			provider_user_id = COALESCE(EXCLUDED.provider_user_id, provider_tokens.provider_user_id),
			updated_at = NOW()`

	// Select provider tokens
	querySelectProviderTokens = `
		SELECT 
			user_id, provider, access_token, refresh_token, token_type,
			scopes, expires_at, provider_user_id, created_at, updated_at, last_used_at
		FROM provider_tokens
		WHERE user_id = $1 AND provider = $2`

	// Update last used at
	queryUpdateProviderTokensLastUsed = `
		UPDATE provider_tokens SET last_used_at = NOW() WHERE user_id = $1 AND provider = $2`

	// Delete provider tokens
	queryDeleteProviderTokens = `
		DELETE FROM provider_tokens WHERE user_id = $1 AND provider = $2`

	// Delete all provider tokens for user
	queryDeleteProviderTokensByUser = `
		DELETE FROM provider_tokens WHERE user_id = $1`

	// Check provider tokens exist
	queryProviderTokensExist = `
		SELECT EXISTS(SELECT 1 FROM provider_tokens WHERE user_id = $1 AND provider = $2)`
)

// ============================================================================
// Verification Events Queries
// ============================================================================

const (
	// Insert verification event
	queryInsertVerificationEvent = `
		INSERT INTO verification_events (
			id, aggregate_id, aggregate_type, event_type, event_data, version, metadata, occurred_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	// Select events by aggregate ID
	querySelectEventsByAggregateID = `
		SELECT 
			id, aggregate_id, aggregate_type, event_type, event_data, version, metadata, occurred_at
		FROM verification_events
		WHERE aggregate_id = $1
		ORDER BY version ASC`

	// Select events by aggregate ID from version
	querySelectEventsByAggregateIDFromVersion = `
		SELECT 
			id, aggregate_id, aggregate_type, event_type, event_data, version, metadata, occurred_at
		FROM verification_events
		WHERE aggregate_id = $1 AND version > $2
		ORDER BY version ASC`
)

// ============================================================================
// Read Model Queries
// ============================================================================

const (
	// Get verification view by ID
	queryGetVerificationViewByID = `
		SELECT 
			id, user_id, provider, credential_type, status,
			provider_user_id, username, credential_id,
			failure_reason, failure_code,
			initiated_at, authorized_at, completed_at, failed_at, expires_at
		FROM verifications
		WHERE id = $1`

	// Get verifications by user with pagination
	queryGetVerificationsByUser = `
		SELECT 
			id, provider, status, initiated_at, completed_at
		FROM verifications
		WHERE user_id = $1
		ORDER BY initiated_at DESC
		LIMIT $2 OFFSET $3`

	// Count verifications by user
	queryCountVerificationsByUser = `
		SELECT COUNT(*) FROM verifications WHERE user_id = $1`

	// Get verifications by user and provider
	queryGetVerificationsByUserAndProvider = `
		SELECT 
			id, user_id, provider, credential_type, status,
			provider_user_id, username, credential_id,
			failure_reason, failure_code,
			initiated_at, authorized_at, completed_at, failed_at, expires_at
		FROM verifications
		WHERE user_id = $1 AND provider = $2
		ORDER BY initiated_at DESC`

	// Get latest verification by user and provider
	queryGetLatestVerificationByUserAndProvider = `
		SELECT 
			id, user_id, provider, credential_type, status,
			provider_user_id, username, credential_id,
			failure_reason, failure_code,
			initiated_at, authorized_at, completed_at, failed_at, expires_at
		FROM verifications
		WHERE user_id = $1 AND provider = $2
		ORDER BY initiated_at DESC
		LIMIT 1`

	// Get provider connections for user
	queryGetProviderConnections = `
		SELECT 
			provider,
			BOOL_OR(status = 'completed') AS connected,
			MAX(CASE WHEN status = 'completed' THEN provider_user_id END) AS provider_user_id,
			MAX(CASE WHEN status = 'completed' THEN username END) AS username,
			MAX(CASE WHEN status = 'completed' THEN credential_id END) AS credential_id,
			MAX(CASE WHEN status = 'completed' THEN completed_at END) AS verified_at
		FROM verifications
		WHERE user_id = $1
		GROUP BY provider`

	// Check if provider is connected
	queryCheckProviderConnected = `
		SELECT 
			EXISTS(
				SELECT 1 FROM verifications 
				WHERE user_id = $1 AND provider = $2 AND status = 'completed'
			) AS connected,
			(
				SELECT credential_id FROM verifications 
				WHERE user_id = $1 AND provider = $2 AND status = 'completed'
				ORDER BY completed_at DESC LIMIT 1
			) AS credential_id,
			(
				SELECT username FROM verifications 
				WHERE user_id = $1 AND provider = $2 AND status = 'completed'
				ORDER BY completed_at DESC LIMIT 1
			) AS username`

	// Count verifications by status
	queryCountVerificationsByStatus = `
		SELECT COUNT(*) FROM verifications WHERE status = $1`

	// List verifications with filters
	queryListVerifications = `
		SELECT 
			id, user_id, provider, credential_type, status,
			provider_user_id, username, credential_id,
			failure_reason, failure_code,
			initiated_at, authorized_at, completed_at, failed_at, expires_at
		FROM verifications
		WHERE 1=1`
)
