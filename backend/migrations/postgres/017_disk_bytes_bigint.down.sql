-- Down migration: revert disk_total/disk_used back to DECIMAL(10,2)
BEGIN;

ALTER TABLE host_metrics
  ALTER COLUMN disk_total TYPE DECIMAL(10,2) USING disk_total::numeric;

ALTER TABLE host_metrics
  ALTER COLUMN disk_used TYPE DECIMAL(10,2) USING disk_used::numeric;

COMMIT;
