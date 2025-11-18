-- Migration: Enable TimescaleDB and convert metrics to hypertable
-- This migration:
-- 1. Backs up existing metrics data
-- 2. Recreates metrics table without conflicting primary key
-- 3. Converts to TimescaleDB hypertable
-- 4. Restores data
-- 5. Creates continuous aggregates
-- 6. Adds optimized indexes

-- Create backup table
CREATE TABLE IF NOT EXISTS metrics_backup AS SELECT * FROM metrics;

-- Drop existing metrics table
DROP TABLE IF EXISTS metrics CASCADE;

-- Recreate metrics table with TimescaleDB-compatible structure
CREATE TABLE metrics (
    id BIGSERIAL,
    host_id BIGINT,
    tenant_id BIGINT,
    name VARCHAR(128),
    value NUMERIC,
    timestamp TIMESTAMPTZ NOT NULL,
    labels TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('metrics', 'timestamp', if_not_exists => TRUE);

-- Restore data from backup
INSERT INTO metrics (id, host_id, tenant_id, name, value, timestamp, labels, created_at)
SELECT id, host_id, tenant_id, name, value, timestamp, labels, created_at 
FROM metrics_backup
ON CONFLICT DO NOTHING;

-- Update sequence to continue from max id
SELECT setval('metrics_id_seq', (SELECT COALESCE(MAX(id), 1) FROM metrics), true);

-- Create optimized indexes
CREATE INDEX IF NOT EXISTS idx_metrics_host_id ON metrics (host_id);
CREATE INDEX IF NOT EXISTS idx_metrics_tenant_id ON metrics (tenant_id);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_host_name_time ON metrics (host_id, name, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics (name);

-- Create continuous aggregate for 5-minute intervals
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_5m
WITH (timescaledb.continuous) AS
SELECT 
    host_id,
    tenant_id,
    name,
    time_bucket('5 minutes', timestamp) AS bucket,
    AVG(value) as avg_value,
    MAX(value) as max_value,
    MIN(value) as min_value,
    COUNT(*) as count
FROM metrics
GROUP BY host_id, tenant_id, name, bucket
WITH NO DATA;

-- Create continuous aggregate for 1-hour intervals
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1h
WITH (timescaledb.continuous) AS
SELECT 
    host_id,
    tenant_id,
    name,
    time_bucket('1 hour', timestamp) AS bucket,
    AVG(value) as avg_value,
    MAX(value) as max_value,
    MIN(value) as min_value,
    COUNT(*) as count
FROM metrics
GROUP BY host_id, tenant_id, name, bucket
WITH NO DATA;

-- Create continuous aggregate for 1-day intervals
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1d
WITH (timescaledb.continuous) AS
SELECT 
    host_id,
    tenant_id,
    name,
    time_bucket('1 day', timestamp) AS bucket,
    AVG(value) as avg_value,
    MAX(value) as max_value,
    MIN(value) as min_value,
    COUNT(*) as count
FROM metrics
GROUP BY host_id, tenant_id, name, bucket
WITH NO DATA;

-- Add refresh policies for continuous aggregates
SELECT add_continuous_aggregate_policy('metrics_5m', 
    start_offset => INTERVAL '1 day', 
    end_offset => INTERVAL '5 minutes', 
    schedule_interval => INTERVAL '5 minutes',
    if_not_exists => TRUE);

SELECT add_continuous_aggregate_policy('metrics_1h', 
    start_offset => INTERVAL '7 days', 
    end_offset => INTERVAL '1 hour', 
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

SELECT add_continuous_aggregate_policy('metrics_1d', 
    start_offset => INTERVAL '30 days', 
    end_offset => INTERVAL '1 day', 
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE);

-- Add data retention policy (90 days for raw metrics)
SELECT add_retention_policy('metrics', INTERVAL '90 days', if_not_exists => TRUE);

-- Refresh continuous aggregates with existing data
CALL refresh_continuous_aggregate('metrics_5m', NULL, NULL);
CALL refresh_continuous_aggregate('metrics_1h', NULL, NULL);
CALL refresh_continuous_aggregate('metrics_1d', NULL, NULL);

-- Drop backup table
DROP TABLE IF EXISTS metrics_backup;
