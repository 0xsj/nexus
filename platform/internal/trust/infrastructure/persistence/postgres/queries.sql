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

-- ============================================================================
-- Identity User Projection Queries
-- ============================================================================

-- name: UpsertUserProjection :exec
INSERT INTO trust_user_projections (user_id, active, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (user_id) DO UPDATE SET
    active = EXCLUDED.active,
    updated_at = EXCLUDED.updated_at;

-- name: UserProjectionExists :one
SELECT EXISTS (
    SELECT 1 FROM trust_user_projections WHERE user_id = $1 AND active = true
) AS exists;

-- name: UpdateUserProjectionActive :exec
UPDATE trust_user_projections SET active = $2, updated_at = NOW() WHERE user_id = $1;

-- ============================================================================
-- Credential Projection Queries
-- ============================================================================

-- name: UpsertCredentialProjection :exec
INSERT INTO trust_credential_projections (credential_id, credential_type, subject_did, user_id, status, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (credential_id) DO UPDATE SET
    credential_type = EXCLUDED.credential_type,
    subject_did = EXCLUDED.subject_did,
    user_id = EXCLUDED.user_id,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at;

-- name: CredentialProjectionExists :one
SELECT EXISTS (
    SELECT 1 FROM trust_credential_projections WHERE credential_id = $1 AND status = 'active'
) AS exists;

-- name: CountCredentialProjectionsByUserID :one
SELECT COUNT(*)::integer AS count
FROM trust_credential_projections
WHERE user_id = $1 AND status = 'active';

-- name: UpdateCredentialProjectionStatus :exec
UPDATE trust_credential_projections SET status = $2, updated_at = NOW() WHERE credential_id = $1;

-- name: GetUserIDByDID :one
SELECT user_id
FROM trust_user_projections
WHERE user_id = (SELECT user_id FROM trust_credential_projections WHERE subject_did = $1 LIMIT 1);

-- ============================================================================
-- Organization Projection Queries
-- ============================================================================

-- name: UpsertOrganizationProjection :exec
INSERT INTO trust_organization_projections (organization_id, verification_status, active, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (organization_id) DO UPDATE SET
    verification_status = EXCLUDED.verification_status,
    active = EXCLUDED.active,
    updated_at = EXCLUDED.updated_at;

-- name: GetOrganizationProjection :one
SELECT organization_id, verification_status, active, updated_at
FROM trust_organization_projections
WHERE organization_id = $1;

-- name: UpdateOrganizationProjectionVerified :exec
UPDATE trust_organization_projections SET verification_status = $2, updated_at = NOW() WHERE organization_id = $1;

-- name: UpdateOrganizationProjectionActive :exec
UPDATE trust_organization_projections SET active = $2, updated_at = NOW() WHERE organization_id = $1;
