-- ============================================================================
-- Wallet Tables Migration
-- ============================================================================

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Wallets Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS wallets (
    id                  VARCHAR(64) PRIMARY KEY,
    user_id             VARCHAR(64) NOT NULL,
    address             VARCHAR(256) NOT NULL,
    address_normalized  VARCHAR(256) NOT NULL,
    chain_id            VARCHAR(64) NOT NULL,
    chain_family        VARCHAR(32) NOT NULL,
    did                 VARCHAR(512) NOT NULL,
    label               VARCHAR(128),
    is_primary          BOOLEAN NOT NULL DEFAULT FALSE,
    status              VARCHAR(32) NOT NULL DEFAULT 'active',
    verified_at         TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at        TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT wallets_address_chain_unique UNIQUE (address_normalized, chain_id),
    CONSTRAINT wallets_did_unique UNIQUE (did),
    CONSTRAINT wallets_status_check CHECK (status IN ('active', 'inactive', 'suspended'))
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_address_normalized ON wallets(address_normalized);
CREATE INDEX IF NOT EXISTS idx_wallets_chain_id ON wallets(chain_id);
CREATE INDEX IF NOT EXISTS idx_wallets_did ON wallets(did);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id_primary ON wallets(user_id, is_primary) WHERE is_primary = TRUE;
CREATE INDEX IF NOT EXISTS idx_wallets_user_id_status ON wallets(user_id, status);
CREATE INDEX IF NOT EXISTS idx_wallets_created_at ON wallets(created_at DESC);

-- ============================================================================
-- Wallet Challenges Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS wallet_challenges (
    nonce               VARCHAR(128) PRIMARY KEY,
    message             TEXT NOT NULL,
    address             VARCHAR(256) NOT NULL,
    address_normalized  VARCHAR(256) NOT NULL,
    chain_id            VARCHAR(64) NOT NULL,
    domain              VARCHAR(256) NOT NULL,
    uri                 VARCHAR(512) NOT NULL,
    issued_at           TIMESTAMPTZ NOT NULL,
    expires_at          TIMESTAMPTZ NOT NULL,
    used                BOOLEAN NOT NULL DEFAULT FALSE,
    used_at             TIMESTAMPTZ
);

-- Indexes for challenge lookups
CREATE INDEX IF NOT EXISTS idx_wallet_challenges_expires_at ON wallet_challenges(expires_at);
CREATE INDEX IF NOT EXISTS idx_wallet_challenges_address ON wallet_challenges(address_normalized, chain_id);
CREATE INDEX IF NOT EXISTS idx_wallet_challenges_used ON wallet_challenges(used) WHERE used = FALSE;

-- ============================================================================
-- Wallet Nonces Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS wallet_nonces (
    nonce               VARCHAR(128) PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMPTZ NOT NULL,
    used                BOOLEAN NOT NULL DEFAULT FALSE,
    used_at             TIMESTAMPTZ
);

-- Index for expiration cleanup
CREATE INDEX IF NOT EXISTS idx_wallet_nonces_expires_at ON wallet_nonces(expires_at);
CREATE INDEX IF NOT EXISTS idx_wallet_nonces_used ON wallet_nonces(used) WHERE used = FALSE;

-- ============================================================================
-- Wallet Activities Table (Audit Log)
-- ============================================================================

CREATE TABLE IF NOT EXISTS wallet_activities (
    id                  VARCHAR(64) PRIMARY KEY,
    wallet_id           VARCHAR(64) NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    action              VARCHAR(64) NOT NULL,
    ip_address          VARCHAR(64),
    user_agent          VARCHAR(512),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for activity queries
CREATE INDEX IF NOT EXISTS idx_wallet_activities_wallet_id ON wallet_activities(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_activities_created_at ON wallet_activities(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_activities_action ON wallet_activities(action);

-- ============================================================================
-- Updated At Trigger Function
-- ============================================================================

CREATE OR REPLACE FUNCTION update_wallet_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to wallets table
DROP TRIGGER IF EXISTS trigger_wallets_updated_at ON wallets;
CREATE TRIGGER trigger_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_wallet_updated_at();

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE wallets IS 'Stores user wallet addresses linked to the platform';
COMMENT ON COLUMN wallets.address IS 'Original wallet address as provided';
COMMENT ON COLUMN wallets.address_normalized IS 'Normalized/checksummed wallet address for lookups';
COMMENT ON COLUMN wallets.chain_id IS 'Blockchain chain ID (e.g., 1 for Ethereum mainnet)';
COMMENT ON COLUMN wallets.chain_family IS 'Blockchain family (evm, solana, cosmos)';
COMMENT ON COLUMN wallets.did IS 'Derived did:pkh DID for this wallet';
COMMENT ON COLUMN wallets.is_primary IS 'Whether this is the users primary wallet';
COMMENT ON COLUMN wallets.status IS 'Wallet status: active, inactive, suspended';

COMMENT ON TABLE wallet_challenges IS 'Stores SIWE authentication challenges';
COMMENT ON COLUMN wallet_challenges.nonce IS 'Unique challenge identifier';
COMMENT ON COLUMN wallet_challenges.message IS 'Full SIWE message to be signed';
COMMENT ON COLUMN wallet_challenges.used IS 'Whether the challenge has been consumed';

COMMENT ON TABLE wallet_nonces IS 'Stores nonces for replay protection';
COMMENT ON COLUMN wallet_nonces.nonce IS 'Unique nonce value';
COMMENT ON COLUMN wallet_nonces.used IS 'Whether the nonce has been consumed';

COMMENT ON TABLE wallet_activities IS 'Audit log for wallet activities';
COMMENT ON COLUMN wallet_activities.action IS 'Activity type: link, unlink, login, sign, verify';