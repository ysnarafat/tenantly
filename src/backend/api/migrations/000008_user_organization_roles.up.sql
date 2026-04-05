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

-- Backfill existing users from users.organization_id
INSERT INTO user_organization_roles (user_id, organization_id, role)
SELECT id, organization_id, role
FROM users
WHERE organization_id IS NOT NULL;

-- Indexes for performance
CREATE INDEX idx_user_org_roles_user_id ON user_organization_roles(user_id);
CREATE INDEX idx_user_org_roles_org_id ON user_organization_roles(organization_id);
