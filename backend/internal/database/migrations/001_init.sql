-- Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Hosts Table
CREATE TABLE IF NOT EXISTS hosts (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    ip_address VARCHAR(50) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'offline',
    created_at TIMESTAMP NOT NULL
);
