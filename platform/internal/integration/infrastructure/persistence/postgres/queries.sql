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
