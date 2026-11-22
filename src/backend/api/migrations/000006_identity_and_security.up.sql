-- Identity, security, and notification infrastructure

-- User MFA enrollment
CREATE TABLE IF NOT EXISTS user_mfa (
    user_id    INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret     TEXT NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
);

-- Notification queue trigger function
CREATE OR REPLACE FUNCTION notify_notification_queue_insert()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('notification_queue_channel', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_notification_queue_insert_trigger
    AFTER INSERT ON notification_queue
    FOR EACH ROW EXECUTE FUNCTION notify_notification_queue_insert();

-- Organization onboarding: track email delivery status for invitations
ALTER TABLE user_invitations
    ADD COLUMN email_delivery_status VARCHAR(20) NOT NULL DEFAULT 'unknown'
        CHECK (email_delivery_status IN ('unknown', 'pending', 'sent', 'failed')),
    ADD COLUMN last_sent_at TIMESTAMP;