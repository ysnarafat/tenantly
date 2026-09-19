-- Notify listeners immediately when a notification is queued, so the .NET
-- notification service can react without waiting for its next poll cycle.
-- The payload carries only the new row's id; consumers still re-query the
-- table rather than trusting the payload as the source of truth.
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
