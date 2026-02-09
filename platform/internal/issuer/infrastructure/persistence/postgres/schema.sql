-- ============================================================================
-- Issuer Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Issuer events store
CREATE TABLE IF NOT EXISTS issuer_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Issuer',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_issuer_event_version UNIQUE (aggregate_id, version)
);
CREATE INDEX IF NOT EXISTS idx_issuer_events_aggregate_id ON issuer_events(aggregate_id);

-- Template events store
CREATE TABLE IF NOT EXISTS template_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Template',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_template_event_version UNIQUE (aggregate_id, version)
);
CREATE INDEX IF NOT EXISTS idx_template_events_aggregate_id ON template_events(aggregate_id);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Issuers projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS issuers (
    id                  UUID PRIMARY KEY,
    organization_id     VARCHAR(255) NOT NULL UNIQUE,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    did                 VARCHAR(500),
    webhook_url         VARCHAR(500),
    api_key_hash        VARCHAR(255),
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    branding            JSONB,
    version             INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_issuer_status CHECK (status IN ('pending', 'active', 'suspended', 'revoked'))
);
CREATE INDEX IF NOT EXISTS idx_issuers_organization_id ON issuers(organization_id);
CREATE INDEX IF NOT EXISTS idx_issuers_status ON issuers(status);
CREATE INDEX IF NOT EXISTS idx_issuers_did ON issuers(did);

-- Templates projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS templates (
    id                  UUID PRIMARY KEY,
    issuer_id           UUID NOT NULL REFERENCES issuers(id),
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    schema_type         VARCHAR(100) NOT NULL,
    claim_mappings      JSONB,
    default_values      JSONB,
    expiration_days     INTEGER NOT NULL DEFAULT 0,
    auto_approve        BOOLEAN NOT NULL DEFAULT false,
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    version             INTEGER NOT NULL DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_template_status CHECK (status IN ('active', 'archived'))
);
CREATE INDEX IF NOT EXISTS idx_templates_issuer_id ON templates(issuer_id);
CREATE INDEX IF NOT EXISTS idx_templates_schema_type ON templates(schema_type);
CREATE INDEX IF NOT EXISTS idx_templates_status ON templates(status);
CREATE INDEX IF NOT EXISTS idx_templates_issuer_status ON templates(issuer_id, status);

-- ============================================================================
-- Cross-Context Projection Tables
-- ============================================================================

-- Organization projection (populated via Organization.* events from Organization context)
CREATE TABLE IF NOT EXISTS issuer_organization_projections (
    organization_id     VARCHAR(255) PRIMARY KEY,
    verification_status VARCHAR(50)  NOT NULL DEFAULT 'unverified',
    active              BOOLEAN      NOT NULL DEFAULT true,
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Schema projection (populated via Schema.* events from Schema context)
CREATE TABLE IF NOT EXISTS issuer_schema_projections (
    schema_id   VARCHAR(255) PRIMARY KEY,
    schema_type VARCHAR(255) NOT NULL UNIQUE,
    status      VARCHAR(50)  NOT NULL DEFAULT 'active',
    claims      JSONB,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
