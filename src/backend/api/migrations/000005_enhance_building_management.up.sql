-- Enhanced Building Management System Migration
-- This migration enhances the existing buildings table to fully support the building management system requirements

-- Create building_type_enum with required values
CREATE TYPE building_type_enum AS ENUM ('Residential', 'Commercial', 'Mixed');

-- Add active_status column (rename from active for consistency with requirements)
ALTER TABLE buildings ADD COLUMN active_status BOOLEAN DEFAULT true;

-- Copy data from active to active_status
UPDATE buildings SET active_status = active;

-- Drop the old active column
ALTER TABLE buildings DROP COLUMN active;

-- Modify building_type column to use the new enum type
-- First, add a temporary column with the enum type
ALTER TABLE buildings ADD COLUMN building_type_new building_type_enum;

-- Update the new column with converted values
UPDATE buildings SET building_type_new = building_type::building_type_enum;

-- Drop the old column and rename the new one
ALTER TABLE buildings DROP COLUMN building_type;
ALTER TABLE buildings RENAME COLUMN building_type_new TO building_type;

-- Make building_type NOT NULL
ALTER TABLE buildings ALTER COLUMN building_type SET NOT NULL;

-- Ensure all required fields are NOT NULL and have proper constraints
ALTER TABLE buildings ALTER COLUMN building_name SET NOT NULL;
ALTER TABLE buildings ALTER COLUMN building_code SET NOT NULL;
ALTER TABLE buildings ALTER COLUMN property_id SET NOT NULL;

-- Add constraints for data validation
ALTER TABLE buildings ADD CONSTRAINT chk_building_name_not_empty CHECK (LENGTH(TRIM(building_name)) > 0);
ALTER TABLE buildings ADD CONSTRAINT chk_building_code_not_empty CHECK (LENGTH(TRIM(building_code)) > 0);
ALTER TABLE buildings ADD CONSTRAINT chk_total_floors_positive CHECK (total_floors > 0);
ALTER TABLE buildings ADD CONSTRAINT chk_construction_year_valid 
    CHECK (construction_year IS NULL OR (construction_year >= 1800 AND construction_year <= EXTRACT(YEAR FROM CURRENT_DATE) + 5));

-- Ensure the composite unique index exists for (property_id, building_code)
-- The existing unique constraint already provides the required uniqueness
-- We'll keep the existing constraint and add additional indexes for performance

-- Create optimized indexes for query performance as specified in requirements

-- Index for property_id queries
CREATE INDEX IF NOT EXISTS idx_buildings_property_id_enhanced ON buildings(property_id) WHERE active_status = true;

-- Index for building_type queries  
CREATE INDEX IF NOT EXISTS idx_buildings_type_enhanced ON buildings(building_type) WHERE active_status = true;

-- Index for active_status queries
CREATE INDEX IF NOT EXISTS idx_buildings_active_status ON buildings(active_status);

-- Composite index for property and type queries
CREATE INDEX IF NOT EXISTS idx_buildings_property_type_active ON buildings(property_id, building_type, active_status);

-- GIN index for building-specific metadata attributes (ensure it exists and is optimized)
DROP INDEX IF EXISTS idx_buildings_metadata;
CREATE INDEX idx_buildings_metadata_gin ON buildings USING GIN (metadata) WHERE metadata IS NOT NULL;

-- Additional performance indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_buildings_floors ON buildings(total_floors) WHERE active_status = true;
CREATE INDEX IF NOT EXISTS idx_buildings_elevator ON buildings(has_elevator) WHERE active_status = true;
CREATE INDEX IF NOT EXISTS idx_buildings_construction_year ON buildings(construction_year) WHERE active_status = true AND construction_year IS NOT NULL;

-- Update existing indexes that referenced the old 'active' column
DROP INDEX IF EXISTS idx_buildings_active;
DROP INDEX IF EXISTS idx_buildings_property_active;
DROP INDEX IF EXISTS idx_units_building_active;

-- Recreate indexes with active_status
CREATE INDEX IF NOT EXISTS idx_buildings_active_status_created ON buildings(active_status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_buildings_property_active_status ON buildings(property_id, active_status);

-- Update units table indexes to use active_status for buildings
-- Note: Cannot use subquery in index predicate, so creating a simple index instead
CREATE INDEX IF NOT EXISTS idx_units_building_active_status ON units(building_id);

-- Add comment to document the enhanced building management system
COMMENT ON TABLE buildings IS 'Enhanced building management table supporting residential, commercial, and mixed-use buildings with comprehensive metadata and indexing';
COMMENT ON COLUMN buildings.building_type IS 'Building classification using enum: Residential, Commercial, or Mixed';
COMMENT ON COLUMN buildings.metadata IS 'JSONB field for building-specific attributes like amenities, parking, security systems';
COMMENT ON COLUMN buildings.active_status IS 'Boolean flag indicating if the building is active in the system';