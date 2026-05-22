export type BuildingType = 'Residential' | 'Commercial' | 'Mixed';

export interface BuildingMetadata {
  // Commercial building attributes
  parking_spaces?: number;
  loading_docks?: number;
  security_system?: string;
  business_hours?: string;

  // Residential building attributes
  amenities?: string[];
  security_type?: string;
  maintenance_staff_count?: number;

  // Other attributes
  [key: string]: unknown;
}

export interface Building {
  id: number;
  property_id: number;
  building_name: string;
  building_code: string;
  building_type: BuildingType;
  total_floors?: number;
  has_elevator: boolean;
  construction_year?: number;
  metadata?: BuildingMetadata;
  active_status: boolean;
  created_at: string; // ISO 8601 UTC timestamp
  updated_at: string; // ISO 8601 UTC timestamp
}

export interface BuildingWithStats extends Building {
  property_name: string;
  unit_count: number;
  occupied_units: number;
  total_revenue: number;
  occupancy_rate: number;
}

export interface CreateBuildingRequest {
  property_id: number;
  building_name: string;
  building_code: string;
  building_type: BuildingType;
  total_floors?: number;
  has_elevator?: boolean;
  construction_year?: number;
  metadata?: BuildingMetadata;
}

export interface UpdateBuildingRequest {
  building_name?: string;
  building_type?: BuildingType;
  total_floors?: number;
  has_elevator?: boolean;
  construction_year?: number;
  metadata?: BuildingMetadata;
  active_status?: boolean;
}
