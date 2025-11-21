-- Enhanced indexes for property management system performance

-- Essential property indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_properties_code_active ON properties(property_code, active);
CREATE INDEX IF NOT EXISTS idx_properties_type_active ON properties(property_type, active);
CREATE INDEX IF NOT EXISTS idx_properties_active_created ON properties(active, created_at DESC);

-- Building indexes for property relationships
CREATE INDEX IF NOT EXISTS idx_buildings_property_active ON buildings(property_id, active);

-- Unit indexes for property aggregations
CREATE INDEX IF NOT EXISTS idx_units_property_active ON units(property_id, active);
CREATE INDEX IF NOT EXISTS idx_units_building_active ON units(building_id, active);

-- Lease indexes for occupancy calculations
CREATE INDEX IF NOT EXISTS idx_leases_unit_active ON leases(unit_id, active);

-- Payment indexes for revenue calculations
CREATE INDEX IF NOT EXISTS idx_payments_property_status ON payments(property_id, status);
-- Add unique constraint for active property codes (case-insensitive)
CREATE UNIQUE INDEX IF NOT EXISTS idx_properties_code_unique_active 
ON properties (LOWER(property_code)) 
WHERE active = true;