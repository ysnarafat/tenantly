import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { environment } from '../../../environments/environment';
import { LeaseWithDetails } from '../models/lease.model';

@Injectable({
  providedIn: 'root',
})
export class LeaseService {
  private http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/leases`;

  getActiveLeases(): Observable<LeaseWithDetails[]> {
    return this.http
      .get<{ leases: LeaseWithDetails[] }>(this.apiUrl)
      .pipe(map((res) => res.leases ?? []));
  }
}
