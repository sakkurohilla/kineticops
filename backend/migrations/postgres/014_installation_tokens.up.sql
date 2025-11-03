-- Create installation_tokens table for token-based agent installation
CREATE TABLE IF NOT EXISTS installation_tokens (
    id SERIAL PRIMARY KEY,
    token VARCHAR(64) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_installation_tokens_token ON installation_tokens(token);
CREATE INDEX IF NOT EXISTS idx_installation_tokens_user_id ON installation_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_installation_tokens_expires_at ON installation_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_installation_tokens_used ON installation_tokens(used);

-- Add user_id column to hosts table if it doesn't exist
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;