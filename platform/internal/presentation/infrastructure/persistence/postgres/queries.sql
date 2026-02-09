-- ============================================================================
-- Presentation Event Store Queries
-- ============================================================================

-- name: InsertPresentationEvent :exec
INSERT INTO presentation_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetPresentationEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM presentation_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestPresentationEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM presentation_events
WHERE aggregate_id = $1;

-- name: PresentationAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM presentation_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- ShareLink Event Store Queries
-- ============================================================================

-- name: InsertShareLinkEvent :exec
INSERT INTO share_link_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetShareLinkEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM share_link_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestShareLinkEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM share_link_events
WHERE aggregate_id = $1;

-- name: ShareLinkAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM share_link_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Presentation Projection Queries
-- ============================================================================

-- name: UpsertPresentation :exec
INSERT INTO presentations (
    id, holder_did, credential_ids, disclosure_policy, vp_jwt, purpose,
    status, revoked_at, revocation_reason, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    revoked_at = EXCLUDED.revoked_at,
    revocation_reason = EXCLUDED.revocation_reason,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetPresentationByID :one
SELECT id, holder_did, credential_ids, disclosure_policy, vp_jwt, purpose,
       status, revoked_at, revocation_reason, version, created_at, updated_at
FROM presentations
WHERE id = $1;

-- name: ListPresentationsByHolderDID :many
SELECT id, holder_did, credential_ids, disclosure_policy, vp_jwt, purpose,
       status, revoked_at, revocation_reason, version, created_at, updated_at
FROM presentations
WHERE holder_did = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPresentationsByHolderDID :one
SELECT COUNT(*)::integer AS count
FROM presentations
WHERE holder_did = $1;

-- ============================================================================
-- ShareLink Projection Queries
-- ============================================================================

-- name: UpsertShareLink :exec
INSERT INTO share_links (
    id, presentation_id, token, expires_at, max_views, current_views,
    pin_hash, audience, status, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (id) DO UPDATE SET
    current_views = EXCLUDED.current_views,
    status = EXCLUDED.status,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetShareLinkByID :one
SELECT id, presentation_id, token, expires_at, max_views, current_views,
       pin_hash, audience, status, version, created_at, updated_at
FROM share_links
WHERE id = $1;

-- name: GetShareLinkByToken :one
SELECT id, presentation_id, token, expires_at, max_views, current_views,
       pin_hash, audience, status, version, created_at, updated_at
FROM share_links
WHERE token = $1;

-- name: ListShareLinksByPresentationID :many
SELECT id, presentation_id, token, expires_at, max_views, current_views,
       pin_hash, audience, status, version, created_at, updated_at
FROM share_links
WHERE presentation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountShareLinksByPresentationID :one
SELECT COUNT(*)::integer AS count
FROM share_links
WHERE presentation_id = $1;

-- ============================================================================
-- Access Grant Queries
-- ============================================================================

-- name: InsertAccessGrant :exec
INSERT INTO access_grants (id, share_link_id, verifier_did, accessed_at, ip_address, disclosed_claims)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListAccessGrantsByShareLinkID :many
SELECT id, share_link_id, verifier_did, accessed_at, ip_address, disclosed_claims
FROM access_grants
WHERE share_link_id = $1
ORDER BY accessed_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAccessGrantsByShareLinkID :one
SELECT COUNT(*)::integer AS count
FROM access_grants
WHERE share_link_id = $1;

-- ============================================================================
-- Identity User Projection Queries
-- ============================================================================

-- name: UpsertUserProjection :exec
INSERT INTO presentation_user_projections (user_id, primary_did, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (user_id) DO UPDATE SET
    primary_did = EXCLUDED.primary_did,
    updated_at = EXCLUDED.updated_at;

-- name: GetUserProjectionDID :one
SELECT primary_did
FROM presentation_user_projections
WHERE user_id = $1;

-- name: UpdateUserProjectionDID :exec
UPDATE presentation_user_projections SET primary_did = $2, updated_at = NOW() WHERE user_id = $1;

-- ============================================================================
-- Credential Projection Queries
-- ============================================================================

-- name: UpsertCredentialProjection :exec
INSERT INTO presentation_credential_projections (credential_id, credential_type, subject_did, user_id, status, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (credential_id) DO UPDATE SET
    credential_type = EXCLUDED.credential_type,
    subject_did = EXCLUDED.subject_did,
    user_id = EXCLUDED.user_id,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at;

-- name: GetCredentialProjection :one
SELECT credential_id, credential_type, subject_did, user_id, status, updated_at
FROM presentation_credential_projections
WHERE credential_id = $1;

-- name: CredentialProjectionExists :one
SELECT EXISTS (
    SELECT 1 FROM presentation_credential_projections WHERE credential_id = $1 AND status = 'active'
) AS exists;

-- name: ListCredentialProjectionsByUserID :many
SELECT credential_id
FROM presentation_credential_projections
WHERE user_id = $1 AND status = 'active';

-- name: UpdateCredentialProjectionStatus :exec
UPDATE presentation_credential_projections SET status = $2, updated_at = NOW() WHERE credential_id = $1;

-- name: GetUserIDByDID :one
SELECT user_id
FROM presentation_user_projections
WHERE primary_did = $1;
