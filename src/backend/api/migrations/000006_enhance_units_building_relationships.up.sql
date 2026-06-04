-- Enhanced Units-Building Relationships Migration (Simplified)
-- This migration enhances the units table to support hierarchical building relationships

-- Add hierarchical indexes on units table (property_id, building_id, floor, section)
-- Drop existing hierarchy index to recreate with proper fields
DROP INDEX IF EXISTS idx_units_hierarchy;

-- Create comprehensive hierarchical index for property -> building -> floor -> section queries
CREATE INDEX idx_units_hierarchy_enhanced ON units(property_id, building_id, floor, section) 
    WHERE active = true;

-- Create composite indexes for optimized hierarchical queries

-- Index for building-specific unit queries with floor ordering
CREATE INDEX idx_units_building_floor_section ON units(building_id, floor, section) 
    WHERE active = true;

-- Index for property-level unit queries across all buildings
CREATE INDEX idx_units_property_building_type ON units(property_id, building_id, unit_type) 
    WHERE active = true;

-- Index for hierarchical reporting and analytics
CREATE INDEX idx_units_hierarchy_analytics ON units(property_id, building_id, unit_type, active);

-- Update unit constraints to ensure building-property relationship integrity
-- Add basic validation constraints for data integrity

-- Ensure unit_number is not empty
ALTER TABLE units ADD CONSTRAINT chk_unit_number_not_empty 
    CHECK (LENGTH(TRIM(unit_number)) > 0);

-- Ensure floor is valid (positive number or ground floor = 0)
ALTER TABLE units ADD CONSTRAINT chk_floor_valid 
    CHECK (floor IS NULL OR floor >= 0);

-- Add comments for documentation
COMMENT ON INDEX idx_units_hierarchy_enhanced IS 'Hierarchical index for property -> building -> floor -> section queries';
COMMENT ON INDEX idx_units_building_floor_section IS 'Index for building-specific unit queries with floor and section ordering';
COMMENT ON INDEX idx_units_property_building_type IS 'Index for property-level unit queries across buildings by type';
COMMENT ON INDEX idx_units_hierarchy_analytics IS 'Index optimized for hierarchical reporting and analytics queries';