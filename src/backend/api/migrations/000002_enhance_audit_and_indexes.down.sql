-- Drop triggers
DROP TRIGGER IF EXISTS update_notification_queue_updated_at ON notification_queue;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_leases_updated_at ON leases;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TRIGGER IF EXISTS update_shops_updated_at ON shops;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP TRIGGER IF EXISTS audit_payments_trigger ON payments;
DROP TRIGGER IF EXISTS audit_leases_trigger ON leases;
DROP TRIGGER IF EXISTS audit_tenants_trigger ON tenants;
DROP TRIGGER IF EXISTS audit_shops_trigger ON shops;
DROP TRIGGER IF EXISTS audit_users_trigger ON users;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS audit_trigger_function();

-- Drop indexes
DROP INDEX IF EXISTS idx_notification_queue_status_created;
DROP INDEX IF EXISTS idx_notification_queue_tenant_type;
DROP INDEX IF EXISTS idx_audit_log_record_id;
DROP INDEX IF EXISTS idx_audit_log_table_action;
DROP INDEX IF EXISTS idx_audit_log_user_id;
DROP INDEX IF EXISTS idx_payments_receipt_number;
DROP INDEX IF EXISTS idx_payments_year_month;
DROP INDEX IF EXISTS idx_payments_tenant_status;
DROP INDEX IF EXISTS idx_leases_shop_active;
DROP INDEX IF EXISTS idx_leases_tenant_active;
DROP INDEX IF EXISTS idx_leases_start_date;
DROP INDEX IF EXISTS idx_tenants_nid;
DROP INDEX IF EXISTS idx_tenants_email;
DROP INDEX IF EXISTS idx_tenants_phone;
DROP INDEX IF EXISTS idx_shops_floor_section;
DROP INDEX IF EXISTS idx_shops_property_active;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;