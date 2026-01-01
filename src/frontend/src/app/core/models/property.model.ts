export type PropertyType = 'Residential' | 'Commercial' | 'Mixed';

export type PropertyMetadata = Record<string, any>;

export interface Property {
  id: number;
  property_name: string;
  property_code: string;
  address: string;
  city?: string;
  postal_code?: string;
  property_type: PropertyType;
  total_buildings: number;
  metadata?: PropertyMetadata;
  active: boolean;
  created_at: string; // ISO 8601 UTC timestamp
  updated_at: string; // ISO 8601 UTC timestamp
}

export interface PropertyWithStats extends Property {
  building_count: number;
  unit_count: number;
  occupied_units: number;
  total_revenue: number;
}

export interface CreatePropertyRequest {
  property_name: string;
  property_code: string;
  address: string;
  city?: string;
  postal_code?: string;
  property_type: PropertyType;
  metadata?: PropertyMetadata;
}

export interface UpdatePropertyRequest {
  property_name?: string;
  address?: string;
  city?: string;
  postal_code?: string;
  property_type?: PropertyType;
  total_buildings?: number;
  metadata?: PropertyMetadata;
  active?: boolean;
}
