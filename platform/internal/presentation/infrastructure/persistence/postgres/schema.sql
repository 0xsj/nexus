-- ============================================================================
-- Presentation Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Presentation events
CREATE TABLE IF NOT EXISTS presentation_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Presentation',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_presentation_event_version UNIQUE (aggregate_id, version)
);
CREATE INDEX IF NOT EXISTS idx_presentation_events_aggregate_id ON presentation_events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_presentation_events_event_type ON presentation_events(event_type);

-- ShareLink events
CREATE TABLE IF NOT EXISTS share_link_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'ShareLink',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_share_link_event_version UNIQUE (aggregate_id, version)
);
CREATE INDEX IF NOT EXISTS idx_share_link_events_aggregate_id ON share_link_events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_share_link_events_event_type ON share_link_events(event_type);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Presentations projection
CREATE TABLE IF NOT EXISTS presentations (
    id                  UUID PRIMARY KEY,
    holder_did          VARCHAR(500) NOT NULL,
    credential_ids      JSONB NOT NULL,
    disclosure_policy   JSONB NOT NULL,
    vp_jwt              TEXT,
    purpose             VARCHAR(500),
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    revoked_at          TIMESTAMPTZ,
    revocation_reason   TEXT,
    version             INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_presentation_status CHECK (status IN ('active', 'revoked'))
);
CREATE INDEX IF NOT EXISTS idx_presentations_holder_did ON presentations(holder_did);
CREATE INDEX IF NOT EXISTS idx_presentations_status ON presentations(status);
CREATE INDEX IF NOT EXISTS idx_presentations_holder_status ON presentations(holder_did, status);

-- ShareLinks projection
CREATE TABLE IF NOT EXISTS share_links (
    id                  UUID PRIMARY KEY,
    presentation_id     UUID NOT NULL REFERENCES presentations(id),
    token               VARCHAR(255) NOT NULL UNIQUE,
    expires_at          TIMESTAMPTZ,
    max_views           INTEGER NOT NULL DEFAULT 0,
    current_views       INTEGER NOT NULL DEFAULT 0,
    pin_hash            VARCHAR(255),
    audience            VARCHAR(500),
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    version             INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_share_link_status CHECK (status IN ('active', 'revoked', 'expired'))
);
CREATE INDEX IF NOT EXISTS idx_share_links_presentation_id ON share_links(presentation_id);
CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);
CREATE INDEX IF NOT EXISTS idx_share_links_status ON share_links(status);

-- Access grants (read-only log)
CREATE TABLE IF NOT EXISTS access_grants (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    share_link_id       UUID NOT NULL REFERENCES share_links(id),
    verifier_did        VARCHAR(500),
    accessed_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address          VARCHAR(45),
    disclosed_claims    JSONB
);
CREATE INDEX IF NOT EXISTS idx_access_grants_share_link_id ON access_grants(share_link_id);
