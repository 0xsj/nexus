-- ============================================================================
-- Integration Event Store Queries
-- ============================================================================

-- name: InsertIntegrationEvent :exec
INSERT INTO integration_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetIntegrationEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM integration_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestIntegrationEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM integration_events
WHERE aggregate_id = $1;

-- name: IntegrationAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM integration_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Integration Projection Queries
-- ============================================================================

-- name: UpsertIntegration :exec
INSERT INTO integrations (
    id, user_id, provider_type, status, provider_user_id, provider_username,
    scopes, last_fetch_at, fetch_count, metadata, connected_at, disconnected_at,
    suspended_at, suspension_reason, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    scopes = EXCLUDED.scopes,
    last_fetch_at = EXCLUDED.last_fetch_at,
    fetch_count = EXCLUDED.fetch_count,
    metadata = EXCLUDED.metadata,
    disconnected_at = EXCLUDED.disconnected_at,
    suspended_at = EXCLUDED.suspended_at,
    suspension_reason = EXCLUDED.suspension_reason,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetIntegrationByID :one
SELECT id, user_id, provider_type, status, provider_user_id, provider_username,
       scopes, last_fetch_at, fetch_count, metadata, connected_at, disconnected_at,
       suspended_at, suspension_reason, version, created_at, updated_at
FROM integrations
WHERE id = $1;

-- name: ListIntegrationsByUserID :many
SELECT id, user_id, provider_type, status, provider_user_id, provider_username,
       scopes, last_fetch_at, fetch_count, metadata, connected_at, disconnected_at,
       suspended_at, suspension_reason, version, created_at, updated_at
FROM integrations
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CountIntegrationsByUserID :one
SELECT COUNT(*)::integer AS count
FROM integrations
WHERE user_id = $1;

-- name: GetIntegrationByUserAndProvider :one
SELECT id, user_id, provider_type, status, provider_user_id, provider_username,
       scopes, last_fetch_at, fetch_count, metadata, connected_at, disconnected_at,
       suspended_at, suspension_reason, version, created_at, updated_at
FROM integrations
WHERE user_id = $1 AND provider_type = $2;

-- ============================================================================
-- OAuth Token Storage Queries
-- ============================================================================

-- name: StoreTokens :exec
INSERT INTO integration_tokens (
    integration_id, access_token_encrypted, refresh_token_encrypted,
    token_type, expires_at, scopes, encryption_version
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (integration_id) DO UPDATE SET
    access_token_encrypted = EXCLUDED.access_token_encrypted,
    refresh_token_encrypted = EXCLUDED.refresh_token_encrypted,
    token_type = EXCLUDED.token_type,
    expires_at = EXCLUDED.expires_at,
    scopes = EXCLUDED.scopes,
    encryption_version = EXCLUDED.encryption_version,
    updated_at = NOW();

-- name: GetTokens :one
SELECT integration_id, access_token_encrypted, refresh_token_encrypted,
       token_type, expires_at, scopes, encryption_version, created_at, updated_at
FROM integration_tokens
WHERE integration_id = $1;

-- name: DeleteTokens :exec
DELETE FROM integration_tokens
WHERE integration_id = $1;

-- name: TokensExist :one
SELECT EXISTS (
    SELECT 1 FROM integration_tokens WHERE integration_id = $1
) AS exists;

-- name: ListExpiredTokens :many
SELECT integration_id, access_token_encrypted, refresh_token_encrypted,
       token_type, expires_at, scopes, encryption_version, created_at, updated_at
FROM integration_tokens
WHERE expires_at IS NOT NULL AND expires_at < NOW()
ORDER BY expires_at ASC
LIMIT $1;

-- name: CountTokensByEncryptionVersion :one
SELECT COUNT(*)::integer AS count
FROM integration_tokens
WHERE encryption_version = $1;
