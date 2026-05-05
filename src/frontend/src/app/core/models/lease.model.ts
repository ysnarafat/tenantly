export type LeaseType = 'Residential' | 'Commercial';

export interface Lease {
  id: number;
  unit_id: number;
  tenant_id: number;
  lease_type: LeaseType;
  start_date: string; // ISO 8601 date
  end_date?: string; // ISO 8601 date
  duration_months: number;
  monthly_rent: number;
  security_deposit?: number;
  active: boolean;
  created_at: string; // ISO 8601 UTC timestamp
  updated_at: string; // ISO 8601 UTC timestamp
}

export interface LeaseWithDetails extends Lease {
  building_id: number;
  property_id: number;
  property_name: string;
  building_name: string;
  building_code: string;
  unit_number: string;
  unit_type: string;
  tenant_name: string;
  tenant_phone?: string;
  is_expired: boolean;
  days_remaining: number;
}

export interface CreateLeaseRequest {
  unit_id: number;
  tenant_id: number;
  lease_type: LeaseType;
  start_date: string;
  duration_months: number;
  monthly_rent: number;
  security_deposit?: number;
}

export interface UpdateLeaseRequest {
  lease_type?: LeaseType;
  start_date?: string;
  duration_months?: number;
  monthly_rent?: number;
  security_deposit?: number;
  active?: boolean;
}
