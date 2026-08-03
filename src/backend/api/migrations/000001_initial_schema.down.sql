-- Drop tables in reverse dependency order to handle foreign key constraints
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS notification_queue;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS leases;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS buildings;
DROP TABLE IF EXISTS properties;
DROP TABLE IF EXISTS users;

-- Drop functions (triggers are dropped automatically with their tables)
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS audit_trigger_function();

-- Drop types
DROP TYPE IF EXISTS building_type_enum;
