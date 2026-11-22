-- Payment receipt sequences for tracking payment order
CREATE INDEX IF NOT EXISTS idx_payments_org_status_month_year
    ON payments (organization_id, status, year, month);

CREATE TABLE payment_receipt_sequences (
    organization_id INTEGER PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    year_month VARCHAR(6) NOT NULL DEFAULT '',
    next_seq INTEGER NOT NULL DEFAULT 1
);

-- Payment transactions for partial/installment payments
CREATE TABLE payment_transactions (
    id SERIAL PRIMARY KEY,
    payment_id INTEGER NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(30),
    payment_date DATE NOT NULL,
    receipt_number VARCHAR(50) UNIQUE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_payment_transactions_payment_id ON payment_transactions(payment_id);

COMMENT ON TABLE payment_transactions IS 'One row per amount actually received against a payments row -- supports partial/installment payments without losing history. payments.amount_paid/status/payment_date/payment_method/receipt_number are maintained as a cached snapshot of the latest state, recomputed from this table whenever a transaction is added or removed.';

-- Payment transaction attachments for receipts/documents
CREATE TABLE payment_transaction_attachments (
    id SERIAL PRIMARY KEY,
    payment_transaction_id INTEGER NOT NULL REFERENCES payment_transactions(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size INTEGER NOT NULL CHECK (file_size > 0 AND file_size <= 10485760),
    file_data BYTEA NOT NULL,
    uploaded_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_payment_transaction_attachments_txn_id ON payment_transaction_attachments(payment_transaction_id);

-- Receipt access tokens for sharing receipts
CREATE TABLE receipt_access_tokens (
    id SERIAL PRIMARY KEY,
    payment_id INTEGER NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_receipt_access_tokens_payment_id ON receipt_access_tokens(payment_id);

-- Billing runs and items
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

-- Electricity bill runs and items
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

CREATE INDEX idx_electricity_bill_runs_org_building ON electricity_bill_runs (organization_id, building_id);
CREATE INDEX idx_electricity_bill_run_items_run ON electricity_bill_run_items (billing_run_id);