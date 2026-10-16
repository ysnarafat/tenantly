ALTER TABLE units
    DROP COLUMN IF EXISTS default_lease_type,
    DROP COLUMN IF EXISTS default_monthly_rent,
    DROP COLUMN IF EXISTS default_security_deposit,
    DROP COLUMN IF EXISTS default_duration_months;
