-- Protect tenant NID at rest: store an AES-256-GCM ciphertext, a non-sensitive
-- last-four projection for display/search, and a deterministic keyed hash for
-- uniqueness lookups. The plaintext `nid_number` column is retained here only so
-- the application can back-fill the new columns on startup; it is nulled out
-- once encrypted and should be dropped in a follow-up migration after every
-- environment has completed the back-fill.
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS nid_encrypted TEXT,
    ADD COLUMN IF NOT EXISTS nid_last_four VARCHAR(4),
    ADD COLUMN IF NOT EXISTS nid_hash VARCHAR(64);

-- The plaintext index is no longer used; uniqueness/lookup now goes through the
-- deterministic hash column.
DROP INDEX IF EXISTS idx_tenants_nid;

CREATE INDEX IF NOT EXISTS idx_tenants_nid_hash ON tenants(nid_hash) WHERE nid_hash IS NOT NULL;
