-- Enable TimescaleDB extension and convert the metrics table to a hypertable
-- This migration creates continuous aggregates for common rollups

-- Create extension if not present
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Ensure metrics table exists (created in earlier migrations)
-- Convert metrics.timestamp into a hypertable
SELECT create_hypertable('metrics', 'timestamp', if_not_exists => TRUE);

-- Note: Continuous aggregates and policies should be created outside transaction
-- Run these manually or in a separate migration script if needed:
-- CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_5m WITH (timescaledb.continuous) AS ...
-- SELECT add_continuous_aggregate_policy(...);
-- SELECT add_retention_policy('metrics', INTERVAL '90 days');

