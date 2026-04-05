-- Test seed users for development/testing
-- Password for all users: Test@1234
-- Organizations: org 1 = "Default Organization" (from migration 7)

-- Create a second test organization
INSERT INTO organizations (name, slug, subscription_tier)
VALUES ('Acme Properties', 'acme-properties', 'professional')
ON CONFLICT (slug) DO NOTHING;

-- SUPER_ADMIN: no org, global access
INSERT INTO users (username, email, password_hash, role, active, first_name, last_name, organization_id, status)
VALUES (
    'superadmin',
    'superadmin@test.com',
    '$2a$10$JsI.o99k0.p8chToHtaGGu7931Y5CCeHGmVkOkvJT.uhI44vWPqBW',
    'SUPER_ADMIN',
    true,
    'Super',
    'Admin',
    NULL,
    'active'
) ON CONFLICT (username) DO NOTHING;

-- ORG_ADMIN: manages Default Organization (id=1)
INSERT INTO users (username, email, password_hash, role, active, first_name, last_name, organization_id, status)
VALUES (
    'orgadmin',
    'orgadmin@test.com',
    '$2a$10$JsI.o99k0.p8chToHtaGGu7931Y5CCeHGmVkOkvJT.uhI44vWPqBW',
    'ORG_ADMIN',
    true,
    'Org',
    'Admin',
    1,
    'active'
) ON CONFLICT (username) DO NOTHING;

-- Admin: under Default Organization
INSERT INTO users (username, email, password_hash, role, active, first_name, last_name, organization_id, status)
VALUES (
    'adminuser',
    'admin@test.com',
    '$2a$10$JsI.o99k0.p8chToHtaGGu7931Y5CCeHGmVkOkvJT.uhI44vWPqBW',
    'Admin',
    true,
    'Admin',
    'User',
    1,
    'active'
) ON CONFLICT (username) DO NOTHING;

-- PropertyManager: under Default Organization
INSERT INTO users (username, email, password_hash, role, active, first_name, last_name, organization_id, status)
VALUES (
    'propmanager',
    'propmanager@test.com',
    '$2a$10$JsI.o99k0.p8chToHtaGGu7931Y5CCeHGmVkOkvJT.uhI44vWPqBW',
    'PropertyManager',
    true,
    'Property',
    'Manager',
    1,
    'active'
) ON CONFLICT (username) DO NOTHING;

-- Accountant: under Default Organization
INSERT INTO users (username, email, password_hash, role, active, first_name, last_name, organization_id, status)
VALUES (
    'accountant',
    'accountant@test.com',
    '$2a$10$JsI.o99k0.p8chToHtaGGu7931Y5CCeHGmVkOkvJT.uhI44vWPqBW',
    'Accountant',
    true,
    'Account',
    'Ant',
    1,
    'active'
) ON CONFLICT (username) DO NOTHING;

-- ORG_ADMIN for second org (Acme Properties)
INSERT INTO users (username, email, password_hash, role, active, first_name, last_name, organization_id, status)
SELECT
    'acmeadmin',
    'acmeadmin@test.com',
    '$2a$10$JsI.o99k0.p8chToHtaGGu7931Y5CCeHGmVkOkvJT.uhI44vWPqBW',
    'ORG_ADMIN',
    true,
    'Acme',
    'Admin',
    id,
    'active'
FROM organizations WHERE slug = 'acme-properties'
ON CONFLICT (username) DO NOTHING;
