import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { Organization, CreateOrgRequest, UpdateOrgRequest, OrgStats } from '../models';
import { environment } from '../../../environments/environment';
import { User } from './auth.service';

interface OrgListResponse {
  organizations: Organization[];
  total: number;
}

interface OrgUsersResponse {
  users: User[];
  total: number;
}

@Injectable({
  providedIn: 'root',
})
export class OrganizationService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/organizations`;

  // Current organization context
  public currentOrganization$ = new BehaviorSubject<Organization | null>(null);

  getOrganizations(): Observable<OrgListResponse> {
    return this.http.get<OrgListResponse>(this.apiUrl);
  }

  getOrganization(id: number): Observable<Organization> {
    return this.http.get<Organization>(`${this.apiUrl}/${id}`);
  }

  createOrganization(req: CreateOrgRequest): Observable<Organization> {
    return this.http.post<Organization>(this.apiUrl, req);
  }

  updateOrganization(id: number, req: UpdateOrgRequest): Observable<Organization> {
    return this.http.put<Organization>(`${this.apiUrl}/${id}`, req);
  }

  deleteOrganization(id: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }

  getOrganizationUsers(
    orgId: number,
    params?: { page?: number; limit?: number }
  ): Observable<OrgUsersResponse> {
    let httpParams = new HttpParams();
    if (params?.page) {
      httpParams = httpParams.set('page', params.page.toString());
    }
    if (params?.limit) {
      httpParams = httpParams.set('limit', params.limit.toString());
    }
    return this.http.get<OrgUsersResponse>(`${this.apiUrl}/${orgId}/users`, {
      params: httpParams,
    });
  }

  getOrganizationStats(orgId: number): Observable<OrgStats> {
    return this.http.get<OrgStats>(`${this.apiUrl}/${orgId}/stats`);
  }

  getCurrentOrganization(): Organization | null {
    return this.currentOrganization$.value;
  }

  setCurrentOrganization(org: Organization | null): void {
    this.currentOrganization$.next(org);
  }
}
