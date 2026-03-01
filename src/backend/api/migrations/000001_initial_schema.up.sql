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
    building_name VARCHAR(100) NOT NULL,
    building_code VARCHAR(20) NOT NULL,
    building_type VARCHAR(50) NOT NULL CHECK (building_type IN ('Residential', 'Commercial', 'Mixed')),
    total_floors INTEGER,
    has_elevator BOOLEAN DEFAULT false,
    construction_year INTEGER,
    metadata JSONB, -- Building-specific attributes (parking_spaces, amenities, etc.)
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE(property_id, building_code)
);

-- Create units table (generalized rentable spaces - shops, apartments, offices, etc.)
CREATE TABLE units (
    id SERIAL PRIMARY KEY,
    building_id INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE, -- Denormalized for performance
    unit_number VARCHAR(50) NOT NULL,
    unit_name VARCHAR(100),
    floor INTEGER,
    section VARCHAR(50),
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('Shop', 'Apartment', 'Office', 'Parking', 'Storage', 'Other')),
    monthly_rent DECIMAL(10,2) NOT NULL,
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

-- Create indexes for better performance

-- Property indexes
CREATE INDEX idx_properties_active ON properties(active);
CREATE INDEX idx_properties_property_type ON properties(property_type);
CREATE INDEX idx_properties_metadata ON properties USING GIN(metadata);

-- Building indexes
CREATE INDEX idx_buildings_property_id ON buildings(property_id);
CREATE INDEX idx_buildings_active ON buildings(active);
CREATE INDEX idx_buildings_building_type ON buildings(building_type);
CREATE INDEX idx_buildings_metadata ON buildings USING GIN(metadata);

-- Unit indexes
CREATE INDEX idx_units_building_id ON units(building_id);
CREATE INDEX idx_units_property_id ON units(property_id);
CREATE INDEX idx_units_active ON units(active);
CREATE INDEX idx_units_unit_type ON units(unit_type);
CREATE INDEX idx_units_hierarchy ON units(property_id, building_id, unit_type);
CREATE INDEX idx_units_metadata ON units USING GIN(metadata);

-- Tenant indexes
CREATE INDEX idx_tenants_active ON tenants(active);
CREATE INDEX idx_tenants_phone ON tenants(phone_number);
CREATE INDEX idx_tenants_email ON tenants(email);

-- Lease indexes
CREATE INDEX idx_leases_unit_id ON leases(unit_id);
CREATE INDEX idx_leases_tenant_id ON leases(tenant_id);
CREATE INDEX idx_leases_active ON leases(active);
CREATE INDEX idx_leases_unit_tenant ON leases(unit_id, tenant_id);
CREATE INDEX idx_leases_dates ON leases(start_date, end_date);

-- Payment indexes
CREATE INDEX idx_payments_unit_id ON payments(unit_id);
CREATE INDEX idx_payments_tenant_id ON payments(tenant_id);
CREATE INDEX idx_payments_building_id ON payments(building_id);
CREATE INDEX idx_payments_property_id ON payments(property_id);
CREATE INDEX idx_payments_unit_month_year ON payments(unit_id, month, year);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_due_date ON payments(due_date);
CREATE INDEX idx_payments_hierarchy ON payments(property_id, building_id, unit_id);

-- Notification queue indexes
CREATE INDEX idx_notification_queue_status ON notification_queue(status);
CREATE INDEX idx_notification_queue_created_at ON notification_queue(created_at);
CREATE INDEX idx_notification_queue_tenant_id ON notification_queue(tenant_id);

-- Audit log indexes
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_table_name ON audit_log(table_name);

-- Insert default admin user (password: admin123)
INSERT INTO users (username, email, password_hash, role) 
VALUES ('admin', 'admin@tenantly.com', '$2a$10$V0LgWDqmvW4pcKylTvpwaePABYblvaOdqP7VB/c0ca8k5mL8g..9y', 'Admin');

-- Insert default property for initial setup
INSERT INTO properties (property_name, property_code, address, city, property_type) 
VALUES ('Default Property', 'PROP001', 'Default Address', 'Dhaka', 'Commercial');

-- Insert default building for the default property
INSERT INTO buildings (property_id, building_name, building_code, building_type, total_floors) 
VALUES (1, 'Main Building', 'MAIN', 'Commercial', 5);