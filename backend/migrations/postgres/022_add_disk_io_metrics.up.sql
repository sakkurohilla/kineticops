-- Add disk I/O metrics to host_metrics table
ALTER TABLE host_metrics
ADD COLUMN IF NOT EXISTS disk_read_bytes DECIMAL(15,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS disk_write_bytes DECIMAL(15,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS disk_read_speed DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS disk_write_speed DECIMAL(10,2) DEFAULT 0;
