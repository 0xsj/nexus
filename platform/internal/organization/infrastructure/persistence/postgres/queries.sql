-- ============================================================================
-- Organization Event Store Queries
-- ============================================================================

-- name: InsertOrganizationEvent :exec
INSERT INTO organization_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetOrganizationEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM organization_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestOrganizationEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM organization_events
WHERE aggregate_id = $1;

-- name: OrganizationAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Organization Projection Queries
-- ============================================================================

-- name: UpsertOrganization :exec
INSERT INTO organizations (
    id, name, slug, org_type, description,
    verification_status, did, owner_member_id, member_count,
    deleted, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    verification_status = EXCLUDED.verification_status,
    did = EXCLUDED.did,
    owner_member_id = EXCLUDED.owner_member_id,
    member_count = EXCLUDED.member_count,
    deleted = EXCLUDED.deleted,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetOrganizationByID :one
SELECT id, name, slug, org_type, description,
       verification_status, did, owner_member_id, member_count,
       deleted, version, created_at, updated_at
FROM organizations
WHERE id = $1;

-- name: OrganizationExistsByID :one
SELECT EXISTS (
    SELECT 1 FROM organizations WHERE id = $1
) AS exists;

-- name: GetOrganizationBySlug :one
SELECT id, name, slug, org_type, description,
       verification_status, did, owner_member_id, member_count,
       deleted, version, created_at, updated_at
FROM organizations
WHERE slug = $1;

-- name: OrganizationExistsBySlug :one
SELECT EXISTS (
    SELECT 1 FROM organizations WHERE slug = $1
) AS exists;

-- name: ListOrganizations :many
SELECT id, name, slug, org_type, description,
       verification_status, did, owner_member_id, member_count,
       deleted, version, created_at, updated_at
FROM organizations
WHERE NOT deleted
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountOrganizations :one
SELECT COUNT(*)::integer AS count
FROM organizations
WHERE NOT deleted;
