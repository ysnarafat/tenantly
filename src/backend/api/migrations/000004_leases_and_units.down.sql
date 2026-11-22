-- Drop indexes and columns
DROP INDEX IF EXISTS idx_leases_renewed_from_lease_id;

ALTER TABLE leases DROP COLUMN IF EXISTS renewed_from_lease_id;
ALTER TABLE leases DROP COLUMN IF EXISTS end_reason;

DROP INDEX IF EXISTS idx_lease_charges_lease_id;
ALTER TABLE leases DROP COLUMN IF EXISTS custom_fields;

DROP TABLE IF EXISTS lease_charges;

DROP INDEX IF EXISTS idx_leases_unit_active_unique;

DROP TABLE IF EXISTS lease_attachments CASCADE;
DROP TABLE IF EXISTS tenant_attachments CASCADE;