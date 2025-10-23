CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL,
    metric_name VARCHAR(128) NOT NULL,
    operator VARCHAR(8) NOT NULL,         -- e.g. '>', '<', '==', '!='
    threshold DOUBLE PRECISION NOT NULL,
    "window" INTEGER,                      -- minutes (time window to aggregate/trigger)
    frequency INTEGER DEFAULT 1,           -- how many times metric must breach rule within window
    notification_webhook TEXT,             -- webhook URL, nullable
    escalation_policy JSONB,               -- escalation steps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL,
    rule_id INT REFERENCES alert_rules(id) ON DELETE CASCADE,
    metric_name VARCHAR(128) NOT NULL,
    host_id INT,
    value DOUBLE PRECISION,
    status VARCHAR(32) NOT NULL,         -- OPEN, ACK, CLOSED, ESCALATED, etc.
    dedup_hash VARCHAR(128),             -- for deduplication
    escalated_level INT DEFAULT 0,
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP,
    notification_sent BOOLEAN DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts (status);
CREATE INDEX IF NOT EXISTS idx_alerts_rule ON alerts (rule_id);
CREATE INDEX IF NOT EXISTS idx_alert_rules_metric_tenant ON alert_rules (metric_name, tenant_id);
