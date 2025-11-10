-- Migration: compute disk_usage from disk_used/disk_total using a trigger
-- This ensures disk_usage is always canonical and derived from bytes stored

CREATE OR REPLACE FUNCTION compute_host_metrics_disk_usage()
RETURNS trigger AS $$
BEGIN
  IF NEW.disk_total IS NOT NULL AND NEW.disk_total <> 0 THEN
    NEW.disk_usage := (NEW.disk_used::double precision / NULLIF(NEW.disk_total,0)::double precision) * 100.0;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_compute_disk_usage
BEFORE INSERT OR UPDATE ON host_metrics
FOR EACH ROW
EXECUTE FUNCTION compute_host_metrics_disk_usage();
