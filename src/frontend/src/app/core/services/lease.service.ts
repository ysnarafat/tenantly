import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { environment } from '../../../environments/environment';

export type LeaseType = 'Residential' | 'Commercial';

export interface LeaseDue {
  lease_id: number;
  tenant_id: number;
  tenant_name: string;
  unit_id: number;
  unit_number: string;
  unit_type: string;
  building_id: number;
  building_name: string;
  building_code: string;
  property_id: number;
  property_name: string;
  monthly_rent: number;
  days_overdue: number;
  organization_id: number;
}

export interface DueSummary {
  total_due_amount: number;
  total_tenants_due: number;
}

export type LeaseEndReason = 'Expired' | 'Terminated' | 'Renewed';

export interface Lease {
  id: number;
  unit_id: number;
  tenant_id: number;
  lease_type: LeaseType;
  start_date: string;
  end_date: string;
  duration_months: number;
  monthly_rent: number;
  security_deposit: number;
  active: boolean;
  organization_id: number;
  // Present once a lease has stopped being active — see LeaseEndReason.
  end_reason?: LeaseEndReason;
  // Present when this lease was created by renewing an earlier one.
  renewed_from_lease_id?: number;
  created_at: string;
  updated_at: string;
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

export interface Payment {
  id: number;
  lease_id: number;
  amount: number;
  payment_date: string;
  payment_method: string;
  status: string;
  notes?: string;
  created_at: string;
}

export interface LeaseDetailResponse extends LeaseWithDetails {
  payment_history?: Payment[];
  total_payments?: number;
  status: 'Active' | 'Expired' | 'Terminated';
  termination_date?: string;
}

export interface PaginationInfo {
  current_page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface LeaseListResponse {
  leases: LeaseWithDetails[];
  pagination: PaginationInfo;
}

export interface CreateLeaseRequest {
  unit_id: number;
  tenant_id: number;
  lease_type: LeaseType;
  start_date: string;
  end_date?: string;
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

export interface TerminateLeaseRequest {
  termination_date?: string;
}

// Starts a new lease term for the same unit/tenant instead of mutating the
// current lease — fields left unset carry the corresponding value forward
// from the lease being renewed. See LeaseService.renewLease.
export interface RenewLeaseRequest {
  start_date?: string;
  duration_months: number;
  monthly_rent?: number;
  security_deposit?: number;
  lease_type?: LeaseType;
}

@Injectable({
  providedIn: 'root',
})
export class LeaseService {
  private http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/leases`;

  getAllLeases(page: number = 1, pageSize: number = 10): Observable<LeaseListResponse> {
    const params = new HttpParams()
      .set('page', page.toString())
      .set('page_size', pageSize.toString());

    return this.http.get<LeaseListResponse>(this.apiUrl, { params });
  }

  getLeaseById(id: number): Observable<LeaseDetailResponse> {
    return this.http.get<LeaseDetailResponse>(`${this.apiUrl}/${id}`);
  }

  createLease(request: CreateLeaseRequest): Observable<LeaseWithDetails> {
    return this.http.post<LeaseWithDetails>(this.apiUrl, request);
  }

  updateLease(id: number, request: UpdateLeaseRequest): Observable<LeaseWithDetails> {
    return this.http.put<LeaseWithDetails>(`${this.apiUrl}/${id}`, request);
  }

  deleteLease(id: number): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.apiUrl}/${id}`);
  }

  terminateLease(id: number, request: TerminateLeaseRequest = {}): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/${id}/terminate`, request);
  }

  renewLease(id: number, request: RenewLeaseRequest): Observable<LeaseWithDetails> {
    return this.http.post<LeaseWithDetails>(`${this.apiUrl}/${id}/renew`, request);
  }

  getLeasesByUnit(
    unitId: number,
    page: number = 1,
    pageSize: number = 10
  ): Observable<LeaseListResponse> {
    const params = new HttpParams()
      .set('page', page.toString())
      .set('page_size', pageSize.toString());

    return this.http.get<LeaseListResponse>(`${this.apiUrl}/unit/${unitId}`, { params });
  }

  getLeasesByTenant(
    tenantId: number,
    page: number = 1,
    pageSize: number = 10
  ): Observable<LeaseListResponse> {
    const params = new HttpParams()
      .set('page', page.toString())
      .set('page_size', pageSize.toString());

    return this.http.get<LeaseListResponse>(`${this.apiUrl}/tenant/${tenantId}`, { params });
  }

  getActiveLeases(): Observable<LeaseWithDetails[]> {
    return this.http
      .get<{ leases: LeaseWithDetails[] }>(this.apiUrl)
      .pipe(map((res) => res.leases ?? []));
  }

  getExpiringLeases(days = 30): Observable<LeaseWithDetails[]> {
    const params = new HttpParams().set('expiring_in_days', days.toString());
    return this.http.get<LeaseWithDetails[]>(this.apiUrl, { params });
  }

  getExpiredLeases(): Observable<LeaseWithDetails[]> {
    const params = new HttpParams().set('expired', 'true');
    return this.http.get<LeaseWithDetails[]>(this.apiUrl, { params });
  }

  private calculateEndDate(startDate: string, durationMonths: number): string {
    const start = new Date(startDate);
    const end = new Date(start);
    end.setMonth(end.getMonth() + durationMonths);
    return end.toISOString().split('T')[0];
  }

  getLeasesDue(): Observable<LeaseDue[]> {
    return this.http.get<LeaseDue[]>(`${this.apiUrl}/due`);
  }

  getDueSummary(): Observable<DueSummary> {
    return this.http.get<DueSummary>(`${this.apiUrl}/due/summary`);
  }
}
