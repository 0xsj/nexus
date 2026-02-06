-- ============================================================================
-- Organization Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Organization events store
CREATE TABLE IF NOT EXISTS organization_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Organization',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_organization_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_organization_events_aggregate_id ON organization_events(aggregate_id);
CREATE INDEX idx_organization_events_event_type ON organization_events(event_type);
CREATE INDEX idx_organization_events_occurred_at ON organization_events(occurred_at);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Organizations projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS organizations (
    id                  UUID PRIMARY KEY,
    name                TEXT NOT NULL,
    slug                TEXT NOT NULL UNIQUE,
    org_type            TEXT NOT NULL,
    description         TEXT,
    verification_status TEXT NOT NULL DEFAULT 'unverified',
    did                 TEXT,
    owner_member_id     TEXT NOT NULL,
    member_count        INTEGER NOT NULL DEFAULT 1,
    deleted             BOOLEAN NOT NULL DEFAULT FALSE,
    version             INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_owner_member_id ON organizations(owner_member_id);
CREATE INDEX idx_organizations_verification_status ON organizations(verification_status);
CREATE INDEX idx_organizations_deleted ON organizations(deleted);
