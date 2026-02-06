-- ============================================================================
-- Wallet Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Wallet events store
CREATE TABLE IF NOT EXISTS wallet_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Wallet',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_wallet_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_wallet_events_aggregate_id ON wallet_events(aggregate_id);
CREATE INDEX idx_wallet_events_event_type ON wallet_events(event_type);
CREATE INDEX idx_wallet_events_occurred_at ON wallet_events(occurred_at);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Wallets projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS wallets (
    id              UUID PRIMARY KEY,
    user_id         VARCHAR(255) NOT NULL,
    address         VARCHAR(100) NOT NULL UNIQUE,
    chain           VARCHAR(50) NOT NULL,
    label           VARCHAR(100),
    status          VARCHAR(20) NOT NULL DEFAULT 'unverified',
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    did             VARCHAR(500),
    linked_at       TIMESTAMPTZ NOT NULL,
    verified_at     TIMESTAMPTZ,
    version         INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_wallet_status CHECK (status IN ('unverified', 'active', 'revoked'))
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_address ON wallets(address);
CREATE INDEX idx_wallets_chain ON wallets(chain);
CREATE INDEX idx_wallets_user_primary ON wallets(user_id, is_primary);
