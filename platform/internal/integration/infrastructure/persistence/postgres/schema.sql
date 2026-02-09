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

-- ============================================================================
-- OAuth Token Storage (Encrypted)
-- ============================================================================

-- Stores encrypted OAuth tokens for integrations
CREATE TABLE IF NOT EXISTS integration_tokens (
    integration_id      UUID PRIMARY KEY REFERENCES integrations(id) ON DELETE CASCADE,

    -- Encrypted token data (AES-256-GCM encrypted)
    access_token_encrypted  BYTEA NOT NULL,
    refresh_token_encrypted BYTEA,

    -- Token metadata
    token_type          VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    expires_at          TIMESTAMPTZ,
    scopes              TEXT[], -- Array of granted scopes

    -- Encryption metadata (for key rotation)
    encryption_version  INTEGER NOT NULL DEFAULT 1,

    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for finding expired tokens
CREATE INDEX IF NOT EXISTS idx_integration_tokens_expires_at
    ON integration_tokens(expires_at)
    WHERE expires_at IS NOT NULL;

-- Index for encryption version (for key rotation)
CREATE INDEX IF NOT EXISTS idx_integration_tokens_encryption_version
    ON integration_tokens(encryption_version);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_integration_tokens_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER integration_tokens_updated_at
    BEFORE UPDATE ON integration_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_integration_tokens_updated_at();
