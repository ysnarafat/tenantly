import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import {
  Unit,
  UnitWithDetails,
  UnitType,
  CreateUnitRequest,
  UpdateUnitRequest,
  UnitListResponse,
} from '../models';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class UnitService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/units`;

  getUnits(params?: {
    property_id?: number;
    building_id?: number;
    unit_type?: UnitType;
    active?: boolean;
  }): Observable<UnitListResponse> {
    let httpParams = new HttpParams();
    if (params?.property_id) {
      httpParams = httpParams.set('property_id', params.property_id.toString());
    }
    if (params?.building_id) {
      httpParams = httpParams.set('building_id', params.building_id.toString());
    }
    if (params?.unit_type) {
      httpParams = httpParams.set('unit_type', params.unit_type);
    }
    if (params?.active !== undefined) {
      httpParams = httpParams.set('active', params.active.toString());
    }
    return this.http.get<UnitListResponse>(this.apiUrl, { params: httpParams });
  }

  getUnitsByBuilding(buildingId: number): Observable<UnitListResponse> {
    return this.http
      .get<{
        data: Unit[];
        meta: { total: number; page: number; page_size: number };
      }>(`${environment.apiUrl}/buildings/${buildingId}/units/list`)
      .pipe(
        map((response) => ({
          units: response.data,
          pagination: {
            current_page: response.meta.page,
            page_size: response.meta.page_size,
            total_items: response.meta.total,
            total_pages: Math.ceil(response.meta.total / response.meta.page_size),
            has_next: response.meta.page * response.meta.page_size < response.meta.total,
            has_prev: response.meta.page > 1,
          },
        }))
      );
  }

  getUnit(id: number): Observable<Unit> {
    return this.http.get<Unit>(`${this.apiUrl}/${id}`);
  }

  getUnitWithDetails(id: number): Observable<UnitWithDetails> {
    return this.http.get<UnitWithDetails>(`${this.apiUrl}/${id}/details`);
  }

  createUnit(unit: CreateUnitRequest): Observable<Unit> {
    return this.http.post<Unit>(this.apiUrl, unit);
  }

  updateUnit(id: number, unit: UpdateUnitRequest): Observable<Unit> {
    return this.http.put<Unit>(`${this.apiUrl}/${id}`, unit);
  }

  deleteUnit(id: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }

  // Bulk operations
  createBulkUnits(units: CreateUnitRequest[]): Observable<Unit[]> {
    return this.http.post<Unit[]>(`${this.apiUrl}/bulk`, units);
  }
}
