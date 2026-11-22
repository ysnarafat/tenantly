import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { PublicPlan } from '../models/plan.model';

@Injectable({ providedIn: 'root' })
export class PlanService {
  private readonly http = inject(HttpClient);

  getPublicPlans(): Observable<{ plans: PublicPlan[] }> {
    return this.http.get<{ plans: PublicPlan[] }>(`${environment.apiUrl}/public/plans`);
  }
}
