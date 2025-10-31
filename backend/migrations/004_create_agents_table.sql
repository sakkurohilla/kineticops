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

-- Create services table for discovered services
CREATE TABLE IF NOT EXISTS host_services (
    id SERIAL PRIMARY KEY,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50),
    pid INTEGER,
    memory_usage BIGINT,
    cpu_usage FLOAT,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(host_id, name)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_agents_host_id ON agents(host_id);
CREATE INDEX IF NOT EXISTS idx_agents_token ON agents(token);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_host_services_host_id ON host_services(host_id);