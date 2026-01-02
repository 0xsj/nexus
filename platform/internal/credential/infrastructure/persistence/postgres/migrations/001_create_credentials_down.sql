-- Drop credentials table and related objects
DROP TRIGGER IF EXISTS credentials_updated_at ON credentials;
DROP FUNCTION IF EXISTS update_credentials_updated_at();
DROP TABLE IF EXISTS credentials;