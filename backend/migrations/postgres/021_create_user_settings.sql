-- Create user_settings table
CREATE TABLE IF NOT EXISTS user_settings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    
    -- Account settings
    company_name VARCHAR(200),
    timezone VARCHAR(50) DEFAULT 'Asia/Kolkata',
    date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD',
    
    -- Notification settings
    email_notifications BOOLEAN DEFAULT true,
    slack_notifications BOOLEAN DEFAULT false,
    webhook_notifications BOOLEAN DEFAULT false,
    alert_email VARCHAR(255),
    slack_webhook TEXT,
    custom_webhook TEXT,
    
    -- Security settings
    require_mfa BOOLEAN DEFAULT false,
    session_timeout INTEGER DEFAULT 30,
    password_expiry INTEGER DEFAULT 90,
    
    -- Data retention settings
    metrics_retention INTEGER DEFAULT 30,
    logs_retention INTEGER DEFAULT 7,
    traces_retention INTEGER DEFAULT 7,
    
    -- Performance settings
    auto_refresh BOOLEAN DEFAULT true,
    refresh_interval INTEGER DEFAULT 30,
    max_dashboard_widgets INTEGER DEFAULT 20,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id
CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);
