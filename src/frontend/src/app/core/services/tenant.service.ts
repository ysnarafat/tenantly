import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { CreateTenantRequest, Tenant, TenantListResponse } from '../models/tenant.model';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class TenantService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/tenants`;

  createTenant(tenant: CreateTenantRequest): Observable<Tenant> {
    return this.http.post<Tenant>(this.apiUrl, tenant);
  }

  getAllTenants(page = 1, pageSize = 10): Observable<TenantListResponse> {
    return this.http.get<TenantListResponse>(this.apiUrl, {
      params: {
        page: page.toString(),
        page_size: pageSize.toString(),
      },
    });
  }
}
