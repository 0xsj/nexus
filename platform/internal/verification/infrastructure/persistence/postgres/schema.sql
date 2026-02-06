-- ============================================================================
-- Verification Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Verification events store
CREATE TABLE IF NOT EXISTS verification_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Verification',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_verification_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_verification_events_aggregate_id ON verification_events(aggregate_id);
CREATE INDEX idx_verification_events_event_type ON verification_events(event_type);
CREATE INDEX idx_verification_events_occurred_at ON verification_events(occurred_at);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Verifications projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS verifications (
    id              UUID PRIMARY KEY,
    user_id         TEXT NOT NULL,
    provider_type   TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'oauth_started',
    oauth_state     TEXT NOT NULL,
    error_message   TEXT,
    error_code      TEXT,
    credential_id   TEXT,
    version         INTEGER NOT NULL DEFAULT 0,
    started_at      TIMESTAMPTZ NOT NULL,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verifications_user_id ON verifications(user_id);
CREATE INDEX idx_verifications_provider_type ON verifications(provider_type);
CREATE INDEX idx_verifications_status ON verifications(status);
CREATE INDEX idx_verifications_user_provider ON verifications(user_id, provider_type);
