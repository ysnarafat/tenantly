-- Reverse NID-at-rest protection. Note: encrypted values cannot be converted
-- back to plaintext here; any rows whose nid_number was nulled during back-fill
-- will lose their NID on rollback.
DROP INDEX IF EXISTS idx_tenants_nid_hash;

CREATE INDEX IF NOT EXISTS idx_tenants_nid ON tenants(nid_number) WHERE nid_number IS NOT NULL;

ALTER TABLE tenants
    DROP COLUMN IF EXISTS nid_encrypted,
    DROP COLUMN IF EXISTS nid_last_four,
    DROP COLUMN IF EXISTS nid_hash;
