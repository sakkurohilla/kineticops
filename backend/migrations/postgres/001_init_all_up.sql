-- Complete Database Migration for KineticOps
-- Clean version - NO sample data

-- Drop tables in reverse order to handle foreign key constraints
DROP TABLE IF EXISTS alerts CASCADE;
DROP TABLE IF EXISTS host_alerts CASCADE;
DROP TABLE IF EXISTS host_services CASCADE;
DROP TABLE IF EXISTS host_metrics CASCADE;
DROP TABLE IF EXISTS hosts CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(128) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    oauth_provider VARCHAR(32),
    oauth_id VARCHAR(128),
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create audit_logs table
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    event VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    details TEXT
);

-- Create hosts table - CRITICAL: No unique constraint on hostname alone
CREATE TABLE hosts (
    id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL,
    ip VARCHAR(64) NOT NULL,
    os VARCHAR(128) DEFAULT 'linux',
    agent_status VARCHAR(32) DEFAULT 'offline',
    status VARCHAR(32) DEFAULT 'inactive',
    tenant_id INTEGER NOT NULL DEFAULT 1,
    tags TEXT,
    "group" VARCHAR(128) DEFAULT 'default',
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_sync TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reg_token VARCHAR(128) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- SSH Configuration
    ssh_user VARCHAR(255) DEFAULT 'root',
    ssh_password VARCHAR(255) DEFAULT '',
    ssh_port INTEGER DEFAULT 22,
    ssh_key TEXT,
    
    -- Agent Configuration
    agent_token VARCHAR(128),
    
    -- Additional fields
    description TEXT
);

-- Add this after the hosts table creation and before host_metrics
CREATE TABLE metrics (
    id SERIAL PRIMARY KEY,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    tenant_id INTEGER NOT NULL,
    name VARCHAR(128) NOT NULL,
    value DECIMAL(10,4) NOT NULL,
    labels TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_metrics_host_id ON metrics(host_id);
CREATE INDEX idx_metrics_timestamp ON metrics(timestamp);
CREATE INDEX idx_metrics_name ON metrics(name);


-- Create host_metrics table
CREATE TABLE host_metrics (
    id SERIAL PRIMARY KEY,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    cpu_usage DECIMAL(5,2) DEFAULT 0,
    memory_usage DECIMAL(5,2) DEFAULT 0,
    memory_total DECIMAL(10,2) DEFAULT 0,
    memory_used DECIMAL(10,2) DEFAULT 0,
    disk_usage DECIMAL(5,2) DEFAULT 0,
    disk_total DECIMAL(10,2) DEFAULT 0,
    disk_used DECIMAL(10,2) DEFAULT 0,
    network_in DECIMAL(10,2) DEFAULT 0,
    network_out DECIMAL(10,2) DEFAULT 0,
    uptime BIGINT DEFAULT 0,
    load_average VARCHAR(64) DEFAULT '',
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create host_services table
CREATE TABLE host_services (
    id SERIAL PRIMARY KEY,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    service_name VARCHAR(255) NOT NULL,
    status VARCHAR(32) DEFAULT 'stopped',
    port INTEGER,
    memory_usage DECIMAL(5,2) DEFAULT 0,
    cpu_usage DECIMAL(5,2) DEFAULT 0,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create host_alerts table
CREATE TABLE host_alerts (
    id SERIAL PRIMARY KEY,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    alert_type VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    message TEXT NOT NULL,
    is_resolved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- Create alerts table
CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    host_id INTEGER REFERENCES hosts(id) ON DELETE CASCADE,
    alert_type VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    message TEXT NOT NULL,
    is_resolved BOOLEAN DEFAULT FALSE,
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_hosts_tenant_id ON hosts(tenant_id);
CREATE INDEX idx_hosts_agent_status ON hosts(agent_status);
CREATE INDEX idx_hosts_group ON hosts("group");
CREATE INDEX idx_hosts_hostname ON hosts(hostname);

-- CRITICAL: Allow same hostname per tenant, not globally unique
CREATE UNIQUE INDEX idx_hosts_hostname_tenant_unique ON hosts(hostname, tenant_id);

CREATE INDEX idx_host_metrics_host_id ON host_metrics(host_id);
CREATE UNIQUE INDEX idx_host_metrics_host_id_unique ON host_metrics(host_id);
CREATE INDEX idx_host_metrics_timestamp ON host_metrics(timestamp);
CREATE INDEX idx_host_services_host_id ON host_services(host_id);
CREATE INDEX idx_host_alerts_host_id ON host_alerts(host_id);
CREATE INDEX idx_host_alerts_resolved ON host_alerts(is_resolved);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_alerts_tenant_id ON alerts(tenant_id);
CREATE INDEX idx_alerts_host_id ON alerts(host_id);
CREATE INDEX idx_alerts_triggered_at ON alerts(triggered_at);