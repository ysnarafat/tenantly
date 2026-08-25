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

// amount_paid, status, and receipt_number are intentionally absent —
// recording money received must go through PaymentService.recordTransaction
// instead (see PaymentTransaction below), so partial/installment payments
// accumulate correctly instead of overwriting each other. Status and the
// receipt number are always derived/generated server-side either way.
export interface UpdatePaymentRequest {
  payment_method?: string;
  payment_date?: string;
  notes?: string;
}

// One amount actually received against a payment. A payment can be settled
// across several of these (partial/installment payments) — amount_paid on
// the parent Payment is the sum of its transactions.
export interface PaymentTransaction {
  id: number;
  payment_id: number;
  amount: number;
  payment_method?: string;
  payment_date: string; // ISO 8601 date
  receipt_number?: string;
  notes?: string;
  created_at: string; // ISO 8601 UTC timestamp
}

export interface CreatePaymentTransactionRequest {
  amount: number;
  payment_method?: string;
  payment_date?: string; // defaults to today when omitted
  notes?: string;
}

// Max file size accepted for a payment transaction attachment, in bytes —
// mirrors models.MaxAttachmentFileSize on the backend.
export const MAX_ATTACHMENT_FILE_SIZE = 10 * 1024 * 1024;

// MIME types accepted for a payment transaction attachment — mirrors
// models.AllowedAttachmentContentTypes on the backend. The backend is the
// source of truth (it sniffs actual file bytes); this list is only used to
// give the user an immediate error before an upload is attempted.
export const ALLOWED_ATTACHMENT_CONTENT_TYPES = [
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'application/pdf',
];

// A file (receipt photo, bKash/Nagad screenshot, etc.) evidencing one
// installment of a payment.
export interface PaymentTransactionAttachment {
  id: number;
  payment_transaction_id: number;
  file_name: string;
  content_type: string;
  file_size: number;
  uploaded_by?: number;
  created_at: string; // ISO 8601 UTC timestamp
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
  /** Sum of unpaid amounts (Due/Partial/Overdue) across all periods for this unit. */
  outstanding_balance: number;
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
