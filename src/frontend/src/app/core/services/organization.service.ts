import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, BehaviorSubject } from 'rxjs';
import { tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import {
  Organization,
  CreateOrgRequest,
  UpdateOrgRequest,
  OrgStats,
  UserOrganization,
  SetOrganizationRequest,
  SetOrganizationResponse,
} from '../models/organization.model';
import { User } from './auth.service';
import { authStorage } from '../../shared/utils/auth-storage.utils';

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
  private authApiUrl = `${environment.apiUrl}/auth`;
  private readonly ORG_ID_KEY = 'tenantly_current_org_id';

  public currentOrganization$ = new BehaviorSubject<Organization | null>(null);

  // --- Organization CRUD ---

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

  // --- Current Organization Context ---

  getCurrentOrganization(): Organization | null {
    return this.currentOrganization$.value;
  }

  setCurrentOrganization(org: Organization | null): void {
    this.currentOrganization$.next(org);
  }

  // --- Organization Switching (multi-org support) ---

  setOrganization(organizationId: number): Observable<SetOrganizationResponse> {
    const request: SetOrganizationRequest = { organization_id: organizationId };
    return this.http
      .post<SetOrganizationResponse>(`${this.authApiUrl}/set-organization`, request)
      .pipe(
        tap((response) => {
          authStorage.setItem(this.ORG_ID_KEY, organizationId.toString());
          authStorage.setItem(
            'tenantly_current_org',
            JSON.stringify(response.organization.organization)
          );
          this.currentOrganization$.next(response.organization.organization);
        })
      );
  }

  getCurrentOrganizationId(): number | null {
    const orgId = authStorage.getItem(this.ORG_ID_KEY);
    return orgId ? parseInt(orgId, 10) : null;
  }

  storeOrganizationContext(orgId: number, organization: UserOrganization): void {
    authStorage.setItem(this.ORG_ID_KEY, orgId.toString());
    authStorage.setItem('tenantly_current_org', JSON.stringify(organization.organization));
  }

  restoreOrganizationContext(): void {
    const orgData = authStorage.getItem('tenantly_current_org');
    if (orgData) {
      try {
        const org: Organization = JSON.parse(orgData);
        this.currentOrganization$.next(org);
      } catch (e) {
        console.error('Failed to restore organization context:', e);
      }
    }
  }

  clearOrganizationContext(): void {
    authStorage.removeItem(this.ORG_ID_KEY);
    authStorage.removeItem('tenantly_current_org');
    this.currentOrganization$.next(null);
  }
}
