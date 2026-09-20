-- Create organizations table (top-level container for multi-tenancy)
CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    subscription_tier VARCHAR(20) NOT NULL DEFAULT 'basic'
        CHECK (subscription_tier IN ('basic', 'professional', 'enterprise')),
    max_users INTEGER DEFAULT 50,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Create user_invitations table (for user onboarding flow)
CREATE TABLE user_invitations (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL,
    invitation_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    accepted_by_user_id INTEGER REFERENCES users(id),
    invited_by_user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Create user_organization_roles junction table for multi-organization support
-- This allows a user to belong to multiple organizations with different roles
CREATE TABLE user_organization_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL
        CHECK (role IN ('SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'PropertyManager', 'Accountant')),
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE(user_id, organization_id)
);

-- Insert default organization for existing data
INSERT INTO organizations (name, slug, subscription_tier)
VALUES ('Default Organization', 'default', 'basic');

-- Alter users table to add organization context and profile fields
ALTER TABLE users
    ADD COLUMN organization_id INTEGER REFERENCES organizations(id),
    ADD COLUMN first_name VARCHAR(100),
    ADD COLUMN last_name VARCHAR(100),
    ADD COLUMN status VARCHAR(20) DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'pending_invite'));

-- Assign existing users to default organization
UPDATE users SET
    organization_id = 1,
    first_name = 'System',
    last_name = 'Admin',
    status = 'active';

-- Add organization_id to all data tables (all scoped to organizations)
ALTER TABLE properties ADD COLUMN organization_id INTEGER REFERENCES organizations(id);
UPDATE properties SET organization_id = 1;
ALTER TABLE properties ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE buildings ADD COLUMN organization_id INTEGER REFERENCES organizations(id);
UPDATE buildings SET organization_id = 1;
ALTER TABLE buildings ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE units ADD COLUMN organization_id INTEGER REFERENCES organizations(id);
UPDATE units SET organization_id = 1;
ALTER TABLE units ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE tenants ADD COLUMN organization_id INTEGER REFERENCES organizations(id);
UPDATE tenants SET organization_id = 1;
ALTER TABLE tenants ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE leases ADD COLUMN organization_id INTEGER REFERENCES organizations(id);
UPDATE leases SET organization_id = 1;
ALTER TABLE leases ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE payments ADD COLUMN organization_id INTEGER REFERENCES organizations(id);
UPDATE payments SET organization_id = 1;
ALTER TABLE payments ALTER COLUMN organization_id SET NOT NULL;

-- Backfill user_organization_roles from users.organization_id
INSERT INTO user_organization_roles (user_id, organization_id, role)
SELECT id, organization_id, role
FROM users
WHERE organization_id IS NOT NULL;

-- Create indexes for performance
CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_active ON organizations(active);
CREATE INDEX idx_user_invitations_org ON user_invitations(organization_id);
CREATE INDEX idx_user_invitations_token ON user_invitations(invitation_token);
CREATE INDEX idx_user_org_roles_user_id ON user_organization_roles(user_id);
CREATE INDEX idx_user_org_roles_org_id ON user_organization_roles(organization_id);
CREATE INDEX idx_users_org_id ON users(organization_id);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_properties_org_id ON properties(organization_id);
CREATE INDEX idx_buildings_org_id ON buildings(organization_id);
CREATE INDEX idx_units_org_id ON units(organization_id);
CREATE INDEX idx_tenants_org_id ON tenants(organization_id);
CREATE INDEX idx_leases_org_id ON leases(organization_id);
CREATE INDEX idx_payments_org_id ON payments(organization_id);

-- Test seed users for development/testing
-- Password for all users: Test@1234
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
