-- Drop function
DROP FUNCTION IF EXISTS delete_expired_sessions();

-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_last_active;
DROP INDEX IF EXISTS idx_sessions_device_id;

-- Drop table
DROP TABLE IF EXISTS sessions;