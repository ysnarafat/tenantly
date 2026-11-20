import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { BillingRun, CreateBillingRunRequest } from '../models/billing-run.model';

@Injectable({ providedIn: 'root' })
export class BillingRunService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/billing-runs`;

  getByPeriod(buildingId: number, month: number, year: number): Observable<BillingRun> {
    const params = new HttpParams()
      .set('building_id', buildingId)
      .set('month', month)
      .set('year', year);
    return this.http.get<BillingRun>(`${this.apiUrl}/period`, { params });
  }

  create(request: CreateBillingRunRequest): Observable<BillingRun> {
    return this.http.post<BillingRun>(this.apiUrl, request);
  }

  updateElectricity(
    runId: number,
    itemId: number,
    meterNumber: string,
    electricityAmount: number
  ): Observable<BillingRun> {
    return this.http.patch<BillingRun>(`${this.apiUrl}/${runId}/items/${itemId}/electricity`, {
      meter_number: meterNumber,
      electricity_amount: electricityAmount,
    });
  }

  finalize(runId: number): Observable<BillingRun> {
    return this.http.post<BillingRun>(`${this.apiUrl}/${runId}/finalize`, {});
  }
}
