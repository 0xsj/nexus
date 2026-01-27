-- ============================================================================
-- Verification Module Tables
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Verifications Table
-- Stores verification attempts and their state
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS verifications (
    -- Primary key
    id              VARCHAR(36) PRIMARY KEY,
    
    -- Core fields
    user_id         VARCHAR(36) NOT NULL,
    provider        VARCHAR(50) NOT NULL,
    credential_type VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- OAuth state reference (for correlation)
    oauth_state     VARCHAR(64),
    
    -- Provider data (populated after authorization)
    provider_user_id VARCHAR(255),
    username         VARCHAR(255),
    email            VARCHAR(255),
    display_name     VARCHAR(255),
    avatar_url       TEXT,
    profile_url      TEXT,
    raw_data         JSONB,
    
    -- Credential reference (populated after issuance)
    credential_id   VARCHAR(36),
    
    -- Failure tracking
    failure_reason  TEXT,
    failure_code    VARCHAR(50),
    
    -- Redirect URL for frontend callback
    redirect_url    TEXT,
    
    -- Timestamps
    initiated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    authorized_at   TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,
    
    -- Event sourcing
    version         INTEGER NOT NULL DEFAULT 1,
    
    -- Audit
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for verifications
CREATE INDEX idx_verifications_user_id ON verifications(user_id);
CREATE INDEX idx_verifications_user_provider ON verifications(user_id, provider);
CREATE INDEX idx_verifications_status ON verifications(status);
CREATE INDEX idx_verifications_oauth_state ON verifications(oauth_state) WHERE oauth_state IS NOT NULL;
CREATE INDEX idx_verifications_expires_at ON verifications(expires_at) WHERE status IN ('pending', 'authorized', 'fetching');
CREATE INDEX idx_verifications_initiated_at ON verifications(initiated_at DESC);

-- ----------------------------------------------------------------------------
-- OAuth States Table
-- Stores OAuth state parameters for CSRF protection
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS oauth_states (
    -- State value is the primary key
    state           VARCHAR(64) PRIMARY KEY,
    
    -- Associated verification
    verification_id VARCHAR(36) NOT NULL,
    
    -- Context
    user_id         VARCHAR(36) NOT NULL,
    provider        VARCHAR(50) NOT NULL,
    
    -- Security
    code_verifier   VARCHAR(128),  -- For PKCE
    nonce           VARCHAR(64),
    
    -- Redirect after completion
    redirect_url    TEXT,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT fk_oauth_states_verification
        FOREIGN KEY (verification_id) 
        REFERENCES verifications(id) 
        ON DELETE CASCADE
);

-- Indexes for oauth_states
CREATE INDEX idx_oauth_states_verification_id ON oauth_states(verification_id);
CREATE INDEX idx_oauth_states_expires_at ON oauth_states(expires_at) WHERE used_at IS NULL;
CREATE INDEX idx_oauth_states_user_provider ON oauth_states(user_id, provider);

-- ----------------------------------------------------------------------------
-- Provider Tokens Table
-- Stores encrypted OAuth tokens for provider access
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS provider_tokens (
    -- Composite primary key
    user_id         VARCHAR(36) NOT NULL,
    provider        VARCHAR(50) NOT NULL,
    
    -- Encrypted tokens (encrypted at application level)
    access_token    TEXT NOT NULL,
    refresh_token   TEXT,
    token_type      VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    
    -- Token metadata
    scopes          TEXT[],
    expires_at      TIMESTAMPTZ,
    
    -- Provider user info (for quick lookups)
    provider_user_id VARCHAR(255),
    
    -- Audit
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at    TIMESTAMPTZ,
    
    -- Primary key
    PRIMARY KEY (user_id, provider)
);

-- Indexes for provider_tokens
CREATE INDEX idx_provider_tokens_user_id ON provider_tokens(user_id);
CREATE INDEX idx_provider_tokens_expires_at ON provider_tokens(expires_at) WHERE expires_at IS NOT NULL;

-- ----------------------------------------------------------------------------
-- Verification Events Table (Event Store)
-- Stores domain events for event sourcing
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS verification_events (
    -- Event identification
    id              VARCHAR(36) PRIMARY KEY,
    aggregate_id    VARCHAR(36) NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL DEFAULT 'Verification',
    
    -- Event data
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    
    -- Versioning
    version         INTEGER NOT NULL,
    
    -- Metadata
    metadata        JSONB,
    
    -- Timestamp
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure version uniqueness per aggregate
    CONSTRAINT uq_verification_events_aggregate_version 
        UNIQUE (aggregate_id, version)
);

-- Indexes for verification_events
CREATE INDEX idx_verification_events_aggregate_id ON verification_events(aggregate_id);
CREATE INDEX idx_verification_events_aggregate_type ON verification_events(aggregate_type);
CREATE INDEX idx_verification_events_event_type ON verification_events(event_type);
CREATE INDEX idx_verification_events_occurred_at ON verification_events(occurred_at DESC);

-- ----------------------------------------------------------------------------
-- Functions
-- ----------------------------------------------------------------------------

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_verification_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER trg_verifications_updated_at
    BEFORE UPDATE ON verifications
    FOR EACH ROW
    EXECUTE FUNCTION update_verification_updated_at();

CREATE TRIGGER trg_provider_tokens_updated_at
    BEFORE UPDATE ON provider_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_verification_updated_at();

-- ----------------------------------------------------------------------------
-- Comments
-- ----------------------------------------------------------------------------

COMMENT ON TABLE verifications IS 'Stores verification attempts linking users to external providers';
COMMENT ON TABLE oauth_states IS 'Stores OAuth state parameters for CSRF protection during OAuth flow';
COMMENT ON TABLE provider_tokens IS 'Stores encrypted OAuth tokens for accessing provider APIs';
COMMENT ON TABLE verification_events IS 'Event store for verification aggregate events';

COMMENT ON COLUMN verifications.status IS 'pending, authorized, fetching, completed, failed, expired, cancelled';
COMMENT ON COLUMN verifications.raw_data IS 'Raw JSON data from provider API response';
COMMENT ON COLUMN oauth_states.code_verifier IS 'PKCE code verifier for enhanced OAuth security';
COMMENT ON COLUMN provider_tokens.access_token IS 'Encrypted OAuth access token';
COMMENT ON COLUMN provider_tokens.refresh_token IS 'Encrypted OAuth refresh token';