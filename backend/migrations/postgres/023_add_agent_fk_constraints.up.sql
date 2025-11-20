-- Migration: Add foreign key constraints to agent_tokens and agent_health
-- Date: 2025-11-20
-- Note: This migration must run AFTER 010_agent_system.up.sql which creates the agents table

-- Add foreign key constraint to agent_tokens
ALTER TABLE agent_tokens
ADD CONSTRAINT fk_agent_tokens_agent
FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE;

-- Add foreign key constraint to agent_health  
ALTER TABLE agent_health
ADD CONSTRAINT fk_agent_health_agent
FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE;

COMMENT ON CONSTRAINT fk_agent_tokens_agent ON agent_tokens IS 'Foreign key to agents table';
COMMENT ON CONSTRAINT fk_agent_health_agent ON agent_health IS 'Foreign key to agents table';
