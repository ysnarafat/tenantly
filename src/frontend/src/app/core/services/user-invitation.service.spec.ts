import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { UserInvitationService } from './user-invitation.service';
import { UserInvitation, InviteUserRequest } from '../models';

describe('UserInvitationService', () => {
  let service: UserInvitationService;
  let httpMock: HttpTestingController;
  const apiUrl = '/api/v1/organizations';
  const invitationsUrl = '/api/v1/invitations';

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [UserInvitationService],
    });
    service = TestBed.inject(UserInvitationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('Initialization', () => {
    it('should be created', () => {
      expect(service).toBeTruthy();
    });
  });

  describe('sendInvitation', () => {
    it('should send an invitation', () => {
      const orgId = 1;
      const inviteReq: InviteUserRequest = {
        email: 'user@example.com',
        firstName: 'John',
        lastName: 'Doe',
        role: 'Admin',
      };

      const mockInvitation: UserInvitation = {
        id: 1,
        organizationId: orgId,
        email: 'user@example.com',
        role: 'Admin',
        token: 'mock-token-123',
        expiresAt: new Date(),
        createdAt: new Date(),
      };

      service.sendInvitation(orgId, inviteReq).subscribe((result) => {
        expect(result.email).toBe('user@example.com');
        expect(result.role).toBe('Admin');
        expect(result.token).toBe('mock-token-123');
      });

      const req = httpMock.expectOne(`${apiUrl}/${orgId}/invitations`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(inviteReq);
      req.flush(mockInvitation);
    });

    it('should send invitation with all user roles', () => {
      const orgId = 1;
      const roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'PropertyManager', 'Accountant'];

      roles.forEach((role) => {
        const inviteReq: InviteUserRequest = {
          email: `user-${role}@example.com`,
          firstName: 'Test',
          lastName: 'User',
          role: role as unknown as InviteUserRequest['role'],
        };

        service.sendInvitation(orgId, inviteReq).subscribe();

        const req = httpMock.expectOne(`${apiUrl}/${orgId}/invitations`);
        expect(req.request.body.role).toBe(role);
        req.flush({} as UserInvitation);
      });
    });

    it('should handle duplicate email error', () => {
      const orgId = 1;
      const inviteReq: InviteUserRequest = {
        email: 'existing@example.com',
        firstName: 'John',
        lastName: 'Doe',
        role: 'Admin',
      };

      service.sendInvitation(orgId, inviteReq).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(409);
        }
      );

      const req = httpMock.expectOne(`${apiUrl}/${orgId}/invitations`);
      req.flush('User already exists or already invited', {
        status: 409,
        statusText: 'Conflict',
      });
    });

    it('should handle invalid email error', () => {
      const orgId = 1;
      const inviteReq: InviteUserRequest = {
        email: 'invalid-email',
        firstName: 'John',
        lastName: 'Doe',
        role: 'Admin',
      };

      service.sendInvitation(orgId, inviteReq).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(400);
        }
      );

      const req = httpMock.expectOne(`${apiUrl}/${orgId}/invitations`);
      req.flush('Invalid email format', { status: 400, statusText: 'Bad Request' });
    });
  });

  describe('getPendingInvitations', () => {
    it('should fetch pending invitations', () => {
      const orgId = 1;
      const mockInvitations: UserInvitation[] = [
        {
          id: 1,
          organizationId: orgId,
          email: 'user1@example.com',
          role: 'Admin',
          token: 'token-1',
          expiresAt: new Date(),
          createdAt: new Date(),
        },
        {
          id: 2,
          organizationId: orgId,
          email: 'user2@example.com',
          role: 'PropertyManager',
          token: 'token-2',
          expiresAt: new Date(),
          createdAt: new Date(),
        },
      ];

      service.getPendingInvitations(orgId).subscribe((result) => {
        expect(result.invitations.length).toBe(2);
        expect(result.total).toBe(2);
        expect(result.invitations[0].email).toBe('user1@example.com');
      });

      const req = httpMock.expectOne(`${apiUrl}/${orgId}/invitations/pending`);
      expect(req.request.method).toBe('GET');
      req.flush({ invitations: mockInvitations, total: 2 });
    });

    it('should support pagination parameters', () => {
      const orgId = 1;
      service.getPendingInvitations(orgId, { page: 2, limit: 10 }).subscribe();

      const req = httpMock.expectOne((r) =>
        r.url.includes(`${apiUrl}/${orgId}/invitations/pending`)
      );
      expect(req.request.params.get('page')).toBe('2');
      expect(req.request.params.get('limit')).toBe('10');
      req.flush({ invitations: [], total: 0 });
    });

    it('should handle empty invitations list', () => {
      const orgId = 1;
      service.getPendingInvitations(orgId).subscribe((result) => {
        expect(result.invitations).toEqual([]);
        expect(result.total).toBe(0);
      });

      const req = httpMock.expectOne(`${apiUrl}/${orgId}/invitations/pending`);
      req.flush({ invitations: [], total: 0 });
    });

    it('should include only page parameter when limit is not provided', () => {
      const orgId = 1;
      service.getPendingInvitations(orgId, { page: 1 }).subscribe();

      const req = httpMock.expectOne((r) =>
        r.url.includes(`${apiUrl}/${orgId}/invitations/pending`)
      );
      expect(req.request.params.get('page')).toBe('1');
      expect(req.request.params.get('limit')).toBeNull();
      req.flush({ invitations: [], total: 0 });
    });
  });

  describe('revokeInvitation', () => {
    it('should revoke an invitation', () => {
      service.revokeInvitation(1).subscribe();

      const req = httpMock.expectOne(`${invitationsUrl}/1`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });

    it('should handle revoking non-existent invitation', () => {
      service.revokeInvitation(999).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(404);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/999`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });

    it('should handle revoking already accepted invitation', () => {
      service.revokeInvitation(1).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(400);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/1`);
      req.flush('Cannot revoke accepted invitation', {
        status: 400,
        statusText: 'Bad Request',
      });
    });
  });

  describe('acceptInvitation', () => {
    it('should accept an invitation with password', () => {
      const token = 'valid-token';
      const password = 'SecurePassword123!';

      const mockUser = {
        id: 1,
        username: 'newuser',
        email: 'user@example.com',
        role: 'Admin',
        active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      service.acceptInvitation(token, password).subscribe((result) => {
        expect(result.username).toBe('newuser');
        expect(result.email).toBe('user@example.com');
      });

      const req = httpMock.expectOne(`${invitationsUrl}/${token}/accept`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ password });
      req.flush(mockUser);
    });

    it('should handle invalid token', () => {
      service.acceptInvitation('invalid-token', 'password').subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(404);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/invalid-token/accept`);
      req.flush('Invalid or expired token', { status: 404, statusText: 'Not Found' });
    });

    it('should handle weak password', () => {
      service.acceptInvitation('token', 'weak').subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(400);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/token/accept`);
      req.flush('Password does not meet requirements', {
        status: 400,
        statusText: 'Bad Request',
      });
    });

    it('should handle expired token', () => {
      service.acceptInvitation('expired-token', 'password').subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(410);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/expired-token/accept`);
      req.flush('Token has expired', { status: 410, statusText: 'Gone' });
    });
  });

  describe('resendInvitation', () => {
    it('should resend an invitation', () => {
      const mockInvitation: UserInvitation = {
        id: 1,
        organizationId: 1,
        email: 'user@example.com',
        role: 'Admin',
        token: 'new-token',
        expiresAt: new Date(),
        createdAt: new Date(),
      };

      service.resendInvitation(1).subscribe((result) => {
        expect(result.token).toBe('new-token');
        expect(result.email).toBe('user@example.com');
      });

      const req = httpMock.expectOne(`${invitationsUrl}/1/resend`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({});
      req.flush(mockInvitation);
    });

    it('should handle resending non-existent invitation', () => {
      service.resendInvitation(999).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(404);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/999/resend`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });

    it('should handle resending already accepted invitation', () => {
      service.resendInvitation(1).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(400);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/1/resend`);
      req.flush('Cannot resend already accepted invitation', {
        status: 400,
        statusText: 'Bad Request',
      });
    });
  });

  describe('validateInvitationToken', () => {
    it('should validate an invitation token', () => {
      const token = 'valid-token';
      const mockInvitation: UserInvitation = {
        id: 1,
        organizationId: 1,
        email: 'user@example.com',
        role: 'Admin',
        token: token,
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000), // 7 days from now
        createdAt: new Date(),
      };

      service.validateInvitationToken(token).subscribe((result) => {
        expect(result.token).toBe(token);
        expect(result.email).toBe('user@example.com');
      });

      const req = httpMock.expectOne(`${invitationsUrl}/${token}/validate`);
      expect(req.request.method).toBe('GET');
      req.flush(mockInvitation);
    });

    it('should handle invalid token validation', () => {
      service.validateInvitationToken('invalid-token').subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(404);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/invalid-token/validate`);
      req.flush('Invalid token', { status: 404, statusText: 'Not Found' });
    });

    it('should handle expired token validation', () => {
      service.validateInvitationToken('expired-token').subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(410);
        }
      );

      const req = httpMock.expectOne(`${invitationsUrl}/expired-token/validate`);
      req.flush('Token has expired', { status: 410, statusText: 'Gone' });
    });
  });

  describe('Multiple concurrent requests', () => {
    it('should handle multiple concurrent invitation operations', () => {
      const orgId = 1;
      const inviteReq: InviteUserRequest = {
        email: 'test@example.com',
        firstName: 'Test',
        lastName: 'User',
        role: 'Admin',
      };

      let completedRequests = 0;

      service.sendInvitation(orgId, inviteReq).subscribe(() => completedRequests++);
      service.getPendingInvitations(orgId).subscribe(() => completedRequests++);
      service.validateInvitationToken('token').subscribe(() => completedRequests++);

      const requests = httpMock.match(() => true);
      expect(requests.length).toBe(3);

      requests[0].flush({} as UserInvitation);
      requests[1].flush({ invitations: [], total: 0 });
      requests[2].flush({} as UserInvitation);

      expect(completedRequests).toBe(3);
    });
  });

  describe('Error handling', () => {
    it('should handle network errors gracefully', () => {
      const orgId = 1;
      const inviteReq: InviteUserRequest = {
        email: 'test@example.com',
        firstName: 'Test',
        lastName: 'User',
        role: 'Admin',
      };

      service.sendInvitation(orgId, inviteReq).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(0);
        }
      );

      const req = httpMock.expectOne(`${apiUrl}/${orgId}/invitations`);
      req.error(new ErrorEvent('Network error'), {
        status: 0,
        statusText: 'Unknown Error',
      });
    });

    it('should handle timeout errors', () => {
      service.getPendingInvitations(1).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(0);
        }
      );

      const req = httpMock.expectOne((r) => r.url.includes('invitations/pending'));
      req.flush('Request timeout', { status: 0, statusText: 'Unknown Error' });
    });
  });
});
