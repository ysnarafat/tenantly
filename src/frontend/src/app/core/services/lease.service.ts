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

@Injectable({
  providedIn: 'root',
})
export class LeaseService {
  private http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/leases`;

  // DEMO MODE: Set to true to enable demo data (disable for production)
  private readonly DEMO_MODE = false;

  private demoLeases: LeaseWithDetails[] = [
    {
      id: 1,
      unit_id: 1,
      tenant_id: 1,
      lease_type: 'Residential',
      start_date: '2024-01-01',
      end_date: '2024-12-31',
      duration_months: 12,
      monthly_rent: 15000,
      security_deposit: 30000,
      active: true,
      organization_id: 1,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      building_id: 1,
      property_id: 1,
      property_name: 'Sunrise Apartments',
      building_name: 'Building A',
      building_code: 'BLD-A',
      unit_number: '101',
      unit_type: '1BHK',
      tenant_name: 'Ahmed Hassan',
      tenant_phone: '+8801712345678',
      is_expired: false,
      days_remaining: 120,
    },
    {
      id: 2,
      unit_id: 2,
      tenant_id: 2,
      lease_type: 'Commercial',
      start_date: '2023-06-01',
      end_date: '2025-05-31',
      duration_months: 24,
      monthly_rent: 20000,
      security_deposit: 40000,
      active: true,
      organization_id: 1,
      created_at: '2023-06-01T00:00:00Z',
      updated_at: '2023-06-01T00:00:00Z',
      building_id: 2,
      property_id: 1,
      property_name: 'Commercial Plaza',
      building_name: 'Building B',
      building_code: 'BLD-B',
      unit_number: '201',
      unit_type: 'Office Space',
      tenant_name: 'Fatima Rahman',
      tenant_phone: '+8801812345678',
      is_expired: false,
      days_remaining: 365,
    },
  ];

  getAllLeases(page: number = 1, pageSize: number = 10): Observable<LeaseListResponse> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const totalItems = this.demoLeases.length;
        const totalPages = Math.ceil(totalItems / pageSize);
        const startIndex = (page - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        const paginatedLeases = this.demoLeases.slice(startIndex, endIndex);

        observer.next({
          leases: paginatedLeases,
          pagination: {
            current_page: page,
            page_size: pageSize,
            total_items: totalItems,
            total_pages: totalPages,
            has_next: page < totalPages,
            has_prev: page > 1,
          },
        });
        observer.complete();
      });
    }

    const params = new HttpParams()
      .set('page', page.toString())
      .set('page_size', pageSize.toString());

    return this.http.get<LeaseListResponse>(this.apiUrl, { params });
  }

  getLeaseById(id: number): Observable<LeaseDetailResponse> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const lease = this.demoLeases.find((l) => l.id === id);
        if (lease) {
          observer.next({
            ...lease,
            status: lease.active ? 'Active' : 'Expired',
            payment_history: [],
            total_payments: 0,
          });
        } else {
          observer.error({ error: 'Lease not found' });
        }
        observer.complete();
      });
    }

    return this.http.get<LeaseDetailResponse>(`${this.apiUrl}/${id}`);
  }

  createLease(request: CreateLeaseRequest): Observable<LeaseWithDetails> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const newLease: LeaseWithDetails = {
          id: Math.max(...this.demoLeases.map((l) => l.id)) + 1,
          ...request,
          end_date:
            request.end_date || this.calculateEndDate(request.start_date, request.duration_months),
          security_deposit: request.security_deposit || 0,
          active: true,
          organization_id: 1,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          building_id: 1,
          property_id: 1,
          property_name: 'Demo Property',
          building_name: 'Demo Building',
          building_code: 'DEMO',
          unit_number: 'DEMO-001',
          unit_type: 'Demo Unit',
          tenant_name: 'Demo Tenant',
          tenant_phone: '+8800000000000',
          is_expired: false,
          days_remaining: 365,
        };
        this.demoLeases.push(newLease);
        observer.next(newLease);
        observer.complete();
      });
    }

    return this.http.post<LeaseWithDetails>(this.apiUrl, request);
  }

  updateLease(id: number, request: UpdateLeaseRequest): Observable<LeaseWithDetails> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const leaseIndex = this.demoLeases.findIndex((l) => l.id === id);
        if (leaseIndex !== -1) {
          this.demoLeases[leaseIndex] = {
            ...this.demoLeases[leaseIndex],
            ...request,
            updated_at: new Date().toISOString(),
          };
          observer.next(this.demoLeases[leaseIndex]);
        } else {
          observer.error({ error: 'Lease not found' });
        }
        observer.complete();
      });
    }

    return this.http.put<LeaseWithDetails>(`${this.apiUrl}/${id}`, request);
  }

  deleteLease(id: number): Observable<{ message: string }> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const leaseIndex = this.demoLeases.findIndex((l) => l.id === id);
        if (leaseIndex !== -1) {
          this.demoLeases.splice(leaseIndex, 1);
          observer.next({ message: 'Lease deleted successfully' });
        } else {
          observer.error({ error: 'Lease not found' });
        }
        observer.complete();
      });
    }

    return this.http.delete<{ message: string }>(`${this.apiUrl}/${id}`);
  }

  terminateLease(id: number, request: TerminateLeaseRequest = {}): Observable<{ message: string }> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const leaseIndex = this.demoLeases.findIndex((l) => l.id === id);
        if (leaseIndex !== -1) {
          this.demoLeases[leaseIndex].active = false;
          this.demoLeases[leaseIndex].updated_at = new Date().toISOString();
          observer.next({ message: 'Lease terminated successfully' });
        } else {
          observer.error({ error: 'Lease not found' });
        }
        observer.complete();
      });
    }

    return this.http.post<{ message: string }>(`${this.apiUrl}/${id}/terminate`, request);
  }

  getLeasesByUnit(
    unitId: number,
    page: number = 1,
    pageSize: number = 10
  ): Observable<LeaseListResponse> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const unitLeases = this.demoLeases.filter((l) => l.unit_id === unitId);
        observer.next({
          leases: unitLeases,
          pagination: {
            current_page: page,
            page_size: pageSize,
            total_items: unitLeases.length,
            total_pages: 1,
            has_next: false,
            has_prev: false,
          },
        });
        observer.complete();
      });
    }

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
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const tenantLeases = this.demoLeases.filter((l) => l.tenant_id === tenantId);
        observer.next({
          leases: tenantLeases,
          pagination: {
            current_page: page,
            page_size: pageSize,
            total_items: tenantLeases.length,
            total_pages: 1,
            has_next: false,
            has_prev: false,
          },
        });
        observer.complete();
      });
    }

    const params = new HttpParams()
      .set('page', page.toString())
      .set('page_size', pageSize.toString());

    return this.http.get<LeaseListResponse>(`${this.apiUrl}/tenant/${tenantId}`, { params });
  }

  getActiveLeases(): Observable<LeaseWithDetails[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const activeLeases = this.demoLeases.filter((l) => l.active);
        observer.next(activeLeases);
        observer.complete();
      });
    }

    return this.http
      .get<{ leases: LeaseWithDetails[] }>(this.apiUrl)
      .pipe(map((res) => res.leases ?? []));
  }

  getExpiringLeases(days = 30): Observable<LeaseWithDetails[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const expiringLeases = this.demoLeases.filter(
          (l) => l.active && l.days_remaining <= days && l.days_remaining > 0
        );
        observer.next(expiringLeases);
        observer.complete();
      });
    }

    const params = new HttpParams().set('expiring_in_days', days.toString());
    return this.http.get<LeaseWithDetails[]>(this.apiUrl, { params });
  }

  getExpiredLeases(): Observable<LeaseWithDetails[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const expiredLeases = this.demoLeases.filter((l) => l.is_expired);
        observer.next(expiredLeases);
        observer.complete();
      });
    }

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
