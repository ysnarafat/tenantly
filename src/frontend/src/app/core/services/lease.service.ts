import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface Lease {
  id: number;
  shop_id: number;
  tenant_id: number;
  start_date: string;
  duration_months: number;
  monthly_rent: number;
  security_deposit?: number;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface LeaseWithDetails extends Lease {
  shop_name: string;
  shop_number: string;
  tenant_name: string;
  tenant_phone: string;
  end_date: string;
  is_expired: boolean;
  days_remaining: number;
}

export interface CreateLeaseRequest {
  shop_id: number;
  tenant_id: number;
  start_date: string;
  duration_months: number;
  monthly_rent: number;
  security_deposit?: number;
}

export interface UpdateLeaseRequest {
  start_date?: string;
  duration_months?: number;
  monthly_rent?: number;
  security_deposit?: number;
  active?: boolean;
}

@Injectable({
  providedIn: 'root',
})
export class LeaseService {
  private http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/api/v1/leases`;

  // DEMO MODE: Set to true to enable demo data (disable for production)
  private readonly DEMO_MODE = true;

  private demoLeases: LeaseWithDetails[] = [
    {
      id: 1,
      shop_id: 1,
      tenant_id: 1,
      start_date: '2024-01-01',
      duration_months: 12,
      monthly_rent: 15000,
      security_deposit: 30000,
      active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      shop_name: 'Electronics Store',
      shop_number: 'S-001',
      tenant_name: 'Ahmed Hassan',
      tenant_phone: '+8801712345678',
      end_date: '2024-12-31',
      is_expired: false,
      days_remaining: 120,
    },
    {
      id: 2,
      shop_id: 2,
      tenant_id: 2,
      start_date: '2023-06-01',
      duration_months: 24,
      monthly_rent: 20000,
      security_deposit: 40000,
      active: true,
      created_at: '2023-06-01T00:00:00Z',
      updated_at: '2023-06-01T00:00:00Z',
      shop_name: 'Fashion Boutique',
      shop_number: 'S-002',
      tenant_name: 'Fatima Rahman',
      tenant_phone: '+8801812345678',
      end_date: '2025-05-31',
      is_expired: false,
      days_remaining: 365,
    },
    {
      id: 3,
      shop_id: 3,
      tenant_id: 3,
      start_date: '2023-01-01',
      duration_months: 12,
      monthly_rent: 12000,
      security_deposit: 24000,
      active: false,
      created_at: '2023-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      shop_name: 'Grocery Store',
      shop_number: 'S-003',
      tenant_name: 'Mohammad Ali',
      tenant_phone: '+8801912345678',
      end_date: '2023-12-31',
      is_expired: true,
      days_remaining: -30,
    },
  ];

  getAllLeases(): Observable<LeaseWithDetails[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        observer.next(this.demoLeases);
        observer.complete();
      });
    }

    return this.http.get<LeaseWithDetails[]>(this.apiUrl);
  }

  getLeaseById(id: number): Observable<LeaseWithDetails> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const lease = this.demoLeases.find(l => l.id === id);
        if (lease) {
          observer.next(lease);
        } else {
          observer.error({ error: 'Lease not found' });
        }
        observer.complete();
      });
    }

    return this.http.get<LeaseWithDetails>(`${this.apiUrl}/${id}`);
  }

  createLease(request: CreateLeaseRequest): Observable<Lease> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const newLease: Lease = {
          id: Math.max(...this.demoLeases.map(l => l.id)) + 1,
          ...request,
          active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        observer.next(newLease);
        observer.complete();
      });
    }

    return this.http.post<Lease>(this.apiUrl, request);
  }

  updateLease(id: number, request: UpdateLeaseRequest): Observable<any> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        observer.next({ message: 'Lease updated successfully' });
        observer.complete();
      });
    }

    return this.http.put(`${this.apiUrl}/${id}`, request);
  }

  deleteLease(id: number): Observable<any> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        observer.next({ message: 'Lease deleted successfully' });
        observer.complete();
      });
    }

    return this.http.delete(`${this.apiUrl}/${id}`);
  }

  getActiveLeases(): Observable<LeaseWithDetails[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const activeLeases = this.demoLeases.filter(l => l.active);
        observer.next(activeLeases);
        observer.complete();
      });
    }

    return this.http.get<LeaseWithDetails[]>(`${this.apiUrl}?active=true`);
  }

  getExpiringLeases(days: number = 30): Observable<LeaseWithDetails[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const expiringLeases = this.demoLeases.filter(l => 
          l.active && l.days_remaining <= days && l.days_remaining > 0
        );
        observer.next(expiringLeases);
        observer.complete();
      });
    }

    return this.http.get<LeaseWithDetails[]>(`${this.apiUrl}?expiring_in_days=${days}`);
  }

  getExpiredLeases(): Observable<LeaseWithDetails[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const expiredLeases = this.demoLeases.filter(l => l.is_expired);
        observer.next(expiredLeases);
        observer.complete();
      });
    }

    return this.http.get<LeaseWithDetails[]>(`${this.apiUrl}?expired=true`);
  }
}