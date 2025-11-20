-- Enhanced Alert Management System Migration
-- This creates production-ready alert tables

-- Drop existing basic alert tables if they exist
DROP TABLE IF EXISTS alerts CASCADE;
DROP TABLE IF EXISTS alert_rules CASCADE;

-- Create enhanced alert_rules table
CREATE TABLE alert_rules (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric_name VARCHAR(128) NOT NULL,
    
    -- Condition Configuration
    operator VARCHAR(32) NOT NULL DEFAULT '>',
    threshold DECIMAL(15,4) NOT NULL,
    secondary_threshold DECIMAL(15,4), -- For range checks
    
    -- Evaluation Configuration
    evaluation_window INTEGER NOT NULL DEFAULT 300, -- seconds
    evaluation_interval INTEGER NOT NULL DEFAULT 60, -- seconds
    consecutive_breaches INTEGER NOT NULL DEFAULT 1,
    
    -- Severity and Status
    severity VARCHAR(32) NOT NULL DEFAULT 'MEDIUM',
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    
    -- Filtering
    host_filters TEXT, -- JSON array of host IDs
    tag_filters TEXT,  -- JSON object for tag-based filtering
    
    -- Notification Configuration
    notification_channels TEXT, -- JSON array of channel configs
    notification_template TEXT, -- Custom message template
    
    -- Escalation Configuration
    escalation_policy TEXT, -- JSON escalation rules
    max_escalation_level INTEGER DEFAULT 3,
    escalation_interval INTEGER DEFAULT 1800, -- seconds
    
    -- Suppression Configuration
    suppression_rules TEXT, -- JSON suppression conditions
    dependency_rules TEXT,  -- JSON dependency relationships
    
    -- Aggregation Configuration
    group_by_labels TEXT,   -- JSON array of labels for grouping
    aggregation_window INTEGER DEFAULT 300,
    max_aggregated_alerts INTEGER DEFAULT 10,
    
    -- Recovery Configuration
    auto_resolve BOOLEAN DEFAULT TRUE,
    resolve_timeout INTEGER DEFAULT 900, -- seconds
    
    -- Maintenance
    maintenance_windows TEXT, -- JSON maintenance schedule
    
    -- Metadata
    tags TEXT,
    labels TEXT,
    runbook_url VARCHAR(512),
    dashboard_url VARCHAR(512),
    
    -- Audit
    created_by INTEGER NOT NULL,
    updated_by INTEGER,
    last_triggered TIMESTAMP,
    trigger_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create enhanced alerts table
CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    rule_id INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    metric_name VARCHAR(128) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(32) NOT NULL DEFAULT 'MEDIUM',
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    value DECIMAL(15,4),
    threshold DECIMAL(15,4),
    dedup_hash VARCHAR(64),
    escalation_level INTEGER DEFAULT 0,
    last_escalated_at TIMESTAMP,
    silenced_until TIMESTAMP,
    silenced_by INTEGER,
    acknowledged_at TIMESTAMP,
    acknowledged_by INTEGER,
    triggered_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    closed_at TIMESTAMP,
    notifications_sent INTEGER DEFAULT 0,
    last_notified_at TIMESTAMP,
    tags TEXT,     -- JSON array of tags
    labels TEXT,   -- JSON object of key-value pairs
    runbook_url VARCHAR(512),
    dashboard_url VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create alert history table for audit trail
CREATE TABLE alert_history (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    tenant_id INTEGER NOT NULL,
    user_id INTEGER, -- NULL for system actions
    action VARCHAR(64) NOT NULL, -- TRIGGERED, ACKNOWLEDGED, SILENCED, RESOLVED, ESCALATED
    old_status VARCHAR(32),
    new_status VARCHAR(32),
    comment TEXT,
    metadata TEXT, -- JSON for additional context
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create alert comments table
CREATE TABLE alert_comments (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    tenant_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create alert assignments table
CREATE TABLE alert_assignments (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    tenant_id INTEGER NOT NULL,
    assigned_to_id INTEGER NOT NULL,
    assigned_by_id INTEGER NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Create alert aggregations table
CREATE TABLE alert_aggregations (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    rule_id INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    group_key VARCHAR(128), -- Hash of grouping criteria
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    alert_count INTEGER DEFAULT 1,
    first_alert_at TIMESTAMP NOT NULL,
    last_alert_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_alert_rules_tenant_id ON alert_rules(tenant_id);
CREATE INDEX idx_alert_rules_metric_name ON alert_rules(metric_name);
CREATE INDEX idx_alert_rules_status ON alert_rules(status);
CREATE INDEX idx_alert_rules_created_by ON alert_rules(created_by);
CREATE INDEX idx_alert_rules_last_triggered ON alert_rules(last_triggered);

CREATE INDEX idx_alerts_tenant_id ON alerts(tenant_id);
CREATE INDEX idx_alerts_rule_id ON alerts(rule_id);
CREATE INDEX idx_alerts_host_id ON alerts(host_id);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_triggered_at ON alerts(triggered_at);
CREATE INDEX idx_alerts_dedup_hash ON alerts(dedup_hash);
CREATE INDEX idx_alerts_silenced_until ON alerts(silenced_until);

CREATE INDEX idx_alert_history_alert_id ON alert_history(alert_id);
CREATE INDEX idx_alert_history_tenant_id ON alert_history(tenant_id);
CREATE INDEX idx_alert_history_user_id ON alert_history(user_id);
CREATE INDEX idx_alert_history_action ON alert_history(action);
CREATE INDEX idx_alert_history_created_at ON alert_history(created_at);

CREATE INDEX idx_alert_comments_alert_id ON alert_comments(alert_id);
CREATE INDEX idx_alert_comments_tenant_id ON alert_comments(tenant_id);
CREATE INDEX idx_alert_comments_user_id ON alert_comments(user_id);

CREATE INDEX idx_alert_assignments_alert_id ON alert_assignments(alert_id);
CREATE INDEX idx_alert_assignments_tenant_id ON alert_assignments(tenant_id);
CREATE INDEX idx_alert_assignments_assigned_to_id ON alert_assignments(assigned_to_id);

CREATE INDEX idx_alert_aggregations_tenant_id ON alert_aggregations(tenant_id);
CREATE INDEX idx_alert_aggregations_rule_id ON alert_aggregations(rule_id);
CREATE INDEX idx_alert_aggregations_group_key ON alert_aggregations(group_key);

-- Add constraints
ALTER TABLE alert_rules ADD CONSTRAINT chk_alert_rules_severity 
    CHECK (severity IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO'));
    
ALTER TABLE alert_rules ADD CONSTRAINT chk_alert_rules_status 
    CHECK (status IN ('ACTIVE', 'INACTIVE', 'PAUSED'));
    
ALTER TABLE alert_rules ADD CONSTRAINT chk_alert_rules_operator 
    CHECK (operator IN ('>', '>=', '<', '<=', '==', '!=', 'contains', 'not_contains'));

ALTER TABLE alerts ADD CONSTRAINT chk_alerts_severity 
    CHECK (severity IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO'));
    
ALTER TABLE alerts ADD CONSTRAINT chk_alerts_status 
    CHECK (status IN ('OPEN', 'ACKNOWLEDGED', 'SILENCED', 'RESOLVED', 'CLOSED'));

-- Create triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_alerts_updated_at BEFORE UPDATE ON alerts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_alert_comments_updated_at BEFORE UPDATE ON alert_comments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_alert_aggregations_updated_at BEFORE UPDATE ON alert_aggregations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();