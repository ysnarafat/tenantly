-- Remove the reporting role and its grants. Wrapped defensively to mirror the up
-- migration (skips cleanly where the migration user lacks privileges).
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'tenantly_reporting') THEN
        REVOKE ALL ON ALL TABLES IN SCHEMA public FROM tenantly_reporting;
        REVOKE ALL (
            id, name, tenant_type, phone_number, email,
            nid_last_four, address, active, organization_id, created_at, updated_at
        ) ON tenants FROM tenantly_reporting;
        DROP ROLE tenantly_reporting;
    END IF;
EXCEPTION
    WHEN insufficient_privilege THEN
        RAISE NOTICE 'Skipping tenantly_reporting role teardown: migration user lacks privileges.';
END $$;
