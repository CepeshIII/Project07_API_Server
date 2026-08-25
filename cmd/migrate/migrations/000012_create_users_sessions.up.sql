CREATE TABLE IF NOT EXISTS users_sessions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id bigserial NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_agent VARCHAR(255),
    ip_address VARCHAR(45),
    is_revoked BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_sessions_user_id ON users_sessions(user_id);