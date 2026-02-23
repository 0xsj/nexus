-- ============================================================================
-- Schema Context Database Schema
-- ============================================================================
-- Credential schema definitions and versions.
-- Supports both built-in schemas and custom issuer schemas.
-- ============================================================================

-- Schemas table (main aggregate)
CREATE TABLE IF NOT EXISTS schemas (
    -- Primary key
    id UUID PRIMARY KEY,

    -- Schema identification
    schema_type VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',

    -- Current version
    current_version VARCHAR(50) NOT NULL,

    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'active',

    -- Optional issuer (NULL for built-in schemas)
    issuer_id UUID,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Schema claims table (claim definitions for each schema)
CREATE TABLE IF NOT EXISTS schema_claims (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign key to schema
    schema_id UUID NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,

    -- Claim definition
    key VARCHAR(64) NOT NULL,
    data_type VARCHAR(50) NOT NULL,
    required BOOLEAN NOT NULL DEFAULT false,
    display_name VARCHAR(128),
    description VARCHAR(512),
    disclosable BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,

    -- Constraints (stored as JSONB for flexibility)
    constraints JSONB,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique claim key per schema
    UNIQUE (schema_id, key)
);

-- Schema versions table (version history)
CREATE TABLE IF NOT EXISTS schema_versions (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign key to schema
    schema_id UUID NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,

    -- Version info
    version VARCHAR(50) NOT NULL,
    change_summary TEXT,

    -- Snapshot of claims at this version (denormalized for historical queries)
    claims_snapshot JSONB NOT NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique version per schema
    UNIQUE (schema_id, version)
);

-- ============================================================================
-- Indexes
-- ============================================================================

-- Schema lookups
CREATE INDEX IF NOT EXISTS idx_schemas_schema_type
ON schemas (schema_type);

CREATE INDEX IF NOT EXISTS idx_schemas_status
ON schemas (status);

CREATE INDEX IF NOT EXISTS idx_schemas_issuer_id
ON schemas (issuer_id)
WHERE issuer_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_schemas_created_at
ON schemas (created_at DESC);

-- Schema claims lookups
CREATE INDEX IF NOT EXISTS idx_schema_claims_schema_id
ON schema_claims (schema_id);

CREATE INDEX IF NOT EXISTS idx_schema_claims_key
ON schema_claims (schema_id, key);

-- Schema versions lookups
CREATE INDEX IF NOT EXISTS idx_schema_versions_schema_id
ON schema_versions (schema_id);

CREATE INDEX IF NOT EXISTS idx_schema_versions_version
ON schema_versions (schema_id, version);

-- Full-text search on name and description
CREATE INDEX IF NOT EXISTS idx_schemas_search
ON schemas USING GIN (to_tsvector('english', name || ' ' || description));

-- ============================================================================
-- Issuer Projections (Cross-Context Read Model)
-- ============================================================================
-- Populated by IssuerProjector consuming Issuer.* events from the Issuer context.
-- Used by Schema command handlers to validate issuer ownership of custom schemas.

CREATE TABLE IF NOT EXISTS schema_issuer_projections (
    issuer_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);