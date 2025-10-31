-- Create agents table for host monitoring agents
CREATE TABLE IF NOT EXISTS agents (
    id SERIAL PRIMARY KEY,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'pending',
    version VARCHAR(50),
    os_info JSONB,
    system_info JSONB,
    last_heartbeat TIMESTAMP,
    installation_log TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Update host_services table to match agent model
ALTER TABLE host_services DROP COLUMN IF EXISTS service_name;
ALTER TABLE host_services ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE host_services ADD COLUMN IF NOT EXISTS pid INTEGER DEFAULT 0;
ALTER TABLE host_services DROP COLUMN IF EXISTS port;
ALTER TABLE host_services ADD COLUMN IF NOT EXISTS discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE host_services ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Create unique constraint for host_services
DROP INDEX IF EXISTS idx_host_services_unique;
CREATE UNIQUE INDEX idx_host_services_unique ON host_services(host_id, name);

-- Create indexes for agents
CREATE INDEX IF NOT EXISTS idx_agents_host_id ON agents(host_id);
CREATE INDEX IF NOT EXISTS idx_agents_token ON agents(token);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_host_services_host_id_new ON host_services(host_id);

-- Update existing host_services data if any
UPDATE host_services SET name = COALESCE(service_name, 'unknown') WHERE name = '' OR name IS NULL;