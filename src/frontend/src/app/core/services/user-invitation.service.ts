import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { UserInvitation, InviteUserRequest } from '../models';
import { environment } from '../../../environments/environment';
import { User } from './auth.service';

interface InvitationListResponse {
  invitations: UserInvitation[];
  total: number;
}

@Injectable({
  providedIn: 'root',
})
export class UserInvitationService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/organizations`;

  sendInvitation(orgId: number, req: InviteUserRequest): Observable<UserInvitation> {
    return this.http.post<UserInvitation>(`${this.apiUrl}/${orgId}/invitations`, req);
  }

  getPendingInvitations(
    orgId: number,
    params?: { page?: number; limit?: number }
  ): Observable<InvitationListResponse> {
    let httpParams = new HttpParams();
    if (params?.page) {
      httpParams = httpParams.set('page', params.page.toString());
    }
    if (params?.limit) {
      httpParams = httpParams.set('limit', params.limit.toString());
    }
    return this.http.get<InvitationListResponse>(`${this.apiUrl}/${orgId}/invitations/pending`, {
      params: httpParams,
    });
  }

  revokeInvitation(inviteId: number, orgId: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${orgId}/invitations/${inviteId}`);
  }

  acceptInvitation(token: string, password: string): Observable<User> {
    return this.http.post<User>(`${environment.apiUrl}/invitations/${token}/accept`, {
      password,
    });
  }

  resendInvitation(inviteId: number): Observable<UserInvitation> {
    return this.http.post<UserInvitation>(
      `${environment.apiUrl}/invitations/${inviteId}/resend`,
      {}
    );
  }

  validateInvitationToken(token: string): Observable<UserInvitation> {
    return this.http.get<UserInvitation>(`${environment.apiUrl}/invitations/${token}/validate`);
  }
}
