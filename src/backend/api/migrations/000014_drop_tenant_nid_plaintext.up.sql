-- Drop the drained plaintext nid_number column now that NID is stored encrypted.
-- Fail-safe: refuse to drop if any row still has un-encrypted plaintext (i.e. the
-- application-level back-fill has not yet run in this environment). This prevents
-- silent data loss if an environment upgrades straight to this migration before
-- booting on the build that performs the back-fill.
DO $$
DECLARE
    unmigrated INTEGER;
BEGIN
    SELECT COUNT(*) INTO unmigrated
    FROM tenants
    WHERE nid_number IS NOT NULL AND nid_number != '' AND nid_encrypted IS NULL;

    IF unmigrated > 0 THEN
        RAISE EXCEPTION 'Cannot drop nid_number: % row(s) still hold un-encrypted plaintext NID. Boot on the previous build first so the startup back-fill can encrypt them.', unmigrated;
    END IF;

    ALTER TABLE tenants DROP COLUMN IF EXISTS nid_number;
END $$;
