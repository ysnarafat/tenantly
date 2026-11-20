CREATE TABLE billing_runs (
    id SERIAL PRIMARY KEY,
    property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    month INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year INTEGER NOT NULL CHECK (year >= 2020),
    status VARCHAR(20) NOT NULL DEFAULT 'Draft' CHECK (status IN ('Draft', 'Finalized')),
    created_by INTEGER REFERENCES users(id),
    finalized_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE (property_id, month, year)
);

CREATE TABLE billing_run_items (
    id SERIAL PRIMARY KEY,
    billing_run_id INTEGER NOT NULL REFERENCES billing_runs(id) ON DELETE CASCADE,
    lease_id INTEGER NOT NULL REFERENCES leases(id),
    unit_id INTEGER NOT NULL REFERENCES units(id),
    tenant_id INTEGER NOT NULL REFERENCES tenants(id),
    building_id INTEGER NOT NULL REFERENCES buildings(id),
    monthly_rent DECIMAL(10,2) NOT NULL CHECK (monthly_rent >= 0),
    service_charge DECIMAL(10,2) NOT NULL CHECK (service_charge >= 0),
    electricity DECIMAL(10,2) NOT NULL DEFAULT 0 CHECK (electricity >= 0),
    payment_id INTEGER UNIQUE REFERENCES payments(id),
    UNIQUE (billing_run_id, unit_id)
);

CREATE INDEX idx_billing_runs_org_property ON billing_runs (organization_id, property_id);
CREATE INDEX idx_billing_run_items_run ON billing_run_items (billing_run_id);
