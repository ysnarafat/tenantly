import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { map, Observable } from 'rxjs';
import { UserInvitation, InviteUserRequest } from '../models';
import { environment } from '../../../environments/environment';
import { LoginResponse } from './auth.service';

interface InvitationListResponse {
  invitations: UserInvitation[];
  total: number;
}

export interface RegisterWithInvitationRequest {
  email: string;
  first_name: string;
  last_name: string;
  password: string;
  invitation_token: string;
}

interface InvitationValidationResponse {
  invitation: {
    id: number;
    organization_id: number;
    organization_name: string;
    email: string;
    role: UserInvitation['role'];
    expires_at: string;
  };
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

  getPendingInvitations(orgId: number): Observable<InvitationListResponse> {
    return this.http.get<InvitationListResponse>(`${this.apiUrl}/${orgId}/invitations`);
  }

  revokeInvitation(inviteId: number, orgId: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${orgId}/invitations/${inviteId}`);
  }

  validateInvitationToken(token: string): Observable<UserInvitation> {
    const params = new HttpParams().set('token', token);
    return this.http
      .get<InvitationValidationResponse>(`${environment.apiUrl}/invitations/validate`, { params })
      .pipe(
        map(({ invitation }) => ({
          id: invitation.id,
          organizationId: invitation.organization_id,
          organizationName: invitation.organization_name,
          email: invitation.email,
          role: invitation.role,
          token,
          expiresAt: new Date(invitation.expires_at),
          createdAt: new Date(),
        }))
      );
  }

  registerWithInvitation(request: RegisterWithInvitationRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(
      `${environment.apiUrl}/auth/register-with-invitation`,
      request,
      { withCredentials: true }
    );
  }
}
