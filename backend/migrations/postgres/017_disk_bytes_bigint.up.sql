-- Migration: store disk totals/used as BIGINT (bytes)
-- This migration converts existing disk_total and disk_used DECIMAL values
-- into BIGINT by rounding. It preserves existing values but gives us
-- capacity to store raw byte counts from agents.

BEGIN;

ALTER TABLE host_metrics
  ALTER COLUMN disk_total TYPE BIGINT USING ROUND(COALESCE(disk_total,0))::BIGINT;

ALTER TABLE host_metrics
  ALTER COLUMN disk_used TYPE BIGINT USING ROUND(COALESCE(disk_used,0))::BIGINT;

COMMIT;
