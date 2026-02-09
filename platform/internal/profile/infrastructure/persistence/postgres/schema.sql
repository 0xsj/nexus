-- ============================================================================
-- Profile Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- Profile events store
CREATE TABLE IF NOT EXISTS profile_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Profile',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_profile_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_profile_events_aggregate_id ON profile_events(aggregate_id);
CREATE INDEX idx_profile_events_event_type ON profile_events(event_type);
CREATE INDEX idx_profile_events_occurred_at ON profile_events(occurred_at);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Profiles projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS profiles (
    id              UUID PRIMARY KEY,
    user_id         TEXT NOT NULL UNIQUE,
    display_name    TEXT NOT NULL,
    headline        TEXT NOT NULL DEFAULT '',
    bio             TEXT NOT NULL DEFAULT '',
    vanity_slug     TEXT NOT NULL UNIQUE,
    badge_count     INTEGER NOT NULL DEFAULT 0,
    version         INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);
CREATE INDEX idx_profiles_vanity_slug ON profiles(vanity_slug);

-- ============================================================================
-- Cross-Context Projection Tables
-- ============================================================================

-- Credential projection (populated via Credential.* events from Credential context)
CREATE TABLE IF NOT EXISTS profile_credential_projections (
    credential_id   VARCHAR(255) PRIMARY KEY,
    credential_type VARCHAR(255) NOT NULL DEFAULT '',
    status          VARCHAR(50)  NOT NULL DEFAULT 'active',
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
