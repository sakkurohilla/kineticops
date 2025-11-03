-- Reset Database Script - Clear all data and reset sequences

-- Disable foreign key checks temporarily
SET session_replication_role = replica;

-- Clear all data from tables
TRUNCATE TABLE metrics RESTART IDENTITY CASCADE;
TRUNCATE TABLE host_metrics RESTART IDENTITY CASCADE;
TRUNCATE TABLE timeseries_metrics RESTART IDENTITY CASCADE;
TRUNCATE TABLE logs RESTART IDENTITY CASCADE;
TRUNCATE TABLE alerts RESTART IDENTITY CASCADE;
TRUNCATE TABLE workflows RESTART IDENTITY CASCADE;
TRUNCATE TABLE agents RESTART IDENTITY CASCADE;
TRUNCATE TABLE installation_tokens RESTART IDENTITY CASCADE;
TRUNCATE TABLE hosts RESTART IDENTITY CASCADE;

-- Reset sequences to start from 1
ALTER SEQUENCE hosts_id_seq RESTART WITH 1;
ALTER SEQUENCE agents_id_seq RESTART WITH 1;
ALTER SEQUENCE metrics_id_seq RESTART WITH 1;
ALTER SEQUENCE host_metrics_id_seq RESTART WITH 1;
ALTER SEQUENCE logs_id_seq RESTART WITH 1;
ALTER SEQUENCE alerts_id_seq RESTART WITH 1;
ALTER SEQUENCE workflows_id_seq RESTART WITH 1;
ALTER SEQUENCE installation_tokens_id_seq RESTART WITH 1;

-- Re-enable foreign key checks
SET session_replication_role = DEFAULT;

-- Verify sequences are reset
SELECT 
    schemaname,
    sequencename,
    last_value,
    start_value,
    increment_by
FROM pg_sequences 
WHERE schemaname = 'public'
ORDER BY sequencename;

VACUUM ANALYZE;

SELECT 'Database reset complete - all data cleared and sequences reset to 1' as status;