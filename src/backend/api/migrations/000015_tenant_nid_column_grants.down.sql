-- Revoke this database's grants to the reporting role. Wrapped defensively to
-- mirror the up migration (skips cleanly where the migration user lacks
-- privileges).
--
-- The role itself is deliberately NOT dropped. Roles are cluster-wide while
-- grants are per-database, so DROP ROLE here reaches outside this migration's
-- own database: it fails outright whenever another database on the same server
-- still grants to the role, and it races with any other database migrating at
-- the same moment — either way leaving the migration dirty. After the revokes
-- below the role holds no privileges here and cannot log in, so leaving it in
-- place is inert. Operators removing it for good can DROP ROLE by hand once no
-- database grants to it.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'tenantly_reporting') THEN
        REVOKE ALL ON ALL TABLES IN SCHEMA public FROM tenantly_reporting;
        REVOKE ALL (
            id, name, tenant_type, phone_number, email,
            nid_last_four, address, active, organization_id, created_at, updated_at
        ) ON tenants FROM tenantly_reporting;
    END IF;
EXCEPTION
    WHEN insufficient_privilege THEN
        RAISE NOTICE 'Skipping tenantly_reporting revoke: migration user lacks privileges.';
    WHEN undefined_object THEN
        -- The role vanished between the EXISTS check and the REVOKEs, i.e. an
        -- operator dropped it by hand while this ran.
        RAISE NOTICE 'tenantly_reporting no longer exists; nothing to revoke.';
END $$;
