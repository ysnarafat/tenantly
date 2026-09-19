-- Building type enum
CREATE TYPE building_type_enum AS ENUM ('Residential', 'Commercial', 'Mixed');

-- Create users table with extensible role system
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- Extensible: Admin, PropertyManager, Accountant, or future roles
    active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Create properties table (top-level real estate assets)
CREATE TABLE properties (
    id SERIAL PRIMARY KEY,
    property_name VARCHAR(200) NOT NULL,
    property_code VARCHAR(50) UNIQUE NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100),
    postal_code VARCHAR(20),
    property_type VARCHAR(50) NOT NULL CHECK (property_type IN ('Residential', 'Commercial', 'Mixed')),
    total_buildings INTEGER DEFAULT 0,
    metadata JSONB, -- Property-specific attributes
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Create buildings table (physical structures within properties)
CREATE TABLE buildings (
    id SERIAL PRIMARY KEY,
    property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    building_name VARCHAR(100) NOT NULL CONSTRAINT chk_building_name_not_empty CHECK (LENGTH(TRIM(building_name)) > 0),
    building_code VARCHAR(20) NOT NULL CONSTRAINT chk_building_code_not_empty CHECK (LENGTH(TRIM(building_code)) > 0),
    building_type building_type_enum NOT NULL,
    total_floors INTEGER CONSTRAINT chk_total_floors_positive CHECK (total_floors > 0),
    has_elevator BOOLEAN DEFAULT false,
    construction_year INTEGER CONSTRAINT chk_construction_year_valid
        CHECK (construction_year IS NULL OR (construction_year >= 1800 AND construction_year <= EXTRACT(YEAR FROM CURRENT_DATE) + 5)),
    metadata JSONB, -- Building-specific attributes (parking_spaces, amenities, etc.)
    active_status BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE(property_id, building_code)
);

COMMENT ON TABLE buildings IS 'Building management table supporting residential, commercial, and mixed-use buildings with comprehensive metadata and indexing';
COMMENT ON COLUMN buildings.building_type IS 'Building classification using enum: Residential, Commercial, or Mixed';
COMMENT ON COLUMN buildings.metadata IS 'JSONB field for building-specific attributes like amenities, parking, security systems';
COMMENT ON COLUMN buildings.active_status IS 'Boolean flag indicating if the building is active in the system';

-- Create units table (generalized rentable spaces - shops, apartments, offices, etc.)
CREATE TABLE units (
    id SERIAL PRIMARY KEY,
    building_id INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE, -- Denormalized for performance
    unit_number VARCHAR(50) NOT NULL CONSTRAINT chk_unit_number_not_empty CHECK (LENGTH(TRIM(unit_number)) > 0),
    unit_name VARCHAR(100),
    floor INTEGER CONSTRAINT chk_floor_valid CHECK (floor IS NULL OR floor >= 0),
    section VARCHAR(50),
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('Shop', 'Apartment', 'Office', 'Parking', 'Storage', 'Other')),
    metadata JSONB, -- Type-specific attributes (area_sqft, bedrooms, etc.)
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE(building_id, unit_number)
);

-- Create tenants table
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    tenant_type VARCHAR(20) NOT NULL CHECK (tenant_type IN ('Individual', 'Business')),
    phone_number VARCHAR(20),
    email VARCHAR(100),
    nid_number VARCHAR(20),
    address TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Create leases table (links tenants to units)
CREATE TABLE leases (
    id SERIAL PRIMARY KEY,
    unit_id INTEGER NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    lease_type VARCHAR(20) NOT NULL CHECK (lease_type IN ('Residential', 'Commercial')),
    start_date DATE NOT NULL,
    end_date DATE,
    duration_months INTEGER NOT NULL,
    monthly_rent DECIMAL(10,2) NOT NULL,
    security_deposit DECIMAL(10,2),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Create payments table
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    unit_id INTEGER NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    building_id INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE, -- Denormalized for reporting
    property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE, -- Denormalized for reporting
    month INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year INTEGER NOT NULL,
    amount_due DECIMAL(10,2) NOT NULL,
    amount_paid DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'Due' CHECK (status IN ('Paid', 'Due', 'Partial', 'Overdue')),
    payment_method VARCHAR(50),
    payment_date DATE,
    notes TEXT,
    receipt_number VARCHAR(50) UNIQUE,
    due_date DATE,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE(unit_id, month, year)
);

-- Create notification_queue table
CREATE TABLE notification_queue (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    unit_id INTEGER NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    property_id INTEGER REFERENCES properties(id) ON DELETE CASCADE,
    building_id INTEGER REFERENCES buildings(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    notification_type VARCHAR(20) NOT NULL CHECK (notification_type IN ('SMS', 'Email', 'Reminder', 'Acknowledgment')),
    recipient VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'Pending' CHECK (status IN ('Pending', 'Sent', 'Failed')),
    retry_count INTEGER DEFAULT 0,
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Create audit_log table
CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    table_name VARCHAR(50) NOT NULL,
    record_id INTEGER,
    old_values JSONB,
    new_values JSONB,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Create password_reset_tokens table
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Indexes

-- User indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(active);

-- Property indexes
CREATE INDEX idx_properties_active ON properties(active);
CREATE INDEX idx_properties_property_type ON properties(property_type);
CREATE INDEX idx_properties_metadata ON properties USING GIN(metadata);
CREATE INDEX idx_properties_code ON properties(property_code);
CREATE INDEX idx_properties_code_active ON properties(property_code, active);
CREATE INDEX idx_properties_type_active ON properties(property_type, active);
CREATE INDEX idx_properties_active_created ON properties(active, created_at DESC);
CREATE UNIQUE INDEX idx_properties_code_unique_active ON properties (LOWER(property_code)) WHERE active = true;

-- Building indexes
CREATE INDEX idx_buildings_property_id ON buildings(property_id);
CREATE INDEX idx_buildings_building_type ON buildings(building_type);
CREATE INDEX idx_buildings_code ON buildings(property_id, building_code);
CREATE INDEX idx_buildings_property_id_enhanced ON buildings(property_id) WHERE active_status = true;
CREATE INDEX idx_buildings_type_enhanced ON buildings(building_type) WHERE active_status = true;
CREATE INDEX idx_buildings_active_status ON buildings(active_status);
CREATE INDEX idx_buildings_property_type_active ON buildings(property_id, building_type, active_status);
CREATE INDEX idx_buildings_metadata_gin ON buildings USING GIN (metadata) WHERE metadata IS NOT NULL;
CREATE INDEX idx_buildings_floors ON buildings(total_floors) WHERE active_status = true;
CREATE INDEX idx_buildings_elevator ON buildings(has_elevator) WHERE active_status = true;
CREATE INDEX idx_buildings_construction_year ON buildings(construction_year) WHERE active_status = true AND construction_year IS NOT NULL;
CREATE INDEX idx_buildings_active_status_created ON buildings(active_status, created_at DESC);
CREATE INDEX idx_buildings_property_active_status ON buildings(property_id, active_status);

COMMENT ON INDEX idx_buildings_property_id_enhanced IS 'Index for property_id queries filtered to active buildings';
COMMENT ON INDEX idx_buildings_type_enhanced IS 'Index for building_type queries filtered to active buildings';

-- Unit indexes
CREATE INDEX idx_units_building_id ON units(building_id);
CREATE INDEX idx_units_property_id ON units(property_id);
CREATE INDEX idx_units_active ON units(active);
CREATE INDEX idx_units_unit_type ON units(unit_type);
CREATE INDEX idx_units_metadata ON units USING GIN(metadata);
CREATE INDEX idx_units_floor_section ON units(floor, section) WHERE floor IS NOT NULL AND section IS NOT NULL;
CREATE INDEX idx_units_property_active ON units(property_id, active);
CREATE INDEX idx_units_building_active_status ON units(building_id);
CREATE INDEX idx_units_hierarchy_enhanced ON units(property_id, building_id, floor, section) WHERE active = true;
CREATE INDEX idx_units_building_floor_section ON units(building_id, floor, section) WHERE active = true;
CREATE INDEX idx_units_property_building_type ON units(property_id, building_id, unit_type) WHERE active = true;
CREATE INDEX idx_units_hierarchy_analytics ON units(property_id, building_id, unit_type, active);

COMMENT ON INDEX idx_units_hierarchy_enhanced IS 'Hierarchical index for property -> building -> floor -> section queries';
COMMENT ON INDEX idx_units_building_floor_section IS 'Index for building-specific unit queries with floor and section ordering';
COMMENT ON INDEX idx_units_property_building_type IS 'Index for property-level unit queries across buildings by type';
COMMENT ON INDEX idx_units_hierarchy_analytics IS 'Index optimized for hierarchical reporting and analytics queries';

-- Tenant indexes
CREATE INDEX idx_tenants_active ON tenants(active);
CREATE INDEX idx_tenants_phone ON tenants(phone_number);
CREATE INDEX idx_tenants_email ON tenants(email);
CREATE INDEX idx_tenants_nid ON tenants(nid_number) WHERE nid_number IS NOT NULL;

-- Lease indexes
CREATE INDEX idx_leases_unit_id ON leases(unit_id);
CREATE INDEX idx_leases_tenant_id ON leases(tenant_id);
CREATE INDEX idx_leases_active ON leases(active);
CREATE INDEX idx_leases_unit_tenant ON leases(unit_id, tenant_id);
CREATE INDEX idx_leases_dates ON leases(start_date, end_date);
CREATE INDEX idx_leases_start_date ON leases(start_date);
CREATE INDEX idx_leases_tenant_active ON leases(tenant_id, active);
CREATE INDEX idx_leases_unit_active ON leases(unit_id, active);
CREATE INDEX idx_leases_end_date ON leases(end_date);

-- Payment indexes
CREATE INDEX idx_payments_unit_id ON payments(unit_id);
CREATE INDEX idx_payments_tenant_id ON payments(tenant_id);
CREATE INDEX idx_payments_building_id ON payments(building_id);
CREATE INDEX idx_payments_property_id ON payments(property_id);
CREATE INDEX idx_payments_unit_month_year ON payments(unit_id, month, year);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_due_date ON payments(due_date);
CREATE INDEX idx_payments_hierarchy ON payments(property_id, building_id, unit_id);
CREATE INDEX idx_payments_tenant_status ON payments(tenant_id, status);
CREATE INDEX idx_payments_year_month ON payments(year, month);
CREATE INDEX idx_payments_receipt_number ON payments(receipt_number) WHERE receipt_number IS NOT NULL;
CREATE INDEX idx_payments_property_status ON payments(property_id, status);

-- Notification queue indexes
CREATE INDEX idx_notification_queue_status ON notification_queue(status);
CREATE INDEX idx_notification_queue_created_at ON notification_queue(created_at);
CREATE INDEX idx_notification_queue_tenant_id ON notification_queue(tenant_id);
CREATE INDEX idx_notification_queue_tenant_type ON notification_queue(tenant_id, notification_type);
CREATE INDEX idx_notification_queue_status_created ON notification_queue(status, created_at);

-- Audit log indexes
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_table_name ON audit_log(table_name);
CREATE INDEX idx_audit_log_table_action ON audit_log(table_name, action);
CREATE INDEX idx_audit_log_record_id ON audit_log(record_id) WHERE record_id IS NOT NULL;

-- Password reset token indexes
CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX idx_password_reset_tokens_used ON password_reset_tokens(used);
CREATE INDEX idx_password_reset_tokens_token_valid ON password_reset_tokens(token, used, expires_at);

-- Audit trigger function: records INSERT/UPDATE/DELETE on tracked tables into audit_log
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

-- updated_at trigger function: keeps updated_at current on every row update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW() AT TIME ZONE 'UTC';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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

-- Insert default admin user (password: admin123)
INSERT INTO users (username, email, password_hash, role)
VALUES ('admin', 'admin@tenantly.com', '$2a$10$V0LgWDqmvW4pcKylTvpwaePABYblvaOdqP7VB/c0ca8k5mL8g..9y', 'Admin');

-- Insert default property for initial setup
INSERT INTO properties (property_name, property_code, address, city, property_type)
VALUES ('Default Property', 'PROP001', 'Default Address', 'Dhaka', 'Commercial');

-- Insert default building for the default property
INSERT INTO buildings (property_id, building_name, building_code, building_type, total_floors)
VALUES (1, 'Main Building', 'MAIN', 'Commercial', 5);
