-- Lease lifecycle enhancements: end reason and renewal tracking
ALTER TABLE leases ADD COLUMN end_reason VARCHAR(20)
    CHECK (end_reason IN ('Expired', 'Terminated', 'Renewed'));
ALTER TABLE leases ADD COLUMN renewed_from_lease_id INTEGER REFERENCES leases(id);

CREATE INDEX idx_leases_renewed_from_lease_id ON leases(renewed_from_lease_id)
    WHERE renewed_from_lease_id IS NOT NULL;

COMMENT ON COLUMN leases.end_reason IS 'Why this lease stopped being active: naturally Expired, ended early via Terminated, or replaced via Renewed';
COMMENT ON COLUMN leases.renewed_from_lease_id IS 'If this lease was created by renewing an earlier one, the lease it replaced';

-- Recurring per-lease charges (utility, service charge, etc.)
CREATE TABLE lease_charges (
    id SERIAL PRIMARY KEY,
    lease_id INTEGER NOT NULL REFERENCES leases(id) ON DELETE CASCADE,
    charge_type VARCHAR(30) NOT NULL
        CHECK (charge_type IN ('Utility', 'ServiceCharge', 'Maintenance', 'Parking', 'Other')),
    label VARCHAR(100) NOT NULL,
    amount DECIMAL(10,2) NOT NULL CHECK (amount >= 0),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_lease_charges_lease_id ON lease_charges(lease_id);

COMMENT ON TABLE lease_charges IS 'Recurring per-lease charges (utility, service charge, etc.) added on top of monthly_rent when generating payments';
COMMENT ON COLUMN lease_charges.active IS 'Discontinued charges are deactivated rather than deleted, to keep billing history explainable';

-- Custom, org-defined key/value pairs not covered by structured lease fields
ALTER TABLE leases ADD COLUMN custom_fields JSONB;
COMMENT ON COLUMN leases.custom_fields IS 'Arbitrary org-defined key/value pairs not covered by structured lease fields';

-- Prevent concurrent active leases on same unit (database-level constraint)
CREATE UNIQUE INDEX idx_leases_unit_active_unique ON leases(unit_id) WHERE active = true;

-- Document attachments for tenants and leases
CREATE TABLE tenant_attachments (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    blob BYTEA NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_file_size CHECK (file_size <= 10485760),
    CONSTRAINT unique_tenant_attachment UNIQUE (tenant_id, file_name)
);

CREATE TABLE lease_attachments (
    id SERIAL PRIMARY KEY,
    lease_id INTEGER NOT NULL REFERENCES leases(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    blob BYTEA NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_file_size CHECK (file_size <= 10485760),
    CONSTRAINT unique_lease_attachment UNIQUE (lease_id, file_name)
);

CREATE INDEX idx_tenant_attachments_tenant_id ON tenant_attachments(tenant_id);
CREATE INDEX idx_lease_attachments_lease_id ON lease_attachments(lease_id);