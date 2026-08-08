-- Re-add the plaintext column. Note: the original plaintext values are gone
-- (encrypted and dropped); the restored column is empty.
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS nid_number VARCHAR(20);
