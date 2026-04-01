import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AuditLog, AuditFilter } from '../models';
import { environment } from '../../../environments/environment';

interface AuditLogListResponse {
  logs: AuditLog[];
  total: number;
}

@Injectable({
  providedIn: 'root',
})
export class AuditLogService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/organizations`;

  getAuditLogs(
    orgId: number,
    filters?: AuditFilter & { page?: number; limit?: number }
  ): Observable<AuditLogListResponse> {
    let httpParams = new HttpParams();

    if (filters?.page) {
      httpParams = httpParams.set('page', filters.page.toString());
    }
    if (filters?.limit) {
      httpParams = httpParams.set('limit', filters.limit.toString());
    }
    if (filters?.action) {
      httpParams = httpParams.set('action', filters.action);
    }
    if (filters?.resourceType) {
      httpParams = httpParams.set('resourceType', filters.resourceType);
    }
    if (filters?.userId) {
      httpParams = httpParams.set('userId', filters.userId.toString());
    }
    if (filters?.startDate) {
      httpParams = httpParams.set('startDate', filters.startDate.toISOString());
    }
    if (filters?.endDate) {
      httpParams = httpParams.set('endDate', filters.endDate.toISOString());
    }

    return this.http.get<AuditLogListResponse>(`${this.apiUrl}/${orgId}/audit-logs`, {
      params: httpParams,
    });
  }

  exportAuditLogs(orgId: number, filters?: AuditFilter): Observable<Blob> {
    let httpParams = new HttpParams();

    if (filters?.action) {
      httpParams = httpParams.set('action', filters.action);
    }
    if (filters?.resourceType) {
      httpParams = httpParams.set('resourceType', filters.resourceType);
    }
    if (filters?.userId) {
      httpParams = httpParams.set('userId', filters.userId.toString());
    }
    if (filters?.startDate) {
      httpParams = httpParams.set('startDate', filters.startDate.toISOString());
    }
    if (filters?.endDate) {
      httpParams = httpParams.set('endDate', filters.endDate.toISOString());
    }

    return this.http.get<Blob>(`${this.apiUrl}/${orgId}/audit-logs/export`, {
      params: httpParams,
      responseType: 'blob' as 'json',
    });
  }
}
