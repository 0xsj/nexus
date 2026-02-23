-- ============================================================================
-- Schema Queries
-- ============================================================================
-- sqlc query definitions for the Schema context.
-- ============================================================================

-- ============================================================================
-- Schema CRUD
-- ============================================================================

-- name: InsertSchema :exec
INSERT INTO schemas (
    id,
    schema_type,
    name,
    description,
    current_version,
    status,
    issuer_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: UpdateSchema :exec
UPDATE schemas
SET
    name = $2,
    description = $3,
    current_version = $4,
    status = $5,
    updated_at = $6
WHERE id = $1;

-- name: UpdateSchemaMetadata :exec
UPDATE schemas
SET
    name = COALESCE(NULLIF($2, ''), name),
    description = COALESCE(NULLIF($3, ''), description),
    updated_at = $4
WHERE id = $1;

-- name: UpdateSchemaStatus :exec
UPDATE schemas
SET
    status = $2,
    updated_at = $3
WHERE id = $1;

-- name: UpdateSchemaVersion :exec
UPDATE schemas
SET
    current_version = $2,
    updated_at = $3
WHERE id = $1;

-- name: GetSchemaByID :one
SELECT *
FROM schemas
WHERE id = $1;

-- name: GetSchemaByType :one
SELECT *
FROM schemas
WHERE schema_type = $1;

-- name: SchemaExistsByID :one
SELECT EXISTS(
    SELECT 1 FROM schemas WHERE id = $1
) AS exists;

-- name: SchemaExistsByType :one
SELECT EXISTS(
    SELECT 1 FROM schemas WHERE schema_type = $1
) AS exists;

-- name: GetSchemaIDByType :one
SELECT id
FROM schemas
WHERE schema_type = $1;

-- name: ListSchemas :many
SELECT *
FROM schemas
WHERE
    (@status::VARCHAR = '' OR status = @status)
    AND (@issuer_id::UUID IS NULL OR issuer_id = @issuer_id)
    AND (
        @built_in_only::BOOLEAN IS NULL
        OR (@built_in_only = true AND issuer_id IS NULL)
        OR (@built_in_only = false AND issuer_id IS NOT NULL)
    )
    AND (
        @search_query::VARCHAR = ''
        OR to_tsvector('english', name || ' ' || description) @@ plainto_tsquery('english', @search_query)
    )
ORDER BY
    CASE WHEN @order_by::VARCHAR = 'name' AND @order_dir::VARCHAR = 'asc' THEN name END ASC,
    CASE WHEN @order_by::VARCHAR = 'name' AND @order_dir::VARCHAR = 'desc' THEN name END DESC,
    CASE WHEN @order_by::VARCHAR = 'schema_type' AND @order_dir::VARCHAR = 'asc' THEN schema_type END ASC,
    CASE WHEN @order_by::VARCHAR = 'schema_type' AND @order_dir::VARCHAR = 'desc' THEN schema_type END DESC,
    CASE WHEN @order_by::VARCHAR = 'updated_at' AND @order_dir::VARCHAR = 'asc' THEN updated_at END ASC,
    CASE WHEN @order_by::VARCHAR = 'updated_at' AND @order_dir::VARCHAR = 'desc' THEN updated_at END DESC,
    CASE WHEN @order_by::VARCHAR = '' OR @order_by = 'created_at' THEN created_at END DESC
LIMIT @page_limit
OFFSET @page_offset;

-- name: CountSchemas :one
SELECT COUNT(*)
FROM schemas
WHERE
    (@status::VARCHAR = '' OR status = @status)
    AND (@issuer_id::UUID IS NULL OR issuer_id = @issuer_id)
    AND (
        @built_in_only::BOOLEAN IS NULL
        OR (@built_in_only = true AND issuer_id IS NULL)
        OR (@built_in_only = false AND issuer_id IS NOT NULL)
    )
    AND (
        @search_query::VARCHAR = ''
        OR to_tsvector('english', name || ' ' || description) @@ plainto_tsquery('english', @search_query)
    );

-- ============================================================================
-- Schema Claims CRUD
-- ============================================================================

-- name: InsertSchemaClaim :exec
INSERT INTO schema_claims (
    schema_id,
    key,
    data_type,
    required,
    display_name,
    description,
    disclosable,
    sort_order,
    constraints,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
);

-- name: DeleteSchemaClaimsBySchemaID :exec
DELETE FROM schema_claims
WHERE schema_id = $1;

-- name: GetSchemaClaimsBySchemaID :many
SELECT *
FROM schema_claims
WHERE schema_id = $1
ORDER BY sort_order ASC, key ASC;

-- name: GetSchemaClaimByKey :one
SELECT *
FROM schema_claims
WHERE schema_id = $1 AND key = $2;

-- name: CountSchemaClaimsBySchemaID :one
SELECT COUNT(*)
FROM schema_claims
WHERE schema_id = $1;

-- ============================================================================
-- Schema Versions CRUD
-- ============================================================================

-- name: InsertSchemaVersion :exec
INSERT INTO schema_versions (
    schema_id,
    version,
    change_summary,
    claims_snapshot,
    created_at
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetSchemaVersionsBySchemaID :many
SELECT *
FROM schema_versions
WHERE schema_id = $1
ORDER BY created_at DESC;

-- name: GetSchemaVersion :one
SELECT *
FROM schema_versions
WHERE schema_id = $1 AND version = $2;

-- name: SchemaVersionExists :one
SELECT EXISTS(
    SELECT 1 FROM schema_versions WHERE schema_id = $1 AND version = $2
) AS exists;

-- name: CountSchemaVersionsBySchemaID :one
SELECT COUNT(*)
FROM schema_versions
WHERE schema_id = $1;

-- ============================================================================
-- Issuer Projection Queries
-- ============================================================================

-- name: UpsertIssuerProjection :exec
INSERT INTO schema_issuer_projections (issuer_id, name, active, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (issuer_id) DO UPDATE
SET name = EXCLUDED.name, active = EXCLUDED.active, updated_at = NOW();

-- name: UpdateIssuerProjectionActive :exec
UPDATE schema_issuer_projections
SET active = $2, updated_at = NOW()
WHERE issuer_id = $1;

-- name: GetIssuerProjection :one
SELECT * FROM schema_issuer_projections
WHERE issuer_id = $1;

-- name: IssuerProjectionExists :one
SELECT EXISTS(
    SELECT 1 FROM schema_issuer_projections WHERE issuer_id = $1 AND active = true
) AS exists;