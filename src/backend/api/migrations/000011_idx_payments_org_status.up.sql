CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_org_status_month_year
    ON payments (organization_id, status, year, month);
