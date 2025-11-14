-- Make dropping the old constraint safe by using IF EXISTS
-- This avoids noisy errors during AutoMigrate when the constraint is missing.

ALTER TABLE host_metrics DROP CONSTRAINT IF EXISTS "uni_host_metrics_host_id";
