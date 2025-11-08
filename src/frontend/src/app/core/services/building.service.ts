import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { 
  Building, 
  BuildingWithStats, 
  CreateBuildingRequest, 
  UpdateBuildingRequest 
} from '../models';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class BuildingService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/buildings`;

  getBuildings(params?: { property_id?: number; active?: boolean }): Observable<Building[]> {
    let httpParams = new HttpParams();
    if (params?.property_id) {
      httpParams = httpParams.set('property_id', params.property_id.toString());
    }
    if (params?.active !== undefined) {
      httpParams = httpParams.set('active', params.active.toString());
    }
    return this.http.get<Building[]>(this.apiUrl, { params: httpParams });
  }

  getBuildingsByProperty(propertyId: number): Observable<Building[]> {
    return this.http.get<Building[]>(`${environment.apiUrl}/properties/${propertyId}/buildings`);
  }

  getBuilding(id: number): Observable<Building> {
    return this.http.get<Building>(`${this.apiUrl}/${id}`);
  }

  getBuildingWithStats(id: number): Observable<BuildingWithStats> {
    return this.http.get<BuildingWithStats>(`${this.apiUrl}/${id}/stats`);
  }

  createBuilding(building: CreateBuildingRequest): Observable<Building> {
    return this.http.post<Building>(this.apiUrl, building);
  }

  updateBuilding(id: number, building: UpdateBuildingRequest): Observable<Building> {
    return this.http.put<Building>(`${this.apiUrl}/${id}`, building);
  }

  deleteBuilding(id: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}
