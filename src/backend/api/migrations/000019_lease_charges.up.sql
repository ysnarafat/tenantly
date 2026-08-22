-- Recurring charges attached to a lease (utility, service charge, etc.),
-- on top of the base monthly_rent. An open set of named charges rather than
-- fixed columns, since an org may want any number of them per lease.
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
-- (e.g. parking spot #, referral source) — mirrors the metadata JSONB pattern
-- already used by properties/buildings/units.
ALTER TABLE leases ADD COLUMN custom_fields JSONB;
COMMENT ON COLUMN leases.custom_fields IS 'Arbitrary org-defined key/value pairs not covered by structured lease fields';
