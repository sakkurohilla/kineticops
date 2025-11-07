-- Create table to record recently deleted hosts (tombstones) to prevent immediate auto-recreation
CREATE TABLE IF NOT EXISTS deleted_hosts (
    id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL,
    tenant_id INTEGER NOT NULL,
    deleted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_deleted_hosts_hostname ON deleted_hosts(hostname);
