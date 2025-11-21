-- Rollback enhanced property indexes

-- Drop constraints
ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_property_name_not_empty;
ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_property_code_not_empty;
ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_address_not_empty;

-- Drop indexes
DROP INDEX IF EXISTS idx_properties_code_active;
DROP INDEX IF EXISTS idx_properties_type_active;
DROP INDEX IF EXISTS idx_properties_active_created;
DROP INDEX IF EXISTS idx_properties_name_unique_active;
DROP INDEX IF EXISTS idx_properties_code_unique_active;

DROP INDEX IF EXISTS idx_buildings_property_active;

DROP INDEX IF EXISTS idx_units_property_active;
DROP INDEX IF EXISTS idx_units_building_active;

DROP INDEX IF EXISTS idx_leases_unit_active;

DROP INDEX IF EXISTS idx_payments_property_status;
DROP INDEX IF EXISTS idx_payments_unit_status;

DROP INDEX IF EXISTS idx_audit_log_table_record;