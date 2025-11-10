-- Down migration: drop trigger and function that compute disk_usage
DROP TRIGGER IF EXISTS tr_compute_disk_usage ON host_metrics;
DROP FUNCTION IF EXISTS compute_host_metrics_disk_usage();
