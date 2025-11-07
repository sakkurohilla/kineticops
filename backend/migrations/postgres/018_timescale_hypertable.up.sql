-- Enable TimescaleDB extension and convert the metrics table to a hypertable
-- This migration creates continuous aggregates for common rollups

-- Create extension if not present
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Ensure metrics table exists (created in earlier migrations)
-- Convert metrics.timestamp into a hypertable
SELECT create_hypertable('metrics', 'timestamp', if_not_exists => TRUE);

-- Create continuous aggregate for 5 minute rollups
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_5m WITH (timescaledb.continuous) AS
SELECT
  time_bucket('5 minutes', timestamp) AS bucket,
  host_id,
  name,
  AVG(value) AS value
FROM metrics
GROUP BY bucket, host_id, name;

-- Create continuous aggregate for 1 hour rollups
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1h WITH (timescaledb.continuous) AS
SELECT
  time_bucket('1 hour', timestamp) AS bucket,
  host_id,
  name,
  AVG(value) AS value
FROM metrics
GROUP BY bucket, host_id, name;

-- Create continuous aggregate for 1 day rollups
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1d WITH (timescaledb.continuous) AS
SELECT
  time_bucket('1 day', timestamp) AS bucket,
  host_id,
  name,
  AVG(value) AS value
FROM metrics
GROUP BY bucket, host_id, name;

-- Add refresh policies so Timescale maintains aggregates automatically
SELECT add_continuous_aggregate_policy('metrics_5m', start_offset => INTERVAL '1 hour', end_offset => INTERVAL '1 minute', schedule_interval => INTERVAL '1 minute');
SELECT add_continuous_aggregate_policy('metrics_1h', start_offset => INTERVAL '1 day', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '15 minutes');
SELECT add_continuous_aggregate_policy('metrics_1d', start_offset => INTERVAL '7 days', end_offset => INTERVAL '1 day', schedule_interval => INTERVAL '1 hour');

-- Optional: create a retention policy that drops chunks older than 90 days
-- (Administrators can adjust or remove as needed)
SELECT add_retention_policy('metrics', INTERVAL '90 days');
