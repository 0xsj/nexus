-- ============================================================================
-- Profile Event Store Queries
-- ============================================================================

-- name: InsertProfileEvent :exec
INSERT INTO profile_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetProfileEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM profile_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestProfileEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM profile_events
WHERE aggregate_id = $1;

-- name: ProfileAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM profile_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Profile Projection Queries
-- ============================================================================

-- name: UpsertProfile :exec
INSERT INTO profiles (
    id, user_id, display_name, headline, bio,
    vanity_slug, badge_count, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    headline = EXCLUDED.headline,
    bio = EXCLUDED.bio,
    vanity_slug = EXCLUDED.vanity_slug,
    badge_count = EXCLUDED.badge_count,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetProfileByID :one
SELECT id, user_id, display_name, headline, bio,
       vanity_slug, badge_count, version, created_at, updated_at
FROM profiles
WHERE id = $1;

-- name: ProfileExistsByID :one
SELECT EXISTS (
    SELECT 1 FROM profiles WHERE id = $1
) AS exists;

-- name: GetProfileByUserID :one
SELECT id, user_id, display_name, headline, bio,
       vanity_slug, badge_count, version, created_at, updated_at
FROM profiles
WHERE user_id = $1;

-- name: GetProfileByVanitySlug :one
SELECT id, user_id, display_name, headline, bio,
       vanity_slug, badge_count, version, created_at, updated_at
FROM profiles
WHERE vanity_slug = $1;

-- name: ProfileExistsByVanitySlug :one
SELECT EXISTS (
    SELECT 1 FROM profiles WHERE vanity_slug = $1
) AS exists;

-- name: ListProfiles :many
SELECT id, user_id, display_name, headline, bio,
       vanity_slug, badge_count, version, created_at, updated_at
FROM profiles
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountProfiles :one
SELECT COUNT(*)::integer AS count
FROM profiles;

-- ============================================================================
-- Credential Projection Queries
-- ============================================================================

-- name: UpsertCredentialProjection :exec
INSERT INTO profile_credential_projections (credential_id, credential_type, status, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (credential_id) DO UPDATE SET
    credential_type = EXCLUDED.credential_type,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at;

-- name: GetCredentialProjection :one
SELECT credential_id, credential_type, status, updated_at
FROM profile_credential_projections
WHERE credential_id = $1;

-- name: UpdateCredentialProjectionStatus :exec
UPDATE profile_credential_projections SET status = $2, updated_at = NOW() WHERE credential_id = $1;
