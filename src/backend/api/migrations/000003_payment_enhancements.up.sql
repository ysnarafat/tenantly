-- Not CONCURRENTLY: this file has multiple statements, which golang-migrate
-- runs as one implicit transaction, and CONCURRENTLY cannot run inside one.
CREATE INDEX IF NOT EXISTS idx_payments_org_status_month_year
    ON payments (organization_id, status, year, month);

CREATE TABLE payment_receipt_sequences (
    organization_id INTEGER PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    year_month VARCHAR(6) NOT NULL DEFAULT '',
    next_seq INTEGER NOT NULL DEFAULT 1
);
