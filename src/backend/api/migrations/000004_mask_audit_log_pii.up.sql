-- Redact PII and secrets from trigger-generated audit rows.
--
-- The audit triggers store whole-row JSON via row_to_json/to_jsonb, which for
-- tables like `users` and `tenants` includes password hashes and personal data
-- (name, email, phone, national ID, address). mask_audit_json() replaces those
-- values with a placeholder so the audit trail still records that a field
-- changed, without persisting the sensitive value.

CREATE OR REPLACE FUNCTION mask_audit_json(data jsonb) RETURNS jsonb AS $$
DECLARE
    sensitive text[] := ARRAY[
        'password', 'password_hash', 'token', 'invitation_token',
        'refresh_token', 'reset_token',
        'name', 'first_name', 'last_name',
        'email', 'phone_number', 'nid_number', 'address'
    ];
    k text;
BEGIN
    IF data IS NULL THEN
        RETURN NULL;
    END IF;

    FOREACH k IN ARRAY sensitive LOOP
        -- Only redact keys that are present and non-null, so an empty field still
        -- reads as null in the trail rather than looking like it held a value.
        IF data ? k AND jsonb_typeof(data -> k) <> 'null' THEN
            data := jsonb_set(data, ARRAY[k], '"***REDACTED***"'::jsonb, false);
        END IF;
    END LOOP;

    RETURN data;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Rebuild the audit trigger to run row JSON through the mask. Triggers reference
-- the function by name, so replacing the body is enough — no need to recreate them.
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (action, table_name, record_id, old_values, created_at)
        VALUES ('DELETE', TG_TABLE_NAME, OLD.id, mask_audit_json(to_jsonb(OLD)), NOW() AT TIME ZONE 'UTC');
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (action, table_name, record_id, old_values, new_values, created_at)
        VALUES ('UPDATE', TG_TABLE_NAME, NEW.id, mask_audit_json(to_jsonb(OLD)), mask_audit_json(to_jsonb(NEW)), NOW() AT TIME ZONE 'UTC');
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (action, table_name, record_id, new_values, created_at)
        VALUES ('INSERT', TG_TABLE_NAME, NEW.id, mask_audit_json(to_jsonb(NEW)), NOW() AT TIME ZONE 'UTC');
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
