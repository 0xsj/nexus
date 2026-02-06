-- ============================================================================
-- Vouch Event Store Queries
-- ============================================================================

-- name: InsertVouchEvent :exec
INSERT INTO vouch_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetVouchEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM vouch_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestVouchEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM vouch_events
WHERE aggregate_id = $1;

-- name: VouchAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM vouch_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Vouch Projection Queries
-- ============================================================================

-- name: UpsertVouch :exec
INSERT INTO vouches (
    id, voucher_id, vouchee_id, credential_id, claim_key,
    relationship, strength, statement, context, status,
    expires_at, accepted_at, revoked_at, revocation_reason,
    version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    accepted_at = EXCLUDED.accepted_at,
    revoked_at = EXCLUDED.revoked_at,
    revocation_reason = EXCLUDED.revocation_reason,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetVouchByID :one
SELECT id, voucher_id, vouchee_id, credential_id, claim_key,
       relationship, strength, statement, context, status,
       expires_at, accepted_at, revoked_at, revocation_reason,
       version, created_at, updated_at
FROM vouches
WHERE id = $1;

-- name: ListVouchesByVoucheeID :many
SELECT id, voucher_id, vouchee_id, credential_id, claim_key,
       relationship, strength, statement, context, status,
       expires_at, accepted_at, revoked_at, revocation_reason,
       version, created_at, updated_at
FROM vouches
WHERE vouchee_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountVouchesByVoucheeID :one
SELECT COUNT(*)::integer AS count
FROM vouches
WHERE vouchee_id = $1;

-- name: ListVouchesByVoucherID :many
SELECT id, voucher_id, vouchee_id, credential_id, claim_key,
       relationship, strength, statement, context, status,
       expires_at, accepted_at, revoked_at, revocation_reason,
       version, created_at, updated_at
FROM vouches
WHERE voucher_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountVouchesByVoucherID :one
SELECT COUNT(*)::integer AS count
FROM vouches
WHERE voucher_id = $1;

-- ============================================================================
-- Reputation Projection Queries
-- ============================================================================

-- name: GetReputationByUserID :one
SELECT user_id, overall_score, credential_score, vouch_score,
       network_score, vouch_count, last_calculated_at
FROM reputations
WHERE user_id = $1;
