DROP TABLE IF EXISTS hosts;

CREATE TABLE hosts (
    id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) UNIQUE NOT NULL,
    ip VARCHAR(64) NOT NULL,
    os VARCHAR(128),
    agent_status VARCHAR(32),
    tenant_id INTEGER NOT NULL,
    tags TEXT,
    "group" VARCHAR(128),
    last_seen TIMESTAMP,
    reg_token VARCHAR(128) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
