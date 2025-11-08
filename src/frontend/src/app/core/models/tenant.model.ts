export type TenantType = 'Individual' | 'Business';

export interface Tenant {
  id: number;
  name: string;
  tenant_type: TenantType;
  phone_number?: string;
  email?: string;
  nid_number?: string;
  address?: string;
  active: boolean;
  created_at: string; // ISO 8601 UTC timestamp
  updated_at: string; // ISO 8601 UTC timestamp
}

export interface TenantWithLeases extends Tenant {
  active_leases: number;
  total_units: number;
}

export interface CreateTenantRequest {
  name: string;
  tenant_type: TenantType;
  phone_number?: string;
  email?: string;
  nid_number?: string;
  address?: string;
}

export interface UpdateTenantRequest {
  name?: string;
  tenant_type?: TenantType;
  phone_number?: string;
  email?: string;
  nid_number?: string;
  address?: string;
  active?: boolean;
}
