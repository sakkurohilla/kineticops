-- Add revoked flag and revoked_at timestamp to agents table
ALTER TABLE agents
    ADD COLUMN IF NOT EXISTS revoked BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMP NULL;

CREATE INDEX IF NOT EXISTS idx_agents_revoked ON agents(revoked);
