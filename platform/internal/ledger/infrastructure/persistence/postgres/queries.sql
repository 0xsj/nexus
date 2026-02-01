-- ============================================================================
-- Ledger Queries
-- ============================================================================
-- sqlc query definitions for the Ledger context.
-- ============================================================================

-- name: InsertEntry :exec
INSERT INTO ledger_entries (
    id,
    occurred_at,
    recorded_at,
    event_type,
    actor_id,
    actor_type,
    subject_id,
    subject_type,
    metadata,
    context_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: InsertEntryBatch :copyfrom
INSERT INTO ledger_entries (
    id,
    occurred_at,
    recorded_at,
    event_type,
    actor_id,
    actor_type,
    subject_id,
    subject_type,
    metadata,
    context_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: GetEntryByID :one
SELECT *
FROM ledger_entries
WHERE id = $1;

-- name: ListEntries :many
SELECT *
FROM ledger_entries
WHERE
    (CARDINALITY(@event_types::VARCHAR[]) = 0 OR event_type = ANY(@event_types::VARCHAR[]))
    AND (@actor_id::VARCHAR = '' OR actor_id = @actor_id)
    AND (@actor_type::VARCHAR = '' OR actor_type = @actor_type)
    AND (@subject_id::VARCHAR = '' OR subject_id = @subject_id)
    AND (@subject_type::VARCHAR = '' OR subject_type = @subject_type)
    AND (@context_id::VARCHAR = '' OR context_id = @context_id)
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at <= @to_time)
ORDER BY occurred_at DESC
LIMIT @page_size
OFFSET @page_offset;

-- name: CountEntries :one
SELECT COUNT(*)
FROM ledger_entries
WHERE
    (CARDINALITY(@event_types::VARCHAR[]) = 0 OR event_type = ANY(@event_types::VARCHAR[]))
    AND (@actor_id::VARCHAR = '' OR actor_id = @actor_id)
    AND (@actor_type::VARCHAR = '' OR actor_type = @actor_type)
    AND (@subject_id::VARCHAR = '' OR subject_id = @subject_id)
    AND (@subject_type::VARCHAR = '' OR subject_type = @subject_type)
    AND (@context_id::VARCHAR = '' OR context_id = @context_id)
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at <= @to_time);

-- name: GetEntriesBySubject :many
SELECT *
FROM ledger_entries
WHERE
    subject_id = $1
    AND subject_type = $2
    AND (CARDINALITY(@event_types::VARCHAR[]) = 0 OR event_type = ANY(@event_types::VARCHAR[]))
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at <= @to_time)
ORDER BY occurred_at DESC
LIMIT @page_size
OFFSET @page_offset;

-- name: CountEntriesBySubject :one
SELECT COUNT(*)
FROM ledger_entries
WHERE
    subject_id = $1
    AND subject_type = $2
    AND (CARDINALITY(@event_types::VARCHAR[]) = 0 OR event_type = ANY(@event_types::VARCHAR[]))
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at <= @to_time);

-- name: GetEntriesByActor :many
SELECT *
FROM ledger_entries
WHERE
    actor_id = $1
    AND actor_type = $2
    AND (CARDINALITY(@event_types::VARCHAR[]) = 0 OR event_type = ANY(@event_types::VARCHAR[]))
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at <= @to_time)
ORDER BY occurred_at DESC
LIMIT @page_size
OFFSET @page_offset;

-- name: CountEntriesByActor :one
SELECT COUNT(*)
FROM ledger_entries
WHERE
    actor_id = $1
    AND actor_type = $2
    AND (CARDINALITY(@event_types::VARCHAR[]) = 0 OR event_type = ANY(@event_types::VARCHAR[]))
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at <= @to_time);

-- name: GetVerificationLogs :many
SELECT outer_e.*
FROM ledger_entries outer_e
WHERE
    outer_e.actor_type = 'external_verifier'
    AND outer_e.subject_type = 'credential'
    AND outer_e.subject_id IN (
        SELECT inner_e.subject_id
        FROM ledger_entries inner_e
        WHERE inner_e.subject_type = 'credential'
        AND inner_e.actor_id = @user_id
        AND inner_e.actor_type = 'user'
    )
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR outer_e.occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR outer_e.occurred_at <= @to_time)
ORDER BY outer_e.occurred_at DESC
LIMIT @page_size
OFFSET @page_offset;

-- name: CountVerificationLogs :one
SELECT COUNT(*)
FROM ledger_entries outer_e
WHERE
    outer_e.actor_type = 'external_verifier'
    AND outer_e.subject_type = 'credential'
    AND outer_e.subject_id IN (
        SELECT inner_e.subject_id
        FROM ledger_entries inner_e
        WHERE inner_e.subject_type = 'credential'
        AND inner_e.actor_id = @user_id
        AND inner_e.actor_type = 'user'
    )
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR outer_e.occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR outer_e.occurred_at <= @to_time);

-- name: GetEventTypeStats :many
SELECT
    event_type,
    COUNT(*) as count
FROM ledger_entries
WHERE
    (@actor_id::VARCHAR = '' OR actor_id = @actor_id)
    AND (@subject_id::VARCHAR = '' OR subject_id = @subject_id)
    AND (@subject_type::VARCHAR = '' OR subject_type = @subject_type)
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at <= @to_time)
GROUP BY event_type
ORDER BY count DESC;

-- name: GetTotalCount :one
SELECT COUNT(*)
FROM ledger_entries
WHERE
    (@actor_id::VARCHAR = '' OR actor_id = @actor_id)
    AND (@subject_id::VARCHAR = '' OR subject_id = @subject_id)
    AND (@subject_type::VARCHAR = '' OR subject_type = @subject_type)
    AND (@from_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at >= @from_time)
    AND (@to_time::TIMESTAMPTZ = '0001-01-01' OR occurred_at <= @to_time);