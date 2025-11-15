-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255),
    device_name VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Indexes
    CONSTRAINT sessions_expires_at_check CHECK (expires_at > created_at)
);

-- Create indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_last_active ON sessions(last_active_at DESC);
CREATE INDEX idx_sessions_device_id ON sessions(device_id) WHERE device_id IS NOT NULL;

-- Create cleanup function for expired sessions
CREATE OR REPLACE FUNCTION delete_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM sessions WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- Add comments
COMMENT ON TABLE sessions IS 'User authentication sessions';
COMMENT ON COLUMN sessions.user_id IS 'Reference to users table';
COMMENT ON COLUMN sessions.device_id IS 'Unique device identifier';
COMMENT ON COLUMN sessions.expires_at IS 'Session expiration timestamp';