-- Migration: 002_add_linked_dids
-- Description: Adds multi-DID support to identity module
-- Created: 2025-01-06

-- ============================================================================
-- Alter Users Table
-- ============================================================================

-- Add last_login_method column
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS last_login_method VARCHAR(50);

-- Add constraint for valid auth methods
ALTER TABLE users 
ADD CONSTRAINT users_last_login_method_check 
CHECK (last_login_method IS NULL OR last_login_method IN ('wallet', 'email', 'did_auth', 'passkey', 'oauth', 'vp'));

-- Update sessions auth_method constraint to include 'email'
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_auth_method_check;
ALTER TABLE sessions 
ADD CONSTRAINT sessions_auth_method_check 
CHECK (auth_method IN ('wallet', 'email', 'did_auth', 'passkey', 'oauth', 'vp'));

-- ============================================================================
-- User Linked DIDs Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_linked_dids (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    did             VARCHAR(500) NOT NULL,
    source          VARCHAR(50) NOT NULL,
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Metadata
    label           VARCHAR(255),
    metadata        JSONB,
    
    -- Timestamps
    linked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at    TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT user_linked_dids_unique_did UNIQUE (did),
    CONSTRAINT user_linked_dids_source_check CHECK (source IN ('custodial', 'wallet', 'imported'))
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_user_linked_dids_user_id ON user_linked_dids(user_id);
CREATE INDEX IF NOT EXISTS idx_user_linked_dids_did ON user_linked_dids(did);
CREATE INDEX IF NOT EXISTS idx_user_linked_dids_source ON user_linked_dids(source);
CREATE INDEX IF NOT EXISTS idx_user_linked_dids_is_primary ON user_linked_dids(user_id, is_primary) WHERE is_primary = TRUE;

-- Ensure only one primary DID per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_linked_dids_one_primary 
ON user_linked_dids(user_id) 
WHERE is_primary = TRUE;

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE user_linked_dids IS 'Stores all DIDs associated with a user (custodial, wallet-derived, imported)';
COMMENT ON COLUMN user_linked_dids.did IS 'The decentralized identifier string (e.g., did:key:z6Mk..., did:pkh:eip155:1:0x...)';
COMMENT ON COLUMN user_linked_dids.source IS 'How the DID was created: custodial (Nexus-generated), wallet (derived from wallet), imported (user-provided)';
COMMENT ON COLUMN user_linked_dids.is_primary IS 'Whether this is the user''s primary DID for credential issuance';
COMMENT ON COLUMN user_linked_dids.label IS 'User-friendly name for this DID (e.g., "Personal Wallet", "Work Identity")';
COMMENT ON COLUMN user_linked_dids.metadata IS 'Source-specific data (wallet address, chain ID, key algorithm, etc.)';