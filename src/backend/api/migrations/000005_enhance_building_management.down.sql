-- Rollback Enhanced Building Management System Migration

-- Drop enhanced indexes
-- Note: keeping the original unique constraint intact
DROP INDEX IF EXISTS idx_buildings_property_id_enhanced;
DROP INDEX IF EXISTS idx_buildings_type_enhanced;
DROP INDEX IF EXISTS idx_buildings_active_status;
DROP INDEX IF EXISTS idx_buildings_property_type_active;
DROP INDEX IF EXISTS idx_buildings_metadata_gin;
DROP INDEX IF EXISTS idx_buildings_floors;
DROP INDEX IF EXISTS idx_buildings_elevator;
DROP INDEX IF EXISTS idx_buildings_construction_year;
DROP INDEX IF EXISTS idx_buildings_active_status_created;
DROP INDEX IF EXISTS idx_buildings_property_active_status;
DROP INDEX IF EXISTS idx_units_building_active_status;

-- Recreate original indexes
CREATE INDEX idx_buildings_metadata ON buildings USING GIN(metadata);
CREATE INDEX idx_buildings_active ON buildings(active_status);
CREATE INDEX idx_buildings_property_active ON buildings(property_id, active_status);

-- Add back the active column
ALTER TABLE buildings ADD COLUMN active BOOLEAN DEFAULT true;

-- Copy data from active_status to active
UPDATE buildings SET active = active_status;

-- Drop active_status column
ALTER TABLE buildings DROP COLUMN active_status;

-- Revert building_type to VARCHAR with CHECK constraint
ALTER TABLE buildings ADD COLUMN building_type_old VARCHAR(50);
UPDATE buildings SET building_type_old = building_type::text;
ALTER TABLE buildings DROP COLUMN building_type;
ALTER TABLE buildings RENAME COLUMN building_type_old TO building_type;
ALTER TABLE buildings ALTER COLUMN building_type SET NOT NULL;
ALTER TABLE buildings ADD CONSTRAINT buildings_building_type_check 
    CHECK (building_type IN ('Residential', 'Commercial', 'Mixed'));

-- Drop enhanced constraints
ALTER TABLE buildings DROP CONSTRAINT IF EXISTS chk_building_name_not_empty;
ALTER TABLE buildings DROP CONSTRAINT IF EXISTS chk_building_code_not_empty;
ALTER TABLE buildings DROP CONSTRAINT IF EXISTS chk_total_floors_positive;
ALTER TABLE buildings DROP CONSTRAINT IF EXISTS chk_construction_year_valid;

-- The original unique constraint remains intact

-- Drop the building_type_enum
DROP TYPE IF EXISTS building_type_enum;

-- Remove comments
COMMENT ON TABLE buildings IS NULL;
COMMENT ON COLUMN buildings.building_type IS NULL;
COMMENT ON COLUMN buildings.metadata IS NULL;