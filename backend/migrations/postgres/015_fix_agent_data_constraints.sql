-- Fix database constraints for agent data storage
-- This migration addresses the unique constraint issues with host creation and metric storage

-- Add unique constraint on host_metrics.host_id if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'host_metrics_host_id_key'
    ) THEN
        -- Remove duplicate entries first (keep the latest one)
        DELETE FROM host_metrics 
        WHERE id NOT IN (
            SELECT MAX(id) 
            FROM host_metrics 
            GROUP BY host_id
        );
        
        -- Add unique constraint
        ALTER TABLE host_metrics ADD CONSTRAINT host_metrics_host_id_key UNIQUE (host_id);
    END IF;
END $$;

-- Ensure reg_token is not null for existing hosts
UPDATE hosts SET reg_token = CONCAT('auto-', tenant_id, '-', hostname) 
WHERE reg_token IS NULL OR reg_token = '';

-- Add index for better performance on tenant-specific host lookups
CREATE INDEX IF NOT EXISTS idx_hosts_hostname_tenant ON hosts(hostname, tenant_id);

-- Clean up any orphaned host_metrics records
DELETE FROM host_metrics 
WHERE host_id NOT IN (SELECT id FROM hosts);

-- Clean up any orphaned metrics records
DELETE FROM metrics 
WHERE host_id NOT IN (SELECT id FROM hosts);