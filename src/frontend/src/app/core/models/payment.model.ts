export type PaymentStatus = 'Paid' | 'Due' | 'Partial' | 'Overdue';

export interface Payment {
  id: number;
  unit_id: number;
  tenant_id: number;
  building_id: number;
  property_id: number;
  month: number;
  year: number;
  amount_due: number;
  amount_paid: number;
  status: PaymentStatus;
  payment_method?: string;
  payment_date?: string; // ISO 8601 date
  notes?: string;
  receipt_number?: string;
  due_date?: string; // ISO 8601 date
  created_at: string; // ISO 8601 UTC timestamp
  updated_at: string; // ISO 8601 UTC timestamp
}

export interface PaymentWithDetails extends Payment {
  property_name: string;
  building_name: string;
  building_code: string;
  unit_number: string;
  unit_type: string;
  tenant_name: string;
}

export interface CreatePaymentRequest {
  unit_id: number;
  tenant_id: number;
  building_id: number;
  property_id: number;
  month: number;
  year: number;
  amount_due: number;
  amount_paid?: number;
  status?: PaymentStatus;
  payment_method?: string;
  payment_date?: string;
  receipt_number?: string;
  notes?: string;
  due_date?: string;
}

export interface UpdatePaymentRequest {
  amount_paid?: number;
  status?: PaymentStatus;
  payment_method?: string;
  payment_date?: string;
  notes?: string;
  receipt_number?: string;
}

export interface PaymentListResponse {
  payments: PaymentWithDetails[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface DashboardSummary {
  total_due: number;
  total_paid: number;
  total_pending: number;
  total_overdue: number;
  collection_rate: number;
  property_count: number;
  building_count: number;
  unit_count: number;
  tenant_count: number;
}

export interface LeaseSearchResult {
  lease_id: number;
  tenant_id: number;
  tenant_name: string;
  tenant_phone: string;
  property_id: number;
  property_name: string;
  building_id: number;
  building_name: string;
  building_code: string;
  unit_id: number;
  unit_number: string;
  unit_type: string;
  lease_start_date: string; // ISO 8601 date
  lease_end_date: string; // ISO 8601 date
  monthly_rent: number;
  active: boolean;
}

export interface LeaseSearchResponse {
  results: LeaseSearchResult[];
  total: number;
}

export interface GenerateMonthlyPaymentsRequest {
  month: number;
  year: number;
  building_id?: number;
  due_day_of_month?: number;
}

export interface GenerateMonthlyPaymentsResult {
  generated: number;
  skipped: number;
  failed: number;
  errors?: string[];
}
