-- Agent System Tables
-- Note: agents table already exists from migration 003, only add new columns if needed
ALTER TABLE agents ADD COLUMN IF NOT EXISTS setup_method VARCHAR(20) DEFAULT 'automatic';
ALTER TABLE agents ADD COLUMN IF NOT EXISTS installed_at TIMESTAMP;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS cpu_usage DECIMAL(5,2) DEFAULT 0;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS memory_usage DECIMAL(5,2) DEFAULT 0;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS disk_usage DECIMAL(5,2) DEFAULT 0;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS services_count INTEGER DEFAULT 0;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS metadata JSONB;

CREATE TABLE IF NOT EXISTS agent_services (
    id SERIAL PRIMARY KEY,
    agent_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    service_name VARCHAR(255) NOT NULL,
    status VARCHAR(50),
    port INTEGER,
    process_id INTEGER,
    cpu_usage DECIMAL(5,2) DEFAULT 0,
    memory_usage BIGINT DEFAULT 0,
    uptime BIGINT DEFAULT 0,
    last_check TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(agent_id, service_name)
);

CREATE TABLE IF NOT EXISTS agent_installation_logs (
    id SERIAL PRIMARY KEY,
    agent_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    setup_method VARCHAR(20),
    status VARCHAR(50),
    logs TEXT,
    error_message TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Add agent_id to hosts table
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS agent_id INTEGER REFERENCES agents(id);
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS agent_status VARCHAR(50) DEFAULT 'offline';

-- Indexes (agents indexes already created in migration 003, only add new ones)
CREATE INDEX IF NOT EXISTS idx_agents_host_id ON agents(host_id);
-- idx_agents_token already exists from migration 003 (on 'token' column, not 'agent_token')
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agent_services_agent_id ON agent_services(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_installation_logs_agent_id ON agent_installation_logs(agent_id);