-- ============================================================================
-- Verification Module Tables - Rollback
-- ============================================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS trg_verifications_updated_at ON verifications;
DROP TRIGGER IF EXISTS trg_provider_tokens_updated_at ON provider_tokens;

-- Drop function
DROP FUNCTION IF EXISTS update_verification_updated_at();

-- Drop tables in reverse order of creation (respecting foreign keys)
DROP TABLE IF EXISTS verification_events;
DROP TABLE IF EXISTS provider_tokens;
DROP TABLE IF EXISTS oauth_states;
DROP TABLE IF EXISTS verifications;