-- Drop user_mfa
DROP TABLE IF EXISTS user_mfa;

-- Drop notification queue trigger
DROP TRIGGER IF EXISTS notify_notification_queue_insert_trigger ON notification_queue;
DROP FUNCTION IF EXISTS notify_notification_queue_insert();

-- Drop organization onboarding columns
ALTER TABLE user_invitations
    DROP COLUMN IF EXISTS last_sent_at,
    DROP COLUMN IF EXISTS email_delivery_status;