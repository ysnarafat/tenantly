-- Establish least-privilege column access to the tenants table so that only the
-- application service account can read the encrypted NID material. A NOLOGIN
-- group role `tenantly_reporting` is created for analytics/BI/support users: it
-- can read everything EXCEPT the sensitive nid_encrypted and nid_hash columns.
-- Operators attach real users with: GRANT tenantly_reporting TO <bi_user>;
--
-- This is wrapped in an exception handler so environments whose migration user
-- lacks CREATEROLE do not fail startup — the control is simply skipped there and
-- must be applied out-of-band by a DBA.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'tenantly_reporting') THEN
        CREATE ROLE tenantly_reporting NOLOGIN;
    END IF;

    -- Full read on the rest of the schema for reporting utility...
    GRANT SELECT ON ALL TABLES IN SCHEMA public TO tenantly_reporting;

    -- ...but narrow tenants to column-level SELECT that omits the encrypted
    -- ciphertext and lookup hash. Table-level SELECT is revoked first so the
    -- column grant is the only path.
    REVOKE SELECT ON tenants FROM tenantly_reporting;
    GRANT SELECT (
        id, name, tenant_type, phone_number, email,
        nid_last_four, address, active, organization_id, created_at, updated_at
    ) ON tenants TO tenantly_reporting;
EXCEPTION
    WHEN insufficient_privilege THEN
        RAISE NOTICE 'Skipping tenantly_reporting role setup: migration user lacks role/grant privileges. Apply column grants out-of-band.';
END $$;
