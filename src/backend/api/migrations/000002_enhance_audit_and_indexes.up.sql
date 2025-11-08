-- Add additional indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);

-- Add composite indexes for properties and buildings
CREATE INDEX IF NOT EXISTS idx_properties_code ON properties(property_code);
CREATE INDEX IF NOT EXISTS idx_buildings_property_active ON buildings(property_id, active);
CREATE INDEX IF NOT EXISTS idx_buildings_code ON buildings(property_id, building_code);

-- Add composite indexes for units
CREATE INDEX IF NOT EXISTS idx_units_building_active ON units(building_id, active);
CREATE INDEX IF NOT EXISTS idx_units_floor_section ON units(floor, section) WHERE floor IS NOT NULL AND section IS NOT NULL;

-- Add indexes for tenant searches
CREATE INDEX IF NOT EXISTS idx_tenants_phone ON tenants(phone_number) WHERE phone_number IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_email ON tenants(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_nid ON tenants(nid_number) WHERE nid_number IS NOT NULL;

-- Add lease-specific indexes
CREATE INDEX IF NOT EXISTS idx_leases_start_date ON leases(start_date);
CREATE INDEX IF NOT EXISTS idx_leases_tenant_active ON leases(tenant_id, active);
CREATE INDEX IF NOT EXISTS idx_leases_unit_active ON leases(unit_id, active);
CREATE INDEX IF NOT EXISTS idx_leases_end_date ON leases(end_date);

-- Add payment performance indexes
CREATE INDEX IF NOT EXISTS idx_payments_tenant_status ON payments(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_payments_year_month ON payments(year, month);
CREATE INDEX IF NOT EXISTS idx_payments_receipt_number ON payments(receipt_number) WHERE receipt_number IS NOT NULL;

-- Add audit log performance indexes
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_log_table_action ON audit_log(table_name, action);
CREATE INDEX IF NOT EXISTS idx_audit_log_record_id ON audit_log(record_id) WHERE record_id IS NOT NULL;

-- Add notification queue performance indexes
CREATE INDEX IF NOT EXISTS idx_notification_queue_tenant_type ON notification_queue(tenant_id, notification_type);
CREATE INDEX IF NOT EXISTS idx_notification_queue_status_created ON notification_queue(status, created_at);

-- Add triggers for automatic audit logging
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (action, table_name, record_id, old_values, created_at)
        VALUES ('DELETE', TG_TABLE_NAME, OLD.id, row_to_json(OLD), NOW() AT TIME ZONE 'UTC');
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (action, table_name, record_id, old_values, new_values, created_at)
        VALUES ('UPDATE', TG_TABLE_NAME, NEW.id, row_to_json(OLD), row_to_json(NEW), NOW() AT TIME ZONE 'UTC');
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (action, table_name, record_id, new_values, created_at)
        VALUES ('INSERT', TG_TABLE_NAME, NEW.id, row_to_json(NEW), NOW() AT TIME ZONE 'UTC');
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create audit triggers for all main tables
CREATE TRIGGER audit_users_trigger
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_properties_trigger
    AFTER INSERT OR UPDATE OR DELETE ON properties
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_buildings_trigger
    AFTER INSERT OR UPDATE OR DELETE ON buildings
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_units_trigger
    AFTER INSERT OR UPDATE OR DELETE ON units
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_tenants_trigger
    AFTER INSERT OR UPDATE OR DELETE ON tenants
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_leases_trigger
    AFTER INSERT OR UPDATE OR DELETE ON leases
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_payments_trigger
    AFTER INSERT OR UPDATE OR DELETE ON payments
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Add updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW() AT TIME ZONE 'UTC';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create updated_at triggers for all tables
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_properties_updated_at
    BEFORE UPDATE ON properties
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_buildings_updated_at
    BEFORE UPDATE ON buildings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_units_updated_at
    BEFORE UPDATE ON units
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_leases_updated_at
    BEFORE UPDATE ON leases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_queue_updated_at
    BEFORE UPDATE ON notification_queue
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();