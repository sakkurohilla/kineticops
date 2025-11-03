-- APM (Application Performance Monitoring) Tables

-- Applications table
CREATE TABLE IF NOT EXISTS applications (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    host_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(64) DEFAULT 'web',
    language VARCHAR(32),
    framework VARCHAR(64),
    version VARCHAR(32),
    status VARCHAR(32) DEFAULT 'unknown',
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT NOT NULL,
    trace_id VARCHAR(64),
    name VARCHAR(256) NOT NULL,
    type VARCHAR(64) DEFAULT 'web',
    duration DOUBLE PRECISION DEFAULT 0,
    response_time DOUBLE PRECISION DEFAULT 0,
    throughput DOUBLE PRECISION DEFAULT 0,
    error_rate DOUBLE PRECISION DEFAULT 0,
    apdex DOUBLE PRECISION DEFAULT 0,
    status_code INTEGER,
    method VARCHAR(16),
    uri VARCHAR(512),
    user_agent VARCHAR(512),
    remote_ip VARCHAR(45),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Traces table
CREATE TABLE IF NOT EXISTS traces (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    trace_id VARCHAR(64) UNIQUE NOT NULL,
    application_id BIGINT NOT NULL,
    root_span_id VARCHAR(64),
    duration DOUBLE PRECISION DEFAULT 0,
    span_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Spans table
CREATE TABLE IF NOT EXISTS spans (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    trace_id VARCHAR(64) NOT NULL,
    span_id VARCHAR(64) UNIQUE NOT NULL,
    parent_span_id VARCHAR(64),
    application_id BIGINT NOT NULL,
    operation_name VARCHAR(256),
    service_name VARCHAR(128),
    start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    duration DOUBLE PRECISION DEFAULT 0,
    tags TEXT,
    logs TEXT,
    status VARCHAR(32) DEFAULT 'ok',
    error BOOLEAN DEFAULT FALSE,
    error_message VARCHAR(1024),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Error Events table
CREATE TABLE IF NOT EXISTS error_events (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT NOT NULL,
    trace_id VARCHAR(64),
    span_id VARCHAR(64),
    error_class VARCHAR(256),
    error_message VARCHAR(1024),
    stack_trace TEXT,
    fingerprint VARCHAR(64),
    count INTEGER DEFAULT 1,
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Database Queries table
CREATE TABLE IF NOT EXISTS database_queries (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT NOT NULL,
    trace_id VARCHAR(64),
    span_id VARCHAR(64),
    database VARCHAR(128),
    operation VARCHAR(64),
    table_name VARCHAR(128),
    query TEXT,
    duration DOUBLE PRECISION DEFAULT 0,
    rows_affected BIGINT DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- External Services table
CREATE TABLE IF NOT EXISTS external_services (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT NOT NULL,
    trace_id VARCHAR(64),
    span_id VARCHAR(64),
    service_name VARCHAR(128),
    url VARCHAR(512),
    method VARCHAR(16),
    status_code INTEGER,
    duration DOUBLE PRECISION DEFAULT 0,
    response_size BIGINT DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Performance Metrics table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT NOT NULL,
    metric_name VARCHAR(128) NOT NULL,
    metric_type VARCHAR(32) DEFAULT 'gauge',
    value DOUBLE PRECISION DEFAULT 0,
    tags TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Custom Events table
CREATE TABLE IF NOT EXISTS custom_events (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    application_id BIGINT NOT NULL,
    event_type VARCHAR(128) NOT NULL,
    event_name VARCHAR(256),
    attributes TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_applications_tenant_id ON applications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_applications_host_id ON applications(host_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);

CREATE INDEX IF NOT EXISTS idx_transactions_tenant_id ON transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_transactions_application_id ON transactions(application_id);
CREATE INDEX IF NOT EXISTS idx_transactions_trace_id ON transactions(trace_id);
CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp);
CREATE INDEX IF NOT EXISTS idx_transactions_response_time ON transactions(response_time);

CREATE INDEX IF NOT EXISTS idx_traces_tenant_id ON traces(tenant_id);
CREATE INDEX IF NOT EXISTS idx_traces_application_id ON traces(application_id);
CREATE INDEX IF NOT EXISTS idx_traces_trace_id ON traces(trace_id);
CREATE INDEX IF NOT EXISTS idx_traces_timestamp ON traces(timestamp);

CREATE INDEX IF NOT EXISTS idx_spans_tenant_id ON spans(tenant_id);
CREATE INDEX IF NOT EXISTS idx_spans_trace_id ON spans(trace_id);
CREATE INDEX IF NOT EXISTS idx_spans_span_id ON spans(span_id);
CREATE INDEX IF NOT EXISTS idx_spans_parent_span_id ON spans(parent_span_id);
CREATE INDEX IF NOT EXISTS idx_spans_application_id ON spans(application_id);

CREATE INDEX IF NOT EXISTS idx_error_events_tenant_id ON error_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_error_events_application_id ON error_events(application_id);
CREATE INDEX IF NOT EXISTS idx_error_events_fingerprint ON error_events(fingerprint);
CREATE INDEX IF NOT EXISTS idx_error_events_timestamp ON error_events(timestamp);

CREATE INDEX IF NOT EXISTS idx_database_queries_tenant_id ON database_queries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_database_queries_application_id ON database_queries(application_id);
CREATE INDEX IF NOT EXISTS idx_database_queries_duration ON database_queries(duration);
CREATE INDEX IF NOT EXISTS idx_database_queries_timestamp ON database_queries(timestamp);

CREATE INDEX IF NOT EXISTS idx_external_services_tenant_id ON external_services(tenant_id);
CREATE INDEX IF NOT EXISTS idx_external_services_application_id ON external_services(application_id);
CREATE INDEX IF NOT EXISTS idx_external_services_timestamp ON external_services(timestamp);

CREATE INDEX IF NOT EXISTS idx_performance_metrics_tenant_id ON performance_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_application_id ON performance_metrics(application_id);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_name ON performance_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_timestamp ON performance_metrics(timestamp);

CREATE INDEX IF NOT EXISTS idx_custom_events_tenant_id ON custom_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_custom_events_application_id ON custom_events(application_id);
CREATE INDEX IF NOT EXISTS idx_custom_events_type ON custom_events(event_type);
CREATE INDEX IF NOT EXISTS idx_custom_events_timestamp ON custom_events(timestamp);