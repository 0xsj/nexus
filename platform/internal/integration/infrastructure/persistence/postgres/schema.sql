-- ============================================================================
-- Integration Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Integration events store
CREATE TABLE IF NOT EXISTS integration_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Integration',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_integration_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX IF NOT EXISTS idx_integration_events_aggregate_id ON integration_events(aggregate_id);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Integrations projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS integrations (
    id                  UUID PRIMARY KEY,
    user_id             VARCHAR(255) NOT NULL,
    provider_type       VARCHAR(50) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'connected',
    provider_user_id    VARCHAR(255),
    provider_username   VARCHAR(255),
    scopes              JSONB,
    last_fetch_at       TIMESTAMPTZ,
    fetch_count         INTEGER NOT NULL DEFAULT 0,
    metadata            JSONB,
    connected_at        TIMESTAMPTZ,
    disconnected_at     TIMESTAMPTZ,
    suspended_at        TIMESTAMPTZ,
    suspension_reason   TEXT,
    version             INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_integration_status CHECK (status IN ('connected', 'disconnected', 'error', 'suspended')),
    CONSTRAINT unique_user_provider UNIQUE (user_id, provider_type)
);

CREATE INDEX IF NOT EXISTS idx_integrations_user_id ON integrations(user_id);
CREATE INDEX IF NOT EXISTS idx_integrations_provider_type ON integrations(provider_type);
CREATE INDEX IF NOT EXISTS idx_integrations_status ON integrations(status);
CREATE INDEX IF NOT EXISTS idx_integrations_user_status ON integrations(user_id, status);
