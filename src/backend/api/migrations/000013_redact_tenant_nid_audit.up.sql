-- Redact sensitive NID columns from audit_log payloads. The generic audit
-- trigger previously stored row_to_json(OLD/NEW) verbatim, which persisted the
-- (formerly plaintext) NID and now the ciphertext/hash into an easily-queried
-- table. Strip nid_number, nid_encrypted, and nid_hash for the tenants table
-- while keeping the non-sensitive nid_last_four for traceability.
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    old_payload JSONB;
    new_payload JSONB;
BEGIN
    IF TG_OP IN ('DELETE', 'UPDATE') THEN
        old_payload := row_to_json(OLD)::JSONB;
        IF TG_TABLE_NAME = 'tenants' THEN
            old_payload := old_payload - 'nid_number' - 'nid_encrypted' - 'nid_hash';
        END IF;
    END IF;

    IF TG_OP IN ('INSERT', 'UPDATE') THEN
        new_payload := row_to_json(NEW)::JSONB;
        IF TG_TABLE_NAME = 'tenants' THEN
            new_payload := new_payload - 'nid_number' - 'nid_encrypted' - 'nid_hash';
        END IF;
    END IF;

    IF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (action, table_name, record_id, old_values, created_at)
        VALUES ('DELETE', TG_TABLE_NAME, OLD.id, old_payload, NOW() AT TIME ZONE 'UTC');
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (action, table_name, record_id, old_values, new_values, created_at)
        VALUES ('UPDATE', TG_TABLE_NAME, NEW.id, old_payload, new_payload, NOW() AT TIME ZONE 'UTC');
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (action, table_name, record_id, new_values, created_at)
        VALUES ('INSERT', TG_TABLE_NAME, NEW.id, new_payload, NOW() AT TIME ZONE 'UTC');
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
