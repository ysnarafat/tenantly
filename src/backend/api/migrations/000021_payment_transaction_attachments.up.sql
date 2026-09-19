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
