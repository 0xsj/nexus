-- ============================================================================
-- Initial Database Setup
-- ============================================================================

-- Ensure extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create application user (if using separate user from superuser)
-- DO $$
-- BEGIN
--     IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'nexus_app') THEN
--         CREATE ROLE nexus_app WITH LOGIN PASSWORD 'nexus_app_password';
--     END IF;
-- END
-- $$;

-- Grant permissions
-- GRANT CONNECT ON DATABASE nexus TO nexus_app;
-- GRANT USAGE ON SCHEMA public TO nexus_app;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO nexus_app;

-- Create schema for credentials (optional, can use public)
-- CREATE SCHEMA IF NOT EXISTS credentials;

-- Log successful initialization
DO $$
BEGIN
    RAISE NOTICE 'Nexus database initialized successfully';
END
$$;