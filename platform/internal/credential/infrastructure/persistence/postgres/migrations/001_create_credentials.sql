-- ============================================================================
-- Credentials Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS credentials (
    -- Primary key
    id                  UUID PRIMARY KEY,
    
    -- Credential identity
    credential_type     VARCHAR(255) NOT NULL,
    schema_id           VARCHAR(255),
    
    -- Participants
    holder_did          VARCHAR(500) NOT NULL,
    issuer_did          VARCHAR(500) NOT NULL,
    
    -- Status
    status              VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Credential data
    claims              JSONB NOT NULL DEFAULT '{}',
    signed_vc           TEXT,
    
    -- Temporal fields
    issued_at           TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ,
    
    -- Revocation
    revoked_at          TIMESTAMPTZ,
    revoked_by          VARCHAR(500),
    revocation_reason   TEXT,
    
    -- Suspension
    suspended_at        TIMESTAMPTZ,
    suspended_by        VARCHAR(500),
    suspension_reason   TEXT,
    suspended_until     TIMESTAMPTZ,
    
    -- Audit
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version             INTEGER NOT NULL DEFAULT 1,
    
    -- Constraints
    CONSTRAINT credentials_status_check 
        CHECK (status IN ('pending', 'active', 'suspended', 'revoked', 'expired'))
);

-- ============================================================================
-- Indexes
-- ============================================================================

-- Lookup by participant
CREATE INDEX IF NOT EXISTS idx_credentials_holder_did ON credentials(holder_did);
CREATE INDEX IF NOT EXISTS idx_credentials_issuer_did ON credentials(issuer_did);

-- Filtering
CREATE INDEX IF NOT EXISTS idx_credentials_status ON credentials(status);
CREATE INDEX IF NOT EXISTS idx_credentials_credential_type ON credentials(credential_type);

-- Sorting / pagination
CREATE INDEX IF NOT EXISTS idx_credentials_created_at ON credentials(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_credentials_issued_at ON credentials(issued_at DESC);

-- Composite index for common query pattern: holder + status
CREATE INDEX IF NOT EXISTS idx_credentials_holder_status ON credentials(holder_did, status);

-- Composite index for common query pattern: issuer + status
CREATE INDEX IF NOT EXISTS idx_credentials_issuer_status ON credentials(issuer_did, status);

-- ============================================================================
-- Updated At Trigger
-- ============================================================================

CREATE OR REPLACE FUNCTION update_credentials_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS credentials_updated_at ON credentials;
CREATE TRIGGER credentials_updated_at
    BEFORE UPDATE ON credentials
    FOR EACH ROW
    EXECUTE FUNCTION update_credentials_updated_at();