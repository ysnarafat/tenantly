import { LeaseType } from './lease.model';

export type UnitType = 'Shop' | 'Apartment' | 'Office' | 'Parking' | 'Storage' | 'Other';

/**
 * Proposed commercial terms for a unit, used to pre-fill the lease-creation
 * form. These are not a tenancy: nothing is owed and no tenant is implied
 * until a real lease exists. Every field is optional.
 */
export interface UnitLeaseDefaults {
  default_lease_type?: LeaseType;
  default_monthly_rent?: number;
  default_security_deposit?: number;
  default_duration_months?: number;
}

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

export interface Unit extends UnitLeaseDefaults {
  id: number;
  building_id: number;
  property_id: number;
  unit_number: string;
  unit_name?: string;
  floor?: number;
  section?: string;
  unit_type: UnitType;
  metadata?: UnitMetadata;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UnitWithDetails extends Unit {
  property_name: string;
  building_name: string;
  building_code: string;
  tenant_name?: string;
  lease_active: boolean;
}

export interface CreateUnitRequest extends UnitLeaseDefaults {
  building_id: number;
  property_id: number;
  unit_number: string;
  unit_name?: string;
  floor?: number;
  section?: string;
  unit_type: UnitType;
  metadata?: UnitMetadata;
}

export interface UpdateUnitRequest extends UnitLeaseDefaults {
  unit_name?: string;
  floor?: number;
  section?: string;
  unit_type?: UnitType;
  metadata?: UnitMetadata;
  active?: boolean;
}

export interface BulkCreateUnitItem extends UnitLeaseDefaults {
  unit_number: string;
  unit_name?: string;
  floor?: number;
  section?: string;
  unit_type: UnitType;
}

export interface BulkCreateUnitsRequest {
  units: BulkCreateUnitItem[];
}

export interface BulkCreateUnitsResponse {
  message: string;
  building_id: number;
  units: Unit[];
  summary: { units_created: number };
}
