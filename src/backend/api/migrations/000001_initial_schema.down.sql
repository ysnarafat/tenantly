-- Drop tables in reverse order to handle foreign key constraints
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS notification_queue;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS leases;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS buildings;
DROP TABLE IF EXISTS properties;
DROP TABLE IF EXISTS users;