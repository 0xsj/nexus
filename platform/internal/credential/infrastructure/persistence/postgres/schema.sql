-- ============================================================================
-- Credential Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Credential events store
CREATE TABLE IF NOT EXISTS credential_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Credential',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_credential_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_credential_events_aggregate_id ON credential_events(aggregate_id);
CREATE INDEX idx_credential_events_event_type ON credential_events(event_type);
CREATE INDEX idx_credential_events_occurred_at ON credential_events(occurred_at);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Credentials projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS credentials (
    id                  UUID PRIMARY KEY,
    credential_type     VARCHAR(100) NOT NULL,
    issuer_did          VARCHAR(500) NOT NULL,
    subject_did         VARCHAR(500) NOT NULL,
    claims              JSONB NOT NULL,
    issued_at           TIMESTAMPTZ NOT NULL,
    expires_at          TIMESTAMPTZ,
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    revoked_at          TIMESTAMPTZ,
    revocation_reason   TEXT,
    jwt                 TEXT,
    verification_id     VARCHAR(255),
    version             INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_credential_status CHECK (status IN ('active', 'revoked', 'expired'))
);

CREATE INDEX idx_credentials_subject_did ON credentials(subject_did);
CREATE INDEX idx_credentials_issuer_did ON credentials(issuer_did);
CREATE INDEX idx_credentials_status ON credentials(status);
CREATE INDEX idx_credentials_subject_status ON credentials(subject_did, status);
CREATE INDEX idx_credentials_issuer_status ON credentials(issuer_did, status);
