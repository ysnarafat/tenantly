export type BillingRunStatus = 'Draft' | 'Finalized';

export interface BillingRunItem {
  id: number;
  unit_id: number;
  unit_number: string;
  unit_type: string;
  floor: number;
  meter_number: string;
  electricity_amount: number | null;
}

export interface BillingRun {
  id: number;
  building_id: number;
  organization_id: number;
  property_name: string;
  building_name: string;
  building_code: string;
  month: number;
  year: number;
  status: BillingRunStatus;
  finalized_at?: string;
  created_at: string;
  items: BillingRunItem[];
}

export interface CreateBillingRunRequest {
  building_id: number;
  month: number;
  year: number;
}
