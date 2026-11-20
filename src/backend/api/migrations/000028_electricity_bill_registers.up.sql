CREATE TABLE electricity_bill_runs (
    id SERIAL PRIMARY KEY,
    building_id INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    month INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year INTEGER NOT NULL CHECK (year >= 2020),
    status VARCHAR(20) NOT NULL DEFAULT 'Draft' CHECK (status IN ('Draft', 'Finalized')),
    created_by INTEGER REFERENCES users(id),
    finalized_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE (building_id, month, year)
);

CREATE TABLE electricity_bill_run_items (
    id SERIAL PRIMARY KEY,
    billing_run_id INTEGER NOT NULL REFERENCES electricity_bill_runs(id) ON DELETE CASCADE,
    unit_id INTEGER NOT NULL REFERENCES units(id),
    unit_number VARCHAR(50) NOT NULL,
    unit_type VARCHAR(50) NOT NULL,
    floor INTEGER NOT NULL DEFAULT 0,
    meter_number VARCHAR(50) NOT NULL DEFAULT '',
    electricity_amount DECIMAL(10,2) CHECK (electricity_amount IS NULL OR electricity_amount >= 0),
    UNIQUE (billing_run_id, unit_id)
);

CREATE INDEX idx_electricity_bill_runs_org_building
    ON electricity_bill_runs (organization_id, building_id);
CREATE INDEX idx_electricity_bill_run_items_run
    ON electricity_bill_run_items (billing_run_id);
