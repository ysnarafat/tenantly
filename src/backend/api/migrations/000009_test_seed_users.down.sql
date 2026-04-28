DELETE FROM users WHERE username IN ('superadmin','orgadmin','adminuser','propmanager','accountant','acmeadmin');
DELETE FROM organizations WHERE slug = 'acme-properties';
