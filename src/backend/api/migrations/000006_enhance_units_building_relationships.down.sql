-- Rollback Enhanced Units-Building Relationships Migration (Simplified)

-- Drop the hierarchical indexes
DROP INDEX IF EXISTS idx_units_hierarchy_enhanced;
DROP INDEX IF EXISTS idx_units_building_floor_section;
DROP INDEX IF EXISTS idx_units_property_building_type;
DROP INDEX IF EXISTS idx_units_hierarchy_analytics;

-- Drop the validation constraints
ALTER TABLE units DROP CONSTRAINT IF EXISTS chk_unit_number_not_empty;
ALTER TABLE units DROP CONSTRAINT IF EXISTS chk_floor_valid;

-- Recreate the original hierarchy index
CREATE INDEX idx_units_hierarchy ON units(property_id, building_id, unit_type);