-- ============================================================================
-- Issuer Event Store Queries
-- ============================================================================

-- name: InsertIssuerEvent :exec
INSERT INTO issuer_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetIssuerEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM issuer_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestIssuerEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM issuer_events
WHERE aggregate_id = $1;

-- name: IssuerAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM issuer_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Template Event Store Queries
-- ============================================================================

-- name: InsertTemplateEvent :exec
INSERT INTO template_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetTemplateEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM template_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestTemplateEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM template_events
WHERE aggregate_id = $1;

-- name: TemplateAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM template_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Issuer Projection Queries
-- ============================================================================

-- name: UpsertIssuer :exec
INSERT INTO issuers (
    id, organization_id, name, description, did, webhook_url,
    api_key_hash, status, branding, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    did = EXCLUDED.did,
    webhook_url = EXCLUDED.webhook_url,
    api_key_hash = EXCLUDED.api_key_hash,
    status = EXCLUDED.status,
    branding = EXCLUDED.branding,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetIssuerByID :one
SELECT id, organization_id, name, description, did, webhook_url,
       api_key_hash, status, branding, version, created_at, updated_at
FROM issuers
WHERE id = $1;

-- name: GetIssuerByOrganizationID :one
SELECT id, organization_id, name, description, did, webhook_url,
       api_key_hash, status, branding, version, created_at, updated_at
FROM issuers
WHERE organization_id = $1;

-- name: ListIssuers :many
SELECT id, organization_id, name, description, did, webhook_url,
       api_key_hash, status, branding, version, created_at, updated_at
FROM issuers
WHERE (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status')::varchar)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountIssuers :one
SELECT COUNT(*)::integer AS count
FROM issuers
WHERE (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status')::varchar);

-- ============================================================================
-- Template Projection Queries
-- ============================================================================

-- name: UpsertTemplate :exec
INSERT INTO templates (
    id, issuer_id, name, description, schema_type, claim_mappings,
    default_values, expiration_days, auto_approve, status, version,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    claim_mappings = EXCLUDED.claim_mappings,
    default_values = EXCLUDED.default_values,
    expiration_days = EXCLUDED.expiration_days,
    auto_approve = EXCLUDED.auto_approve,
    status = EXCLUDED.status,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetTemplateByID :one
SELECT id, issuer_id, name, description, schema_type, claim_mappings,
       default_values, expiration_days, auto_approve, status, version,
       created_at, updated_at
FROM templates
WHERE id = $1;

-- name: ListTemplatesByIssuerID :many
SELECT id, issuer_id, name, description, schema_type, claim_mappings,
       default_values, expiration_days, auto_approve, status, version,
       created_at, updated_at
FROM templates
WHERE issuer_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTemplatesByIssuerID :one
SELECT COUNT(*)::integer AS count
FROM templates
WHERE issuer_id = $1;
