import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, forkJoin, of } from 'rxjs';
import { map, switchMap } from 'rxjs/operators';
import {
  CreatePaymentRequest,
  DashboardSummary,
  Payment,
  PaymentListResponse,
  UpdatePaymentRequest,
  PaymentWithDetails,
  LeaseSearchResponse,
  GenerateMonthlyPaymentsRequest,
  GenerateMonthlyPaymentsResult,
} from '../models/payment.model';
import { environment } from '../../../environments/environment';

export interface PaymentFilters {
  building_id?: number;
  property_id?: number;
  status?: string;
  month?: number;
  year?: number;
  page?: number;
  page_size?: number;
}

// Statuses the repository groups together as "outstanding" (see the shared
// `status IN ('Due', 'Partial', 'Overdue')` clauses on the backend) — the set a
// bulk-collection screen needs to show, since money is still owed on all three.
const OUTSTANDING_PAYMENT_STATUSES = ['Due', 'Partial', 'Overdue'] as const;

// Backend clamps page_size to 100 (see PaymentService.GetPayments), so any
// organization with more than 100 outstanding payments for a period needs more
// than one page per status.
const MAX_PAGE_SIZE = 100;

@Injectable({ providedIn: 'root' })
export class PaymentService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/payments`;

  getPayments(filters: PaymentFilters = {}): Observable<PaymentListResponse> {
    let params = new HttpParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params = params.set(key, String(value));
      }
    });
    return this.http.get<PaymentListResponse>(this.apiUrl, { params });
  }

  getPayment(id: number): Observable<PaymentWithDetails> {
    return this.http.get<PaymentWithDetails>(`${this.apiUrl}/${id}`);
  }

  createPayment(req: CreatePaymentRequest): Observable<Payment> {
    return this.http.post<Payment>(this.apiUrl, req);
  }

  updatePayment(id: number, req: UpdatePaymentRequest): Observable<Payment> {
    return this.http.put<Payment>(`${this.apiUrl}/${id}`, req);
  }

  bulkCreatePayments(requests: CreatePaymentRequest[]): Observable<unknown> {
    return this.http.post(`${this.apiUrl}/bulk`, requests);
  }

  downloadReceipt(id: number): Observable<Blob> {
    return this.http.get(`${this.apiUrl}/${id}/receipt`, { responseType: 'blob' });
  }

  getDashboardSummary(): Observable<DashboardSummary> {
    return this.http.get<DashboardSummary>(`${environment.apiUrl}/dashboard/summary`);
  }

  getBuildingReport(buildingId: number, startDate: string, endDate: string): Observable<unknown> {
    const params = new HttpParams().set('start_date', startDate).set('end_date', endDate);
    return this.http.get(`${this.apiUrl}/building/${buildingId}/report`, { params });
  }

  getPropertyReport(propertyId: number, startDate: string, endDate: string): Observable<unknown> {
    const params = new HttpParams().set('start_date', startDate).set('end_date', endDate);
    return this.http.get(`${this.apiUrl}/property/${propertyId}/report`, { params });
  }

  searchLeases(query: string): Observable<LeaseSearchResponse> {
    const params = new HttpParams().set('q', query);
    return this.http.get<LeaseSearchResponse>(`${this.apiUrl}/search`, { params });
  }

  generateMonthlyPayments(
    req: GenerateMonthlyPaymentsRequest
  ): Observable<GenerateMonthlyPaymentsResult> {
    return this.http.post<GenerateMonthlyPaymentsResult>(`${this.apiUrl}/generate-monthly`, req);
  }

  /**
   * Every outstanding (Due/Partial/Overdue) payment for a month/year, across as
   * many pages as needed — for a bulk-collection screen that needs the full set
   * up front rather than one paginated page at a time.
   */
  getAllDuePayments(month: number, year: number): Observable<PaymentWithDetails[]> {
    const requests = OUTSTANDING_PAYMENT_STATUSES.map((status) =>
      this.fetchAllPages({ month, year, status })
    );
    return forkJoin(requests).pipe(map((pages) => pages.flat()));
  }

  private fetchAllPages(
    filters: PaymentFilters,
    page = 1,
    accumulated: PaymentWithDetails[] = []
  ): Observable<PaymentWithDetails[]> {
    return this.getPayments({ ...filters, page, page_size: MAX_PAGE_SIZE }).pipe(
      switchMap((response) => {
        const combined = [...accumulated, ...response.payments];
        if (response.payments.length === 0 || combined.length >= response.total) {
          return of(combined);
        }
        return this.fetchAllPages(filters, page + 1, combined);
      })
    );
  }
}
