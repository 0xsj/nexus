-- ============================================================================
-- Wallet Event Store Queries
-- ============================================================================

-- name: InsertWalletEvent :exec
INSERT INTO wallet_events (id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetWalletEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, event_data, version, occurred_at
FROM wallet_events
WHERE aggregate_id = $1
ORDER BY version ASC;

-- name: GetLatestWalletEventVersion :one
SELECT COALESCE(MAX(version), 0)::integer AS version
FROM wallet_events
WHERE aggregate_id = $1;

-- name: WalletAggregateExists :one
SELECT EXISTS (
    SELECT 1 FROM wallet_events WHERE aggregate_id = $1
) AS exists;

-- ============================================================================
-- Wallet Projection Queries
-- ============================================================================

-- name: UpsertWallet :exec
INSERT INTO wallets (
    id, user_id, address, chain, label, status,
    is_primary, did, linked_at, verified_at,
    version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (id) DO UPDATE SET
    label = EXCLUDED.label,
    status = EXCLUDED.status,
    is_primary = EXCLUDED.is_primary,
    did = EXCLUDED.did,
    verified_at = EXCLUDED.verified_at,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at;

-- name: GetWalletByID :one
SELECT id, user_id, address, chain, label, status,
       is_primary, did, linked_at, verified_at,
       version, created_at, updated_at
FROM wallets
WHERE id = $1;

-- name: WalletExistsByID :one
SELECT EXISTS (
    SELECT 1 FROM wallets WHERE id = $1
) AS exists;

-- name: GetWalletByAddress :one
SELECT id, user_id, address, chain, label, status,
       is_primary, did, linked_at, verified_at,
       version, created_at, updated_at
FROM wallets
WHERE address = $1;

-- name: WalletExistsByAddress :one
SELECT EXISTS (
    SELECT 1 FROM wallets WHERE address = $1
) AS exists;

-- name: ListWalletsByUserID :many
SELECT id, user_id, address, chain, label, status,
       is_primary, did, linked_at, verified_at,
       version, created_at, updated_at
FROM wallets
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountWalletsByUserID :one
SELECT COUNT(*)::integer AS count
FROM wallets
WHERE user_id = $1;

-- name: GetPrimaryWalletByUserID :one
SELECT id, user_id, address, chain, label, status,
       is_primary, did, linked_at, verified_at,
       version, created_at, updated_at
FROM wallets
WHERE user_id = $1 AND is_primary = TRUE;
