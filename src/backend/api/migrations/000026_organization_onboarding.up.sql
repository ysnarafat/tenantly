ALTER TABLE user_invitations
    ADD COLUMN email_delivery_status VARCHAR(20) NOT NULL DEFAULT 'unknown'
        CHECK (email_delivery_status IN ('unknown', 'pending', 'sent', 'failed')),
    ADD COLUMN last_sent_at TIMESTAMP;
