-- Remove disk I/O metrics from host_metrics table
ALTER TABLE host_metrics
DROP COLUMN IF EXISTS disk_read_bytes,
DROP COLUMN IF EXISTS disk_write_bytes,
DROP COLUMN IF EXISTS disk_read_speed,
DROP COLUMN IF EXISTS disk_write_speed;
