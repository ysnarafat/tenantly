import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import {
  CreatePaymentRequest,
  DashboardSummary,
  Payment,
  PaymentListResponse,
  UpdatePaymentRequest,
  PaymentWithDetails,
  LeaseSearchResponse,
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
}
