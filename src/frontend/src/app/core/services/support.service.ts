import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import {
  CreateSupportCaseRequest,
  SupportCase,
  SupportCaseMessage,
} from '../models/support-case.model';

@Injectable({ providedIn: 'root' })
export class SupportService {
  private readonly http = inject(HttpClient);
  private readonly casesUrl = `${environment.apiUrl}/support/cases`;

  listCases(): Observable<{ cases: SupportCase[]; total: number }> {
    return this.http.get<{ cases: SupportCase[]; total: number }>(this.casesUrl);
  }

  getCase(caseId: number): Observable<SupportCase> {
    return this.http.get<SupportCase>(`${this.casesUrl}/${caseId}`);
  }

  createCase(request: CreateSupportCaseRequest): Observable<SupportCase> {
    return this.http.post<SupportCase>(this.casesUrl, request);
  }

  addMessage(caseId: number, body: string): Observable<SupportCaseMessage> {
    return this.http.post<SupportCaseMessage>(`${this.casesUrl}/${caseId}/messages`, { body });
  }
}
