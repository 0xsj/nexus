-- Identity Domain Tables
-- Supports DID-first, wallet-native authentication with Web2 bridges

-- ============================================================================
-- Users Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id              VARCHAR(36) PRIMARY KEY,
    did             VARCHAR(255) NOT NULL UNIQUE,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Metadata
    display_name    VARCHAR(255),
    avatar_url      VARCHAR(500),
    bio             TEXT,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended'))
);

CREATE INDEX IF NOT EXISTS idx_users_did ON users(did);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- ============================================================================
-- User Wallets Table (Linked Wallet Addresses)
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_wallets (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address         VARCHAR(255) NOT NULL,
    chain           VARCHAR(50) NOT NULL,
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Verification
    verified        BOOLEAN NOT NULL DEFAULT TRUE,
    verified_at     TIMESTAMPTZ,
    
    -- Timestamps
    linked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT user_wallets_unique_address_chain UNIQUE (address, chain),
    CONSTRAINT user_wallets_chain_check CHECK (chain IN ('ethereum', 'polygon', 'arbitrum', 'optimism', 'base', 'solana', 'cosmos'))
);

CREATE INDEX IF NOT EXISTS idx_user_wallets_user_id ON user_wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_user_wallets_address_chain ON user_wallets(address, chain);

-- ============================================================================
-- User Emails Table (Web2 Bridge)
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_emails (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Verification
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at     TIMESTAMPTZ,
    
    -- Timestamps
    linked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT user_emails_email_lower CHECK (email = LOWER(email))
);

CREATE INDEX IF NOT EXISTS idx_user_emails_user_id ON user_emails(user_id);
CREATE INDEX IF NOT EXISTS idx_user_emails_email ON user_emails(email);

-- ============================================================================
-- Sessions Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS sessions (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(64) NOT NULL UNIQUE,
    
    -- Auth context
    auth_method     VARCHAR(50) NOT NULL,
    wallet_address  VARCHAR(255),
    chain           VARCHAR(50),
    did             VARCHAR(255),
    
    -- Device info
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    device_id       VARCHAR(255),
    
    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    last_active_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT sessions_status_check CHECK (status IN ('active', 'expired', 'revoked')),
    CONSTRAINT sessions_auth_method_check CHECK (auth_method IN ('wallet', 'did_auth', 'passkey', 'oauth', 'vp'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- ============================================================================
-- Connections Table (OAuth Providers)
-- ============================================================================

CREATE TABLE IF NOT EXISTS connections (
    id                  VARCHAR(36) PRIMARY KEY,
    user_id             VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider            VARCHAR(50) NOT NULL,
    provider_user_id    VARCHAR(255) NOT NULL,
    
    -- Profile data
    provider_username   VARCHAR(255),
    provider_email      VARCHAR(255),
    provider_avatar     VARCHAR(500),
    profile_data        JSONB,
    
    -- Tokens (encrypted in production)
    access_token        TEXT,
    refresh_token       TEXT,
    token_expires_at    TIMESTAMPTZ,
    scopes              TEXT[],
    
    -- Status
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Sync tracking
    last_synced_at      TIMESTAMPTZ,
    sync_error          TEXT,
    
    -- Timestamps
    connected_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disconnected_at     TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT connections_unique_user_provider UNIQUE (user_id, provider),
    CONSTRAINT connections_unique_provider_user UNIQUE (provider, provider_user_id),
    CONSTRAINT connections_provider_check CHECK (provider IN ('github', 'linkedin', 'google', 'twitter', 'discord')),
    CONSTRAINT connections_status_check CHECK (status IN ('active', 'inactive', 'revoked', 'expired'))
);

CREATE INDEX IF NOT EXISTS idx_connections_user_id ON connections(user_id);
CREATE INDEX IF NOT EXISTS idx_connections_provider ON connections(provider);
CREATE INDEX IF NOT EXISTS idx_connections_provider_user_id ON connections(provider, provider_user_id);
CREATE INDEX IF NOT EXISTS idx_connections_status ON connections(status);
CREATE INDEX IF NOT EXISTS idx_connections_last_synced_at ON connections(last_synced_at);

-- ============================================================================
-- API Keys Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS api_keys (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    
    -- Key data
    key_hash        VARCHAR(64) NOT NULL UNIQUE,
    key_prefix      VARCHAR(12) NOT NULL,
    
    -- Permissions
    scopes          TEXT[] NOT NULL DEFAULT '{}',
    
    -- Rate limiting
    rate_limit      INTEGER,
    
    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    
    -- Usage tracking
    last_used_at    TIMESTAMPTZ,
    usage_count     BIGINT NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT api_keys_status_check CHECK (status IN ('active', 'expired', 'revoked'))
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_status ON api_keys(status);

-- ============================================================================
-- Refresh Tokens Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id      VARCHAR(36) NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    token_hash      VARCHAR(64) NOT NULL UNIQUE,
    
    -- Status
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id ON refresh_tokens(session_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- ============================================================================
-- Triggers for updated_at
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_connections_updated_at
    BEFORE UPDATE ON connections
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();