-- Drop indexes first (in reverse order)
DROP INDEX IF EXISTS idx_lease_attachments_lease_id;
DROP INDEX IF EXISTS idx_tenant_attachments_tenant_id;

-- Drop tables
DROP TABLE IF EXISTS lease_attachments CASCADE;
DROP TABLE IF EXISTS tenant_attachments CASCADE;
