import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { UserInvitationService } from './user-invitation.service';

describe('UserInvitationService', () => {
  let service: UserInvitationService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [UserInvitationService],
    });
    service = TestBed.inject(UserInvitationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('creates an organization invitation', () => {
    const request = {
      email: 'admin@example.com',
      firstName: 'Ada',
      lastName: 'Lovelace',
      role: 'ORG_ADMIN' as const,
    };

    service.sendInvitation(12, request).subscribe();

    const httpRequest = httpMock.expectOne((request) =>
      request.url.endsWith('/api/v1/organizations/12/invitations')
    );
    expect(httpRequest.request.method).toBe('POST');
    expect(httpRequest.request.body).toEqual(request);
    httpRequest.flush({});
  });

  it('loads pending invitations from the organization endpoint', () => {
    service.getPendingInvitations(12).subscribe((response) => {
      expect(response.total).toBe(0);
      expect(response.invitations).toEqual([]);
    });

    const httpRequest = httpMock.expectOne((request) =>
      request.url.endsWith('/api/v1/organizations/12/invitations')
    );
    expect(httpRequest.request.method).toBe('GET');
    httpRequest.flush({ invitations: [], total: 0 });
  });

  it('revokes an invitation within its organization', () => {
    service.revokeInvitation(34, 12).subscribe();

    const httpRequest = httpMock.expectOne((request) =>
      request.url.endsWith('/api/v1/organizations/12/invitations/34')
    );
    expect(httpRequest.request.method).toBe('DELETE');
    httpRequest.flush(null);
  });

  it('validates a token using the public validation endpoint', () => {
    service.validateInvitationToken('one-time-token').subscribe((invitation) => {
      expect(invitation.email).toBe('admin@example.com');
      expect(invitation.organizationName).toBe('Example Org');
      expect(invitation.token).toBe('one-time-token');
    });

    const httpRequest = httpMock.expectOne(
      (request) =>
        request.url.endsWith('/api/v1/invitations/validate') &&
        request.params.get('token') === 'one-time-token'
    );
    expect(httpRequest.request.method).toBe('GET');
    httpRequest.flush({
      invitation: {
        id: 8,
        organization_id: 12,
        organization_name: 'Example Org',
        email: 'admin@example.com',
        role: 'ORG_ADMIN',
        expires_at: '2026-11-15T00:00:00Z',
      },
    });
  });

  it('registers the invited admin and sends credentials with the request', () => {
    const request = {
      email: 'admin@example.com',
      first_name: 'Ada',
      last_name: 'Lovelace',
      password: 'StrongPass123!',
      invitation_token: 'one-time-token',
    };

    service.registerWithInvitation(request).subscribe();

    const httpRequest = httpMock.expectOne((request) =>
      request.url.endsWith('/api/v1/auth/register-with-invitation')
    );
    expect(httpRequest.request.method).toBe('POST');
    expect(httpRequest.request.body).toEqual(request);
    expect(httpRequest.request.withCredentials).toBeTrue();
    httpRequest.flush({ token: 'access-token' });
  });
});
