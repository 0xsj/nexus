-- ============================================================================
-- Ledger Schema
-- ============================================================================
-- Audit trail for all domain events across the platform.
-- Append-only table optimized for time-range queries.
-- ============================================================================

-- Audit entries table
CREATE TABLE IF NOT EXISTS ledger_entries (
    -- Primary key
    id UUID PRIMARY KEY,

    -- Timestamps
    occurred_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Event identification
    event_type VARCHAR(255) NOT NULL,

    -- Actor (who caused the event)
    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50) NOT NULL,

    -- Subject (what the event is about)
    subject_id VARCHAR(255) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,

    -- Additional context
    metadata JSONB NOT NULL DEFAULT '{}',
    context_id VARCHAR(255)
);

-- ============================================================================
-- Indexes
-- ============================================================================

-- Time-based queries (most common access pattern)
CREATE INDEX IF NOT EXISTS idx_ledger_entries_occurred_at 
    ON ledger_entries (occurred_at DESC);

-- Actor queries (user activity)
CREATE INDEX IF NOT EXISTS idx_ledger_entries_actor 
    ON ledger_entries (actor_id, actor_type, occurred_at DESC);

-- Subject queries (credential history, etc.)
CREATE INDEX IF NOT EXISTS idx_ledger_entries_subject 
    ON ledger_entries (subject_id, subject_type, occurred_at DESC);

-- Event type filtering
CREATE INDEX IF NOT EXISTS idx_ledger_entries_event_type 
    ON ledger_entries (event_type, occurred_at DESC);

-- Correlation ID for tracing related events
CREATE INDEX IF NOT EXISTS idx_ledger_entries_context_id 
    ON ledger_entries (context_id) 
    WHERE context_id IS NOT NULL;

-- Composite index for common filtered queries
CREATE INDEX IF NOT EXISTS idx_ledger_entries_subject_event_type 
    ON ledger_entries (subject_id, subject_type, event_type, occurred_at DESC);