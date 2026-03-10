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

-- Create indexes for performance
CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_active ON organizations(active);
CREATE INDEX idx_user_invitations_org ON user_invitations(organization_id);
CREATE INDEX idx_user_invitations_token ON user_invitations(invitation_token);
CREATE INDEX idx_users_org_id ON users(organization_id);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_properties_org_id ON properties(organization_id);
CREATE INDEX idx_buildings_org_id ON buildings(organization_id);
CREATE INDEX idx_units_org_id ON units(organization_id);
CREATE INDEX idx_tenants_org_id ON tenants(organization_id);
CREATE INDEX idx_leases_org_id ON leases(organization_id);
CREATE INDEX idx_payments_org_id ON payments(organization_id);
