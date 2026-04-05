export type UnitType = 'Shop' | 'Apartment' | 'Office' | 'Parking' | 'Storage' | 'Other';

// Type-specific metadata interfaces
export interface ShopMetadata {
  area_sqft?: number;
  category?: 'retail' | 'food' | 'service';
  has_utilities?: boolean;
  frontage_feet?: number;
}

export interface ApartmentMetadata {
  bedrooms?: number;
  bathrooms?: number;
  area_sqft?: number;
  furnished?: boolean;
  balcony_count?: number;
  parking_included?: boolean;
}

export interface OfficeMetadata {
  area_sqft?: number;
  cabin_count?: number;
  workstations?: number;
  conference_room?: boolean;
  internet_included?: boolean;
}

export interface ParkingMetadata {
  slot_number?: string;
  vehicle_type?: 'car' | 'bike' | 'both';
  covered?: boolean;
  ev_charging?: boolean;
}

export interface StorageMetadata {
  area_sqft?: number;
  climate_controlled?: boolean;
  access_hours?: string;
  security_level?: string;
}

export type UnitMetadata =
  | ShopMetadata
  | ApartmentMetadata
  | OfficeMetadata
  | ParkingMetadata
  | StorageMetadata
  | Record<string, unknown>;

export interface Unit {
  id: number;
  building_id: number;
  property_id: number;
  unit_number: string;
  unit_name?: string;
  floor?: number;
  section?: string;
  unit_type: UnitType;
  monthly_rent: number;
  metadata?: UnitMetadata;
  active: boolean;
  created_at: string; // ISO 8601 UTC timestamp
  updated_at: string; // ISO 8601 UTC timestamp
}

export interface UnitWithDetails extends Unit {
  property_name: string;
  building_name: string;
  building_code: string;
  tenant_name?: string;
  lease_active: boolean;
}

export interface CreateUnitRequest {
  building_id: number;
  property_id: number;
  unit_number: string;
  unit_name?: string;
  floor?: number;
  section?: string;
  unit_type: UnitType;
  monthly_rent: number;
  metadata?: UnitMetadata;
}

export interface UpdateUnitRequest {
  unit_name?: string;
  floor?: number;
  section?: string;
  unit_type?: UnitType;
  monthly_rent?: number;
  metadata?: UnitMetadata;
  active?: boolean;
}
