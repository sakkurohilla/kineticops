CREATE TABLE IF NOT EXISTS metrics (
    id SERIAL PRIMARY KEY,
    host_id INTEGER REFERENCES hosts(id),
    tenant_id INTEGER NOT NULL,
    name VARCHAR(128) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    labels TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_metrics_host_time ON metrics (host_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_tenant_name_time ON metrics (tenant_id, name, timestamp DESC);
