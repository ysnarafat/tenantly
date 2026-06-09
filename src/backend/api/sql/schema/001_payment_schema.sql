-- Consolidated schema for sqlc type checking
-- This represents the final DDL state after all migrations

-- Organizations (from migration 000007)
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

-- Properties (from migration 000001 + 000007)
CREATE TABLE properties (
    id SERIAL PRIMARY KEY,
    property_name VARCHAR(200) NOT NULL,
    property_code VARCHAR(50) UNIQUE NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100),
    postal_code VARCHAR(20),
    property_type VARCHAR(50) NOT NULL CHECK (property_type IN ('Residential', 'Commercial', 'Mixed')),
    total_buildings INTEGER DEFAULT 0,
    metadata JSONB,
    active BOOLEAN DEFAULT true,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Building type enum (from migration 000005)
CREATE TYPE building_type_enum AS ENUM ('Residential', 'Commercial', 'Mixed');

-- Buildings (from migration 000001 + 000005 + 000007)
CREATE TABLE buildings (
    id SERIAL PRIMARY KEY,
    property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    building_name VARCHAR(100) NOT NULL,
    building_code VARCHAR(20) NOT NULL,
    building_type building_type_enum NOT NULL,
    total_floors INTEGER,
    has_elevator BOOLEAN DEFAULT false,
    construction_year INTEGER,
    metadata JSONB,
    active_status BOOLEAN DEFAULT true,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE(property_id, building_code)
);

-- Units (from migration 000001 + 000007)
CREATE TABLE units (
    id SERIAL PRIMARY KEY,
    building_id INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    unit_number VARCHAR(50) NOT NULL,
    unit_name VARCHAR(100),
    floor INTEGER,
    section VARCHAR(50),
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('Shop', 'Apartment', 'Office', 'Parking', 'Storage', 'Other')),
    monthly_rent DECIMAL(10,2) NOT NULL,
    metadata JSONB,
    active BOOLEAN DEFAULT true,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE(building_id, unit_number)
);

-- Tenants (from migration 000001 + 000007)
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    tenant_type VARCHAR(20) NOT NULL CHECK (tenant_type IN ('Individual', 'Business')),
    phone_number VARCHAR(20),
    email VARCHAR(100),
    nid_number VARCHAR(20),
    address TEXT,
    active BOOLEAN DEFAULT true,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Leases (from migration 000001 + 000007)
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
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Payments (from migration 000001 + 000007)
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    unit_id INTEGER NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    building_id INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
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
