import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { 
  Property, 
  PropertyWithStats, 
  CreatePropertyRequest, 
  UpdatePropertyRequest 
} from '../models';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class PropertyService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/properties`;

  getProperties(params?: { active?: boolean }): Observable<Property[]> {
    let httpParams = new HttpParams();
    if (params?.active !== undefined) {
      httpParams = httpParams.set('active', params.active.toString());
    }
    return this.http.get<Property[]>(this.apiUrl, { params: httpParams });
  }

  getProperty(id: number): Observable<Property> {
    return this.http.get<Property>(`${this.apiUrl}/${id}`);
  }

  getPropertyWithStats(id: number): Observable<PropertyWithStats> {
    return this.http.get<PropertyWithStats>(`${this.apiUrl}/${id}/stats`);
  }

  createProperty(property: CreatePropertyRequest): Observable<Property> {
    return this.http.post<Property>(this.apiUrl, property);
  }

  updateProperty(id: number, property: UpdatePropertyRequest): Observable<Property> {
    return this.http.put<Property>(`${this.apiUrl}/${id}`, property);
  }

  deleteProperty(id: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}
