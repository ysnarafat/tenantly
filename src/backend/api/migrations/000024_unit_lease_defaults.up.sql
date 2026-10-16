-- Lease defaults on units.
--
-- These columns hold the *proposed* commercial terms for a unit — the numbers a
-- new lease should be pre-filled with — not an actual tenancy. They exist so a
-- user setting up a building (especially via bulk unit creation) can enter rent
-- and lease terms once per unit instead of retyping them for every lease.
--
-- All nullable: a unit with no defaults simply pre-fills nothing. Nothing in
-- reporting or due generation reads these columns; only the lease-creation form
-- does. Actual tenancies continue to live exclusively in `leases`.
ALTER TABLE units
    ADD COLUMN default_lease_type VARCHAR(20)
        CHECK (default_lease_type IN ('Residential', 'Commercial')),
    ADD COLUMN default_monthly_rent DECIMAL(10, 2)
        CHECK (default_monthly_rent >= 0),
    ADD COLUMN default_security_deposit DECIMAL(10, 2)
        CHECK (default_security_deposit >= 0),
    ADD COLUMN default_duration_months INTEGER
        CHECK (default_duration_months BETWEEN 1 AND 600);
