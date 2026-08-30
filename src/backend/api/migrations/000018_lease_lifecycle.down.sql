DROP INDEX IF EXISTS idx_leases_renewed_from_lease_id;
ALTER TABLE leases DROP COLUMN IF EXISTS renewed_from_lease_id;
ALTER TABLE leases DROP COLUMN IF EXISTS end_reason;
