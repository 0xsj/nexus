-- ============================================================================
-- Verification Event Store Queries
-- ============================================================================

-- name: InsertVerificationEvent :exec
INSERT INTO verification_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetVerificationEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM verification_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestVerificationEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM verification_events
WHERE aggregate_id = $1;

-- name: VerificationAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM verification_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Verification Projection Queries
-- ============================================================================

-- name: UpsertVerification :exec
INSERT INTO verifications (
    id, user_id, provider_type, status, oauth_state,
    error_message, error_code, credential_id, version,
    started_at, completed_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    error_message = EXCLUDED.error_message,
    error_code = EXCLUDED.error_code,
    credential_id = EXCLUDED.credential_id,
    version = EXCLUDED.version,
    completed_at = EXCLUDED.completed_at,
    updated_at = EXCLUDED.updated_at;

-- name: GetVerificationByID :one
SELECT id, user_id, provider_type, status, oauth_state,
       error_message, error_code, credential_id, version,
       started_at, completed_at, created_at, updated_at
FROM verifications
WHERE id = $1;

-- name: VerificationExistsByID :one
SELECT EXISTS (
    SELECT 1 FROM verifications WHERE id = $1
) AS exists;

-- name: ListVerificationsByUserID :many
SELECT id, user_id, provider_type, status, oauth_state,
       error_message, error_code, credential_id, version,
       started_at, completed_at, created_at, updated_at
FROM verifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountVerificationsByUserID :one
SELECT COUNT(*)::integer AS count
FROM verifications
WHERE user_id = $1;

-- name: GetVerificationByUserAndProvider :one
SELECT id, user_id, provider_type, status, oauth_state,
       error_message, error_code, credential_id, version,
       started_at, completed_at, created_at, updated_at
FROM verifications
WHERE user_id = $1 AND provider_type = $2
ORDER BY created_at DESC
LIMIT 1;
