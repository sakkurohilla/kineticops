-- Migration: Add audit log enhancements
-- Date: 2025-11-20

-- Update audit_logs table with new fields
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS tenant_id BIGINT DEFAULT 1;
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS username VARCHAR(255);
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS action VARCHAR(64);
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS resource VARCHAR(128);
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS resource_id VARCHAR(64);
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS ip_address VARCHAR(45);
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS user_agent VARCHAR(512);
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS status VARCHAR(16);
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS error_message VARCHAR(512);
ALTER TABLE IF EXISTS audit_logs ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_audit_tenant ON audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);

-- Agent tokens table for token rotation (no foreign key constraint for now - will be added after agents table exists)
CREATE TABLE IF NOT EXISTS agent_tokens (
    id BIGSERIAL PRIMARY KEY,
    agent_id BIGINT NOT NULL,
    token VARCHAR(128) UNIQUE NOT NULL,
    previous_token VARCHAR(128),
    version INT NOT NULL DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP,
    rotated_at TIMESTAMP,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_agent_tokens_agent ON agent_tokens(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_tokens_token ON agent_tokens(token);
CREATE INDEX IF NOT EXISTS idx_agent_tokens_active ON agent_tokens(is_active);
CREATE INDEX IF NOT EXISTS idx_agent_tokens_expires ON agent_tokens(expires_at);

-- Agent health table (no foreign key constraint for now - will be added after agents table exists)
CREATE TABLE IF NOT EXISTS agent_health (
    id BIGSERIAL PRIMARY KEY,
    agent_id BIGINT UNIQUE NOT NULL,
    health_score INT NOT NULL DEFAULT 100,
    status VARCHAR(32) NOT NULL DEFAULT 'healthy',
    last_heartbeat TIMESTAMP NOT NULL,
    heartbeat_missed INT DEFAULT 0,
    avg_latency FLOAT DEFAULT 0,
    error_rate FLOAT DEFAULT 0,
    data_quality FLOAT DEFAULT 1.0,
    metrics JSONB,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_agent_health_agent ON agent_health(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_health_status ON agent_health(status);

-- Agent versions table
CREATE TABLE IF NOT EXISTS agent_versions (
    version VARCHAR(32) PRIMARY KEY,
    release_date TIMESTAMP NOT NULL,
    is_latest BOOLEAN DEFAULT false,
    is_mandatory BOOLEAN DEFAULT false,
    min_compatible VARCHAR(32),
    changelog TEXT,
    download_url VARCHAR(512),
    checksum VARCHAR(128),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Custom metrics table with type support
CREATE TABLE IF NOT EXISTS custom_metrics (
    id BIGSERIAL PRIMARY KEY,
    host_id BIGINT NOT NULL,
    tenant_id BIGINT DEFAULT 1,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'gauge',
    value FLOAT NOT NULL,
    count BIGINT DEFAULT 0,
    sum FLOAT DEFAULT 0,
    min FLOAT,
    max FLOAT,
    p50 FLOAT,
    p95 FLOAT,
    p99 FLOAT,
    labels JSONB,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_custom_metrics_host ON custom_metrics(host_id);
CREATE INDEX IF NOT EXISTS idx_custom_metrics_name ON custom_metrics(name);
CREATE INDEX IF NOT EXISTS idx_custom_metrics_timestamp ON custom_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_custom_metrics_tenant ON custom_metrics(tenant_id);

-- Metric aggregations table
CREATE TABLE IF NOT EXISTS metric_aggregations (
    id BIGSERIAL PRIMARY KEY,
    host_id BIGINT NOT NULL,
    metric_name VARCHAR(128) NOT NULL,
    interval VARCHAR(16) NOT NULL,
    interval_time TIMESTAMP NOT NULL,
    avg FLOAT,
    min FLOAT,
    max FLOAT,
    sum FLOAT,
    count BIGINT,
    p50 FLOAT,
    p95 FLOAT,
    p99 FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(host_id, metric_name, interval, interval_time)
);

CREATE INDEX IF NOT EXISTS idx_agg_host_metric ON metric_aggregations(host_id, metric_name);
CREATE INDEX IF NOT EXISTS idx_agg_interval_time ON metric_aggregations(interval_time);

-- Metric trends table
CREATE TABLE IF NOT EXISTS metric_trends (
    id BIGSERIAL PRIMARY KEY,
    host_id BIGINT NOT NULL,
    metric_name VARCHAR(128) NOT NULL,
    trend_type VARCHAR(32) NOT NULL,
    confidence FLOAT,
    moving_avg FLOAT,
    std_dev FLOAT,
    slope FLOAT,
    is_anomaly BOOLEAN DEFAULT false,
    anomaly_score FLOAT DEFAULT 0,
    predicted_value FLOAT,
    analyzed_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_trends_host_metric ON metric_trends(host_id, metric_name);
CREATE INDEX IF NOT EXISTS idx_trends_analyzed ON metric_trends(analyzed_at);
CREATE INDEX IF NOT EXISTS idx_trends_anomaly ON metric_trends(is_anomaly) WHERE is_anomaly = true;

-- Add agent version column to agents table
ALTER TABLE IF EXISTS agents ADD COLUMN IF NOT EXISTS agent_version VARCHAR(32) DEFAULT '1.0.0';

COMMENT ON TABLE agent_tokens IS 'Agent authentication tokens with rotation support';
COMMENT ON TABLE agent_health IS 'Agent health monitoring and status tracking';
COMMENT ON TABLE agent_versions IS 'Available agent versions and release information';
COMMENT ON TABLE custom_metrics IS 'Extended metrics with type support (gauge, counter, histogram)';
COMMENT ON TABLE metric_aggregations IS 'Pre-aggregated metrics for faster queries';
COMMENT ON TABLE metric_trends IS 'Trend analysis and anomaly detection results';

-- Add foreign key constraints after agents table exists (migration 010 creates agents table)
-- These will be added in a later migration or manually after 010 runs
