-- Add a target_os column to installation_tokens so tokens can carry an OS hint
ALTER TABLE installation_tokens ADD COLUMN IF NOT EXISTS target_os VARCHAR(32);
