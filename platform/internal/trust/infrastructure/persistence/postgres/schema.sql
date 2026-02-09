-- ============================================================================
-- Trust Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Vouch events store
CREATE TABLE IF NOT EXISTS vouch_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Vouch',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_vouch_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX IF NOT EXISTS idx_vouch_events_aggregate_id ON vouch_events(aggregate_id);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Vouches projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS vouches (
    id                  UUID PRIMARY KEY,
    voucher_id          VARCHAR(255) NOT NULL,
    vouchee_id          VARCHAR(255) NOT NULL,
    credential_id       VARCHAR(255),
    claim_key           VARCHAR(255),
    relationship        VARCHAR(50) NOT NULL,
    strength            INTEGER NOT NULL,
    statement           TEXT,
    context             TEXT,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at          TIMESTAMPTZ,
    accepted_at         TIMESTAMPTZ,
    revoked_at          TIMESTAMPTZ,
    revocation_reason   TEXT,
    version             INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_vouch_status CHECK (status IN ('pending', 'accepted', 'revoked', 'expired'))
);

CREATE INDEX IF NOT EXISTS idx_vouches_voucher_id ON vouches(voucher_id);
CREATE INDEX IF NOT EXISTS idx_vouches_vouchee_id ON vouches(vouchee_id);
CREATE INDEX IF NOT EXISTS idx_vouches_status ON vouches(status);
CREATE INDEX IF NOT EXISTS idx_vouches_vouchee_status ON vouches(vouchee_id, status);
CREATE INDEX IF NOT EXISTS idx_vouches_relationship ON vouches(relationship);
CREATE INDEX IF NOT EXISTS idx_vouches_credential_id ON vouches(credential_id);

-- Reputations projection
CREATE TABLE IF NOT EXISTS reputations (
    user_id             VARCHAR(255) PRIMARY KEY,
    overall_score       INTEGER NOT NULL DEFAULT 0,
    credential_score    INTEGER NOT NULL DEFAULT 0,
    vouch_score         INTEGER NOT NULL DEFAULT 0,
    network_score       INTEGER NOT NULL DEFAULT 0,
    vouch_count         INTEGER NOT NULL DEFAULT 0,
    last_calculated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reputations_overall_score ON reputations(overall_score DESC);

-- ============================================================================
-- Cross-Context Projection Tables
-- ============================================================================

-- Identity user projection (populated via User.* events from Identity context)
CREATE TABLE IF NOT EXISTS trust_user_projections (
    user_id     VARCHAR(255) PRIMARY KEY,
    active      BOOLEAN NOT NULL DEFAULT true,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Credential projection (populated via Credential.* events from Credential context)
CREATE TABLE IF NOT EXISTS trust_credential_projections (
    credential_id   VARCHAR(255) PRIMARY KEY,
    credential_type VARCHAR(255) NOT NULL DEFAULT '',
    subject_did     VARCHAR(512) NOT NULL DEFAULT '',
    user_id         VARCHAR(255) NOT NULL DEFAULT '',
    status          VARCHAR(50)  NOT NULL DEFAULT 'active',
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_trust_cred_proj_subject ON trust_credential_projections(subject_did);
CREATE INDEX IF NOT EXISTS idx_trust_cred_proj_user ON trust_credential_projections(user_id);

-- Organization projection (populated via Organization.* events from Organization context)
CREATE TABLE IF NOT EXISTS trust_organization_projections (
    organization_id     VARCHAR(255) PRIMARY KEY,
    verification_status VARCHAR(50)  NOT NULL DEFAULT 'unverified',
    active              BOOLEAN      NOT NULL DEFAULT true,
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
