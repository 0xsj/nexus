-- ============================================================================
-- User Event Queries
-- ============================================================================

-- name: InsertUserEvent :exec
INSERT INTO identity_user_events (
    id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: GetUserEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM identity_user_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetUserEventsFromVersion :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM identity_user_events
WHERE aggregate_id = $1 AND version > $2
ORDER BY version ASC;

-- name: GetLatestUserEventVersion :one
SELECT COALESCE(MAX(version), 0)::INTEGER AS version
FROM identity_user_events
WHERE aggregate_id = $1;

-- name: UserAggregateExists :one
SELECT EXISTS(
    SELECT 1 FROM identity_user_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Session Event Queries
-- ============================================================================

-- name: InsertSessionEvent :exec
INSERT INTO identity_session_events (
    id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: GetSessionEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM identity_session_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetSessionEventsFromVersion :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM identity_session_events
WHERE aggregate_id = $1 AND version > $2
ORDER BY version ASC;

-- name: GetLatestSessionEventVersion :one
SELECT COALESCE(MAX(version), 0)::INTEGER AS version
FROM identity_session_events
WHERE aggregate_id = $1;

-- name: SessionAggregateExists :one
SELECT EXISTS(
    SELECT 1 FROM identity_session_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- User Projection Queries
-- ============================================================================

-- name: UpsertUser :exec
INSERT INTO identity_users (id, email, display_name, status, primary_did, created_at, updated_at, version)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    status = EXCLUDED.status,
    primary_did = EXCLUDED.primary_did,
    updated_at = EXCLUDED.updated_at,
    version = EXCLUDED.version;

-- name: GetUserByID :one
SELECT id, email, display_name, status, primary_did, created_at, updated_at, version
FROM identity_users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, display_name, status, primary_did, created_at, updated_at, version
FROM identity_users
WHERE email = $1;

-- name: UserExistsByID :one
SELECT EXISTS(SELECT 1 FROM identity_users WHERE id = $1) AS exists;

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM identity_users WHERE email = $1) AS exists;

-- name: DeleteUser :exec
DELETE FROM identity_users WHERE id = $1;

-- ============================================================================
-- User DID Queries
-- ============================================================================

-- name: InsertUserDID :exec
INSERT INTO identity_user_dids (id, user_id, did, is_primary, added_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserDIDs :many
SELECT id, user_id, did, is_primary, added_at
FROM identity_user_dids
WHERE user_id = $1
ORDER BY is_primary DESC, added_at ASC;

-- name: GetUserIDByDID :one
SELECT user_id
FROM identity_user_dids
WHERE did = $1;

-- name: UserExistsByDID :one
SELECT EXISTS(SELECT 1 FROM identity_user_dids WHERE did = $1) AS exists;

-- name: DeleteUserDID :exec
DELETE FROM identity_user_dids WHERE user_id = $1 AND did = $2;

-- name: DeleteAllUserDIDs :exec
DELETE FROM identity_user_dids WHERE user_id = $1;

-- ============================================================================
-- User OAuth Link Queries
-- ============================================================================

-- name: InsertUserOAuthLink :exec
INSERT INTO identity_user_oauth_links (id, user_id, provider, external_id, email, linked_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetUserOAuthLinks :many
SELECT id, user_id, provider, external_id, email, linked_at
FROM identity_user_oauth_links
WHERE user_id = $1
ORDER BY linked_at ASC;

-- name: GetUserIDByOAuthSubject :one
SELECT user_id
FROM identity_user_oauth_links
WHERE provider = $1 AND external_id = $2;

-- name: UserExistsByOAuthSubject :one
SELECT EXISTS(
    SELECT 1 FROM identity_user_oauth_links WHERE provider = $1 AND external_id = $2
) AS exists;

-- name: DeleteUserOAuthLink :exec
DELETE FROM identity_user_oauth_links WHERE user_id = $1 AND provider = $2;

-- name: DeleteAllUserOAuthLinks :exec
DELETE FROM identity_user_oauth_links WHERE user_id = $1;

-- ============================================================================
-- Session Projection Queries
-- ============================================================================

-- name: UpsertSession :exec
INSERT INTO identity_sessions (
    id, user_id, token_hash, auth_method, status, ip_address, user_agent, 
    expires_at, created_at, updated_at, version
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
    token_hash = EXCLUDED.token_hash,
    status = EXCLUDED.status,
    expires_at = EXCLUDED.expires_at,
    updated_at = EXCLUDED.updated_at,
    version = EXCLUDED.version;

-- name: GetSessionByID :one
SELECT id, user_id, token_hash, auth_method, status, ip_address, user_agent,
       expires_at, created_at, updated_at, version
FROM identity_sessions
WHERE id = $1;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, token_hash, auth_method, status, ip_address, user_agent,
       expires_at, created_at, updated_at, version
FROM identity_sessions
WHERE token_hash = $1;

-- name: GetActiveSessionsForUser :many
SELECT id, user_id, token_hash, auth_method, status, ip_address, user_agent,
       expires_at, created_at, updated_at, version
FROM identity_sessions
WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: CountActiveSessionsForUser :one
SELECT COUNT(*)::INTEGER AS count
FROM identity_sessions
WHERE user_id = $1 AND status = 'active' AND expires_at > NOW();

-- name: SessionExistsByID :one
SELECT EXISTS(SELECT 1 FROM identity_sessions WHERE id = $1) AS exists;

-- name: DeleteSession :exec
DELETE FROM identity_sessions WHERE id = $1;

-- name: DeleteUserSessions :exec
DELETE FROM identity_sessions WHERE user_id = $1;

-- ============================================================================
-- Magic Link Queries
-- ============================================================================

-- name: InsertMagicLink :exec
INSERT INTO identity_magic_links (id, token_hash, email, expires_at, used, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetMagicLinkByTokenHash :one
SELECT id, token_hash, email, expires_at, used, used_at, created_at
FROM identity_magic_links
WHERE token_hash = $1;

-- name: MarkMagicLinkUsed :exec
UPDATE identity_magic_links
SET used = TRUE, used_at = NOW()
WHERE token_hash = $1;

-- name: DeleteExpiredMagicLinks :execrows
DELETE FROM identity_magic_links
WHERE expires_at < NOW() OR (used = TRUE AND used_at < NOW() - INTERVAL '1 hour');

-- ============================================================================
-- OAuth State Queries
-- ============================================================================

-- name: InsertOAuthState :exec
INSERT INTO identity_oauth_states (id, state, provider, redirect_url, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetOAuthStateByState :one
SELECT id, state, provider, redirect_url, expires_at, created_at
FROM identity_oauth_states
WHERE state = $1;

-- name: DeleteOAuthState :exec
DELETE FROM identity_oauth_states WHERE state = $1;

-- name: DeleteExpiredOAuthStates :execrows
DELETE FROM identity_oauth_states WHERE expires_at < NOW();