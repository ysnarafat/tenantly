CREATE TABLE receipt_access_tokens (
    id SERIAL PRIMARY KEY,
    token VARCHAR(64) NOT NULL UNIQUE,
    payment_id INTEGER NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_receipt_access_tokens_payment_id ON receipt_access_tokens(payment_id);
