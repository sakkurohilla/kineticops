-- Workflow System Tables
CREATE TABLE IF NOT EXISTS workflow_sessions (
    id SERIAL PRIMARY KEY,
    host_id INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'active',
    authenticated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS service_control_logs (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES agent_services(id),
    service_name VARCHAR(255),
    host_id INTEGER NOT NULL REFERENCES hosts(id),
    action VARCHAR(50) NOT NULL,
    status VARCHAR(50),
    output TEXT,
    error_message TEXT,
    executed_by INTEGER NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_workflow_sessions_host_user ON workflow_sessions(host_id, user_id);
CREATE INDEX IF NOT EXISTS idx_workflow_sessions_token ON workflow_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_workflow_sessions_expires ON workflow_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_service_control_logs_host ON service_control_logs(host_id);
CREATE INDEX IF NOT EXISTS idx_service_control_logs_executed_by ON service_control_logs(executed_by);