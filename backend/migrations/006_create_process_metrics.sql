-- Create process_metrics table for storing per-process metrics
CREATE TABLE IF NOT EXISTS process_metrics (
    id BIGSERIAL,
    host_id BIGINT NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    pid INTEGER NOT NULL,
    name TEXT NOT NULL,
    username TEXT,
    cpu_percent NUMERIC(5,2),
    memory_percent NUMERIC(5,2),
    memory_rss BIGINT, -- bytes
    status TEXT,
    num_threads INTEGER,
    create_time BIGINT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_process_metrics_host_id ON process_metrics(host_id);
CREATE INDEX IF NOT EXISTS idx_process_metrics_timestamp ON process_metrics(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_process_metrics_tenant_id ON process_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_process_metrics_host_timestamp ON process_metrics(host_id, timestamp DESC);

-- Create hypertable for time-series data (if TimescaleDB is enabled)
SELECT create_hypertable('process_metrics', 'timestamp', if_not_exists => TRUE);

-- Add retention policy: keep process metrics for 7 days
SELECT add_retention_policy('process_metrics', INTERVAL '7 days', if_not_exists => TRUE);

-- Create continuous aggregate for hourly process stats
CREATE MATERIALIZED VIEW IF NOT EXISTS process_metrics_1h
WITH (timescaledb.continuous) AS
SELECT 
    host_id,
    tenant_id,
    name,
    time_bucket('1 hour', timestamp) AS bucket,
    AVG(cpu_percent) as avg_cpu,
    MAX(cpu_percent) as max_cpu,
    AVG(memory_percent) as avg_memory,
    MAX(memory_percent) as max_memory,
    AVG(memory_rss) as avg_memory_rss
FROM process_metrics
GROUP BY host_id, tenant_id, name, bucket;

-- Add refresh policy for continuous aggregate
SELECT add_continuous_aggregate_policy('process_metrics_1h',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);
