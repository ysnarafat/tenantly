-- A payments row is a monthly bill (one per unit/month/year); a tenant can
-- pay it off across multiple separate transactions (partial/installment
-- payments). Previously amount_paid/payment_date/payment_method/
-- receipt_number were single mutable columns on `payments`, so a second
-- installment overwrote the first instead of adding to it, and there was no
-- record that more than one payment ever happened. payment_transactions is
-- the source of truth for what was actually paid, when, and how — each
-- transaction gets its own receipt number, matching real invoicing practice.
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

COMMENT ON TABLE payment_transactions IS 'One row per amount actually received against a payments row — supports partial/installment payments without losing history. payments.amount_paid/status/payment_date/payment_method/receipt_number are maintained as a cached snapshot of the latest state, recomputed from this table whenever a transaction is added or removed.';

-- Backfill: preserve every existing payment's recorded amount as one
-- transaction, so pre-existing history isn't lost by this migration.
INSERT INTO payment_transactions (payment_id, amount, payment_method, payment_date, receipt_number, notes, created_at)
SELECT id, amount_paid, NULLIF(payment_method, ''), COALESCE(payment_date, created_at::date), receipt_number, NULLIF(notes, ''), created_at
FROM payments
WHERE amount_paid > 0;
