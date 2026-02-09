-- ============================================================================
-- Credential Event Store Queries
-- ============================================================================

-- name: InsertCredentialEvent :exec
INSERT INTO credential_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetCredentialEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM credential_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestCredentialEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM credential_events
WHERE aggregate_id = $1;

-- name: CredentialAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM credential_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Credential Projection Queries
-- ============================================================================

-- name: UpsertCredential :exec
INSERT INTO credentials (
    id, credential_type, issuer_did, subject_did, claims,
    issued_at, expires_at, status, revoked_at, revocation_reason,
    jwt, verification_id, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    revoked_at = EXCLUDED.revoked_at,
    revocation_reason = EXCLUDED.revocation_reason,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetCredentialByID :one
SELECT id, credential_type, issuer_did, subject_did, claims,
       issued_at, expires_at, status, revoked_at, revocation_reason,
       jwt, verification_id, version, created_at, updated_at
FROM credentials
WHERE id = $1;

-- name: CredentialExistsByID :one
SELECT EXISTS (
    SELECT 1 FROM credentials WHERE id = $1
) AS exists;

-- name: ListCredentialsBySubjectDID :many
SELECT id, credential_type, issuer_did, subject_did, claims,
       issued_at, expires_at, status, revoked_at, revocation_reason,
       jwt, verification_id, version, created_at, updated_at
FROM credentials
WHERE subject_did = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCredentialsBySubjectDID :one
SELECT COUNT(*)::integer AS count
FROM credentials
WHERE subject_did = $1;

-- name: ListCredentialsByIssuerDID :many
SELECT id, credential_type, issuer_did, subject_did, claims,
       issued_at, expires_at, status, revoked_at, revocation_reason,
       jwt, verification_id, version, created_at, updated_at
FROM credentials
WHERE issuer_did = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCredentialsByIssuerDID :one
SELECT COUNT(*)::integer AS count
FROM credentials
WHERE issuer_did = $1;

-- ============================================================================
-- Schema Projection Queries
-- ============================================================================

-- name: UpsertSchemaProjection :exec
INSERT INTO credential_schema_projections (schema_id, schema_type, status, claims, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (schema_id) DO UPDATE SET
    schema_type = EXCLUDED.schema_type,
    status = EXCLUDED.status,
    claims = EXCLUDED.claims,
    updated_at = EXCLUDED.updated_at;

-- name: GetSchemaProjectionByType :one
SELECT schema_id, schema_type, status, claims, updated_at
FROM credential_schema_projections
WHERE schema_type = $1;

-- name: SchemaProjectionExistsByType :one
SELECT EXISTS (
    SELECT 1 FROM credential_schema_projections WHERE schema_type = $1 AND status = 'active'
) AS exists;

-- name: UpdateSchemaProjectionStatus :exec
UPDATE credential_schema_projections SET status = $2, updated_at = NOW() WHERE schema_id = $1;

-- name: UpdateSchemaProjectionClaims :exec
UPDATE credential_schema_projections SET claims = $2, updated_at = NOW() WHERE schema_id = $1;
