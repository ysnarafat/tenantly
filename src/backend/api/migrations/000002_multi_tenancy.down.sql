-- Remove test seed users and second org
DELETE FROM users WHERE username IN ('superadmin','orgadmin','adminuser','propmanager','accountant','acmeadmin');
DELETE FROM organizations WHERE slug = 'acme-properties';

-- Drop junction table
DROP TABLE IF EXISTS user_organization_roles;

-- Remove organization_id columns from data tables
ALTER TABLE payments DROP COLUMN IF EXISTS organization_id;
ALTER TABLE leases DROP COLUMN IF EXISTS organization_id;
ALTER TABLE tenants DROP COLUMN IF EXISTS organization_id;
ALTER TABLE units DROP COLUMN IF EXISTS organization_id;
ALTER TABLE buildings DROP COLUMN IF EXISTS organization_id;
ALTER TABLE properties DROP COLUMN IF EXISTS organization_id;

-- Remove new columns from users table
ALTER TABLE users DROP COLUMN IF EXISTS organization_id;
ALTER TABLE users DROP COLUMN IF EXISTS first_name;
ALTER TABLE users DROP COLUMN IF EXISTS last_name;
ALTER TABLE users DROP COLUMN IF EXISTS status;

-- Drop remaining new tables
DROP TABLE IF EXISTS user_invitations;
DROP TABLE IF EXISTS organizations;
