-- ============================================================================
-- Identity Context Schema
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Event Store Tables (for Event Sourcing)
-- ============================================================================

-- User events store
CREATE TABLE IF NOT EXISTS identity_user_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'User',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure events are ordered per aggregate
    CONSTRAINT unique_user_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_user_events_aggregate_id ON identity_user_events(aggregate_id);
CREATE INDEX idx_user_events_event_type ON identity_user_events(event_type);
CREATE INDEX idx_user_events_occurred_at ON identity_user_events(occurred_at);

-- Session events store
CREATE TABLE IF NOT EXISTS identity_session_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Session',
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    version         INTEGER NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_session_event_version UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_session_events_aggregate_id ON identity_session_events(aggregate_id);
CREATE INDEX idx_session_events_event_type ON identity_session_events(event_type);
CREATE INDEX idx_session_events_occurred_at ON identity_session_events(occurred_at);

-- ============================================================================
-- Read Model / Projection Tables
-- ============================================================================

-- Users projection (denormalized view for queries)
CREATE TABLE IF NOT EXISTS identity_users (
    id              UUID PRIMARY KEY,
    email           VARCHAR(255),
    display_name    VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    primary_did     VARCHAR(500) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version         INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT valid_user_status CHECK (status IN ('pending', 'active', 'suspended', 'deleted'))
);

CREATE UNIQUE INDEX idx_users_email ON identity_users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_status ON identity_users(status);
CREATE INDEX idx_users_display_name ON identity_users(display_name);
CREATE INDEX idx_users_created_at ON identity_users(created_at);

-- User DIDs (one-to-many)
CREATE TABLE IF NOT EXISTS identity_user_dids (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    did         VARCHAR(500) NOT NULL,
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_did UNIQUE (did)
);

CREATE INDEX idx_user_dids_user_id ON identity_user_dids(user_id);
CREATE INDEX idx_user_dids_did ON identity_user_dids(did);

-- User OAuth links (one-to-many)
CREATE TABLE IF NOT EXISTS identity_user_oauth_links (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    provider        VARCHAR(50) NOT NULL,
    external_id     VARCHAR(255) NOT NULL,
    email           VARCHAR(255),
    linked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_oauth_subject UNIQUE (provider, external_id),
    CONSTRAINT unique_user_provider UNIQUE (user_id, provider)
);

CREATE INDEX idx_oauth_links_user_id ON identity_user_oauth_links(user_id);
CREATE INDEX idx_oauth_links_provider ON identity_user_oauth_links(provider);

-- Sessions projection
CREATE TABLE IF NOT EXISTS identity_sessions (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(128) NOT NULL,
    auth_method     VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    ip_address      VARCHAR(45),
    user_agent      VARCHAR(500),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version         INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT valid_session_status CHECK (status IN ('active', 'expired', 'revoked')),
    CONSTRAINT valid_auth_method CHECK (auth_method IN ('magic_link', 'wallet', 'oauth'))
);

CREATE INDEX idx_sessions_user_id ON identity_sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON identity_sessions(token_hash);
CREATE INDEX idx_sessions_status ON identity_sessions(status);
CREATE INDEX idx_sessions_expires_at ON identity_sessions(expires_at);
CREATE INDEX idx_sessions_user_active ON identity_sessions(user_id, status) WHERE status = 'active';

-- ============================================================================
-- Short-lived Token Tables (not event-sourced)
-- ============================================================================

-- Magic links
CREATE TABLE IF NOT EXISTS identity_magic_links (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token_hash  VARCHAR(128) NOT NULL UNIQUE,
    email       VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used        BOOLEAN NOT NULL DEFAULT FALSE,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_magic_links_token_hash ON identity_magic_links(token_hash);
CREATE INDEX idx_magic_links_email ON identity_magic_links(email);
CREATE INDEX idx_magic_links_expires_at ON identity_magic_links(expires_at);

-- OAuth states (CSRF protection)
CREATE TABLE IF NOT EXISTS identity_oauth_states (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    state           VARCHAR(128) NOT NULL UNIQUE,
    provider        VARCHAR(50) NOT NULL,
    redirect_url    VARCHAR(2000) NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_oauth_states_state ON identity_oauth_states(state);
CREATE INDEX idx_oauth_states_expires_at ON identity_oauth_states(expires_at);

-- ============================================================================
-- Cleanup Functions
-- ============================================================================

-- Function to clean up expired magic links
CREATE OR REPLACE FUNCTION cleanup_expired_magic_links()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM identity_magic_links
    WHERE expires_at < NOW() OR (used = TRUE AND used_at < NOW() - INTERVAL '1 hour');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up expired OAuth states
CREATE OR REPLACE FUNCTION cleanup_expired_oauth_states()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM identity_oauth_states
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to mark expired sessions
CREATE OR REPLACE FUNCTION mark_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    UPDATE identity_sessions
    SET status = 'expired', updated_at = NOW()
    WHERE status = 'active' AND expires_at < NOW();
    
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;