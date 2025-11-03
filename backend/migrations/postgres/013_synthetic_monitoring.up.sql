-- Synthetic Monitoring Tables

-- Synthetic Monitors table
CREATE TABLE IF NOT EXISTS synthetic_monitors (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(256) NOT NULL,
    type VARCHAR(32) NOT NULL, -- ping, simple_browser, scripted_browser, api_test
    status VARCHAR(32) DEFAULT 'enabled',
    frequency INTEGER DEFAULT 300, -- seconds
    locations TEXT, -- JSON array of locations
    config TEXT, -- JSON configuration
    script TEXT, -- For scripted monitors
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Synthetic Results table
CREATE TABLE IF NOT EXISTS synthetic_results (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    monitor_id BIGINT NOT NULL,
    location VARCHAR(64) DEFAULT 'default',
    success BOOLEAN DEFAULT FALSE,
    duration DOUBLE PRECISION DEFAULT 0, -- milliseconds
    status_code INTEGER,
    error VARCHAR(1024),
    screenshot VARCHAR(512), -- URL to screenshot
    metrics TEXT, -- JSON metrics
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (monitor_id) REFERENCES synthetic_monitors(id) ON DELETE CASCADE
);

-- Synthetic Alerts table
CREATE TABLE IF NOT EXISTS synthetic_alerts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    monitor_id BIGINT NOT NULL,
    type VARCHAR(64) NOT NULL, -- failure, slow_response, error_rate
    threshold DOUBLE PRECISION DEFAULT 0,
    duration INTEGER DEFAULT 5, -- minutes
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (monitor_id) REFERENCES synthetic_monitors(id) ON DELETE CASCADE
);

-- Browser Sessions table (for Browser Monitoring)
CREATE TABLE IF NOT EXISTS browser_sessions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT,
    session_id VARCHAR(64) UNIQUE NOT NULL,
    user_id VARCHAR(128),
    user_agent VARCHAR(512),
    browser VARCHAR(64),
    browser_version VARCHAR(32),
    os VARCHAR(64),
    device VARCHAR(64),
    country VARCHAR(64),
    region VARCHAR(64),
    city VARCHAR(64),
    start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE,
    duration DOUBLE PRECISION DEFAULT 0,
    page_views INTEGER DEFAULT 0,
    bounced BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Page Views table
CREATE TABLE IF NOT EXISTS page_views (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT,
    session_id VARCHAR(64),
    page_id VARCHAR(64),
    url VARCHAR(1024),
    title VARCHAR(512),
    referrer VARCHAR(1024),
    load_time DOUBLE PRECISION DEFAULT 0,
    dom_ready DOUBLE PRECISION DEFAULT 0,
    first_paint DOUBLE PRECISION DEFAULT 0,
    first_contentful_paint DOUBLE PRECISION DEFAULT 0,
    largest_contentful_paint DOUBLE PRECISION DEFAULT 0,
    cumulative_layout_shift DOUBLE PRECISION DEFAULT 0,
    first_input_delay DOUBLE PRECISION DEFAULT 0,
    time_to_interactive DOUBLE PRECISION DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- JavaScript Errors table
CREATE TABLE IF NOT EXISTS javascript_errors (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT,
    session_id VARCHAR(64),
    page_id VARCHAR(64),
    error_message VARCHAR(1024),
    error_type VARCHAR(128),
    stack_trace TEXT,
    file_name VARCHAR(512),
    line_number INTEGER,
    column_number INTEGER,
    user_agent VARCHAR(512),
    url VARCHAR(1024),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- AJAX Requests table
CREATE TABLE IF NOT EXISTS ajax_requests (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT,
    session_id VARCHAR(64),
    page_id VARCHAR(64),
    url VARCHAR(1024),
    method VARCHAR(16),
    status_code INTEGER,
    duration DOUBLE PRECISION DEFAULT 0,
    request_size BIGINT DEFAULT 0,
    response_size BIGINT DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_synthetic_monitors_tenant_id ON synthetic_monitors(tenant_id);
CREATE INDEX IF NOT EXISTS idx_synthetic_monitors_status ON synthetic_monitors(status);
CREATE INDEX IF NOT EXISTS idx_synthetic_monitors_type ON synthetic_monitors(type);

CREATE INDEX IF NOT EXISTS idx_synthetic_results_tenant_id ON synthetic_results(tenant_id);
CREATE INDEX IF NOT EXISTS idx_synthetic_results_monitor_id ON synthetic_results(monitor_id);
CREATE INDEX IF NOT EXISTS idx_synthetic_results_timestamp ON synthetic_results(timestamp);
CREATE INDEX IF NOT EXISTS idx_synthetic_results_success ON synthetic_results(success);

CREATE INDEX IF NOT EXISTS idx_synthetic_alerts_tenant_id ON synthetic_alerts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_synthetic_alerts_monitor_id ON synthetic_alerts(monitor_id);
CREATE INDEX IF NOT EXISTS idx_synthetic_alerts_enabled ON synthetic_alerts(enabled);

CREATE INDEX IF NOT EXISTS idx_browser_sessions_tenant_id ON browser_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_application_id ON browser_sessions(application_id);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_session_id ON browser_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_browser_sessions_start_time ON browser_sessions(start_time);

CREATE INDEX IF NOT EXISTS idx_page_views_tenant_id ON page_views(tenant_id);
CREATE INDEX IF NOT EXISTS idx_page_views_application_id ON page_views(application_id);
CREATE INDEX IF NOT EXISTS idx_page_views_session_id ON page_views(session_id);
CREATE INDEX IF NOT EXISTS idx_page_views_timestamp ON page_views(timestamp);

CREATE INDEX IF NOT EXISTS idx_javascript_errors_tenant_id ON javascript_errors(tenant_id);
CREATE INDEX IF NOT EXISTS idx_javascript_errors_application_id ON javascript_errors(application_id);
CREATE INDEX IF NOT EXISTS idx_javascript_errors_session_id ON javascript_errors(session_id);
CREATE INDEX IF NOT EXISTS idx_javascript_errors_timestamp ON javascript_errors(timestamp);

CREATE INDEX IF NOT EXISTS idx_ajax_requests_tenant_id ON ajax_requests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ajax_requests_application_id ON ajax_requests(application_id);
CREATE INDEX IF NOT EXISTS idx_ajax_requests_session_id ON ajax_requests(session_id);
CREATE INDEX IF NOT EXISTS idx_ajax_requests_timestamp ON ajax_requests(timestamp);