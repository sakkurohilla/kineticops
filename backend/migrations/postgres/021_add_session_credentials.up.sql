-- Add credentials to workflow sessions for real server authentication
-- These are encrypted and stored only for the duration of the session (1 hour max)
ALTER TABLE workflow_sessions 
ADD COLUMN IF NOT EXISTS username VARCHAR(255),
ADD COLUMN IF NOT EXISTS password_encrypted TEXT,
ADD COLUMN IF NOT EXISTS ssh_key_encrypted TEXT;

-- Add comment to clarify security
COMMENT ON COLUMN workflow_sessions.password_encrypted IS 'Encrypted password for SSH authentication, stored only for session duration';
COMMENT ON COLUMN workflow_sessions.ssh_key_encrypted IS 'Encrypted SSH private key (or PEM file) for authentication, stored only for session duration';
