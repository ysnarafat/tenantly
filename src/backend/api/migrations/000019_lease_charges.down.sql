ALTER TABLE leases DROP COLUMN IF EXISTS custom_fields;

DROP INDEX IF EXISTS idx_lease_charges_lease_id;
DROP TABLE IF EXISTS lease_charges;
