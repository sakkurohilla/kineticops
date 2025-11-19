-- Remove credentials from workflow sessions
ALTER TABLE workflow_sessions 
DROP COLUMN IF EXISTS username,
DROP COLUMN IF EXISTS password_encrypted,
DROP COLUMN IF EXISTS ssh_key_encrypted;
