import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface FinancialLedgerReport {
  organization_id: number;
  payments: any[];
  total: number;
  total_due: number;
  total_paid: number;
  total_pending: number;
  total_overdue: number;
  collection_rate: number;
  generated_at: string;
}

export interface CollectionSummaryReport {
  organization_id: number;
  collection_rate: number;
  total_due: number;
  total_collected: number;
  total_pending: number;
  total_overdue: number;
  aging_buckets: Record<string, number>;
  monthly_trend: Array<{
    month: string;
    collection_rate: number;
    amount_due: number;
    amount_collected: number;
  }>;
  report_period: string;
  generated_at: string;
}

export interface PaymentAnalysisReport {
  organization_id: number;
  payment_methods: Record<string, number>;
  status_distribution: Record<string, number>;
  daily_trend: Record<string, number>;
  total_payments: number;
  report_period: string;
  generated_at: string;
}

export interface DashboardMetrics {
  collection_summary: CollectionSummaryReport;
  payment_analysis: PaymentAnalysisReport;
  generated_at: string;
}

export interface TenantReportEntry {
  tenant_id: number;
  tenant_name: string;
  phone_number: string;
  email: string;
  unit_number: string;
  building_name: string;
  property_name: string;
  lease_start: string | null;
  lease_end: string | null;
  monthly_rent: number;
  lease_active: boolean;
  total_due: number;
  total_paid: number;
  balance_due: number;
}

export interface TenantSummaryReport {
  organization_id: number;
  tenants: TenantReportEntry[];
  total: number;
  active_tenants: number;
  generated_at: string;
}

export interface PropertyAnalyticsEntry {
  property_id: number;
  property_name: string;
  property_code: string;
  property_type: string;
  payment_stats: Record<string, any> | null;
}

export interface PropertyAnalyticsReport {
  organization_id: number;
  properties: PropertyAnalyticsEntry[];
  total: number;
  report_period: string;
  generated_at: string;
}

@Injectable({
  providedIn: 'root',
})
export class ReportService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/reports`;

  getFinancialLedger(
    page: number = 1,
    pageSize: number = 50,
    filters?: Record<string, any>
  ): Observable<FinancialLedgerReport> {
    let params = new HttpParams()
      .set('page', page.toString())
      .set('page_size', pageSize.toString());

    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== null && value !== undefined && value !== '') {
          params = params.set(key, value.toString());
        }
      });
    }

    return this.http.get<FinancialLedgerReport>(`${this.apiUrl}/ledger`, {
      params,
    });
  }

  getCollectionSummary(startDate?: string, endDate?: string): Observable<CollectionSummaryReport> {
    let params = new HttpParams();

    if (startDate) {
      params = params.set('start_date', startDate);
    }
    if (endDate) {
      params = params.set('end_date', endDate);
    }

    return this.http.get<CollectionSummaryReport>(`${this.apiUrl}/collection-summary`, { params });
  }

  getPaymentAnalysis(startDate?: string, endDate?: string): Observable<PaymentAnalysisReport> {
    let params = new HttpParams();

    if (startDate) {
      params = params.set('start_date', startDate);
    }
    if (endDate) {
      params = params.set('end_date', endDate);
    }

    return this.http.get<PaymentAnalysisReport>(`${this.apiUrl}/payment-analysis`, { params });
  }

  getDashboardMetrics(): Observable<DashboardMetrics> {
    return this.http.get<DashboardMetrics>(`${this.apiUrl}/dashboard-metrics`);
  }

  getTenantSummary(): Observable<TenantSummaryReport> {
    return this.http.get<TenantSummaryReport>(`${this.apiUrl}/tenant-summary`);
  }

  getPropertyAnalytics(startDate?: string, endDate?: string): Observable<PropertyAnalyticsReport> {
    let params = new HttpParams();
    if (startDate) params = params.set('start_date', startDate);
    if (endDate) params = params.set('end_date', endDate);
    return this.http.get<PropertyAnalyticsReport>(`${this.apiUrl}/property-analytics`, { params });
  }
}
