CREATE TABLE payment_receipt_sequences (
    organization_id INTEGER PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    year_month VARCHAR(6) NOT NULL DEFAULT '',
    next_seq INTEGER NOT NULL DEFAULT 1
);
