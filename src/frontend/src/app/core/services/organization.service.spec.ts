import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { OrganizationService } from './organization.service';
import { Organization, CreateOrgRequest, UpdateOrgRequest, OrgStats } from '../models';

describe('OrganizationService', () => {
  let service: OrganizationService;
  let httpMock: HttpTestingController;
  const apiUrl = '/api/v1/organizations';

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [OrganizationService],
    });
    service = TestBed.inject(OrganizationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('Initialization', () => {
    it('should be created', () => {
      expect(service).toBeTruthy();
    });

    it('should initialize currentOrganization$ as null', () => {
      expect(service.currentOrganization$.value).toBeNull();
    });
  });

  describe('getOrganizations', () => {
    it('should fetch all organizations', () => {
      const mockOrgs: Organization[] = [
        {
          id: 1,
          name: 'Org 1',
          slug: 'org-1',
          subscriptionTier: 'basic',
          maxUsers: 10,
          active: true,
          createdAt: new Date(),
          updatedAt: new Date(),
        },
        {
          id: 2,
          name: 'Org 2',
          slug: 'org-2',
          subscriptionTier: 'professional',
          maxUsers: 50,
          active: true,
          createdAt: new Date(),
          updatedAt: new Date(),
        },
      ];

      service.getOrganizations().subscribe((result) => {
        expect(result.organizations).toEqual(mockOrgs);
        expect(result.total).toBe(2);
      });

      const req = httpMock.expectOne(apiUrl);
      expect(req.request.method).toBe('GET');
      req.flush({ organizations: mockOrgs, total: 2 });
    });

    it('should handle empty organizations list', () => {
      service.getOrganizations().subscribe((result) => {
        expect(result.organizations).toEqual([]);
        expect(result.total).toBe(0);
      });

      const req = httpMock.expectOne(apiUrl);
      req.flush({ organizations: [], total: 0 });
    });

    it('should handle error response', () => {
      service.getOrganizations().subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(500);
        }
      );

      const req = httpMock.expectOne(apiUrl);
      req.flush('Server error', { status: 500, statusText: 'Internal Server Error' });
    });
  });

  describe('getOrganization', () => {
    it('should fetch a single organization by id', () => {
      const mockOrg: Organization = {
        id: 1,
        name: 'Test Org',
        slug: 'test-org',
        subscriptionTier: 'professional',
        maxUsers: 50,
        active: true,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      service.getOrganization(1).subscribe((result) => {
        expect(result).toEqual(mockOrg);
      });

      const req = httpMock.expectOne(`${apiUrl}/1`);
      expect(req.request.method).toBe('GET');
      req.flush(mockOrg);
    });

    it('should handle 404 not found error', () => {
      service.getOrganization(999).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(404);
        }
      );

      const req = httpMock.expectOne(`${apiUrl}/999`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });
  });

  describe('createOrganization', () => {
    it('should create a new organization', () => {
      const createReq: CreateOrgRequest = {
        name: 'New Org',
        slug: 'new-org',
        subscriptionTier: 'professional',
        maxUsers: 50,
      };

      const mockOrg: Organization = {
        id: 1,
        ...createReq,
        active: true,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      service.createOrganization(createReq).subscribe((result) => {
        expect(result.id).toBe(1);
        expect(result.name).toBe('New Org');
        expect(result.slug).toBe('new-org');
        expect(result.subscriptionTier).toBe('professional');
        expect(result.active).toBe(true);
      });

      const req = httpMock.expectOne(apiUrl);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(createReq);
      req.flush(mockOrg);
    });

    it('should handle validation error on create', () => {
      const invalidReq: CreateOrgRequest = {
        name: '',
        slug: '',
        subscriptionTier: 'basic',
        maxUsers: 0,
      };

      service.createOrganization(invalidReq).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(400);
        }
      );

      const req = httpMock.expectOne(apiUrl);
      req.flush('Validation failed', { status: 400, statusText: 'Bad Request' });
    });

    it('should send correct POST body structure', () => {
      const createReq: CreateOrgRequest = {
        name: 'Test',
        slug: 'test',
        subscriptionTier: 'enterprise',
        maxUsers: 1000,
      };

      service.createOrganization(createReq).subscribe();

      const req = httpMock.expectOne(apiUrl);
      expect(req.request.body).toEqual({
        name: 'Test',
        slug: 'test',
        subscriptionTier: 'enterprise',
        maxUsers: 1000,
      });
      req.flush({} as Organization);
    });
  });

  describe('updateOrganization', () => {
    it('should update an organization', () => {
      const updateReq: UpdateOrgRequest = {
        name: 'Updated Org',
        maxUsers: 100,
      };

      const mockOrg: Organization = {
        id: 1,
        name: 'Updated Org',
        slug: 'new-org',
        subscriptionTier: 'professional',
        maxUsers: 100,
        active: true,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      service.updateOrganization(1, updateReq).subscribe((result) => {
        expect(result.name).toBe('Updated Org');
        expect(result.maxUsers).toBe(100);
      });

      const req = httpMock.expectOne(`${apiUrl}/1`);
      expect(req.request.method).toBe('PUT');
      expect(req.request.body).toEqual(updateReq);
      req.flush(mockOrg);
    });

    it('should support partial updates', () => {
      const updateReq: UpdateOrgRequest = {
        maxUsers: 200,
      };

      service.updateOrganization(1, updateReq).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/1`);
      expect(req.request.body).toEqual({ maxUsers: 200 });
      req.flush({} as Organization);
    });

    it('should handle update conflict', () => {
      const updateReq: UpdateOrgRequest = { name: 'Updated' };

      service.updateOrganization(1, updateReq).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(409);
        }
      );

      const req = httpMock.expectOne(`${apiUrl}/1`);
      req.flush('Conflict', { status: 409, statusText: 'Conflict' });
    });
  });

  describe('deleteOrganization', () => {
    it('should delete an organization', () => {
      service.deleteOrganization(1).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/1`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });

    it('should handle delete of non-existent org', () => {
      service.deleteOrganization(999).subscribe(
        () => fail('should have failed'),
        (error) => {
          expect(error.status).toBe(404);
        }
      );

      const req = httpMock.expectOne(`${apiUrl}/999`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });
  });

  describe('getOrganizationUsers', () => {
    it('should fetch organization users', () => {
      const mockUsers = [
        {
          id: 1,
          username: 'user1',
          email: 'user1@test.com',
          role: 'Admin',
          active: true,
          created_at: '',
          updated_at: '',
        },
        {
          id: 2,
          username: 'user2',
          email: 'user2@test.com',
          role: 'PropertyManager',
          active: true,
          created_at: '',
          updated_at: '',
        },
      ];

      service.getOrganizationUsers(1).subscribe((result) => {
        expect(result.users.length).toBe(2);
        expect(result.total).toBe(2);
      });

      const req = httpMock.expectOne(`${apiUrl}/1/users`);
      expect(req.request.method).toBe('GET');
      req.flush({ users: mockUsers, total: 2 });
    });

    it('should support pagination parameters', () => {
      service.getOrganizationUsers(1, { page: 2, limit: 25 }).subscribe();

      const req = httpMock.expectOne((r) => r.url.includes(`${apiUrl}/1/users`));
      expect(req.request.params.get('page')).toBe('2');
      expect(req.request.params.get('limit')).toBe('25');
      req.flush({ users: [], total: 0 });
    });

    it('should handle empty users list', () => {
      service.getOrganizationUsers(1).subscribe((result) => {
        expect(result.users).toEqual([]);
        expect(result.total).toBe(0);
      });

      const req = httpMock.expectOne(`${apiUrl}/1/users`);
      req.flush({ users: [], total: 0 });
    });
  });

  describe('getOrganizationStats', () => {
    it('should fetch organization statistics', () => {
      const mockStats: OrgStats = {
        totalUsers: 50,
        usersByRole: {
          SUPER_ADMIN: 1,
          ORG_ADMIN: 2,
          Admin: 5,
          PropertyManager: 20,
          Accountant: 22,
        },
        pendingInvitations: 3,
        activeProperties: 10,
      };

      service.getOrganizationStats(1).subscribe((result) => {
        expect(result.totalUsers).toBe(50);
        expect(result.pendingInvitations).toBe(3);
        expect(result.activeProperties).toBe(10);
      });

      const req = httpMock.expectOne(`${apiUrl}/1/stats`);
      expect(req.request.method).toBe('GET');
      req.flush(mockStats);
    });
  });

  describe('getCurrentOrganization', () => {
    it('should return null when no organization is set', () => {
      expect(service.getCurrentOrganization()).toBeNull();
    });

    it('should return current organization when set', () => {
      const mockOrg: Organization = {
        id: 1,
        name: 'Test Org',
        slug: 'test-org',
        subscriptionTier: 'basic',
        maxUsers: 10,
        active: true,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      service.setCurrentOrganization(mockOrg);
      expect(service.getCurrentOrganization()).toEqual(mockOrg);
    });
  });

  describe('setCurrentOrganization', () => {
    it('should update currentOrganization$ BehaviorSubject', (done) => {
      const mockOrg: Organization = {
        id: 1,
        name: 'Test Org',
        slug: 'test-org',
        subscriptionTier: 'basic',
        maxUsers: 10,
        active: true,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      let emissionCount = 0;
      service.currentOrganization$.subscribe((org) => {
        emissionCount++;
        if (emissionCount === 2) {
          // First emission is null, second is the org
          expect(org?.id).toBe(1);
          expect(org?.name).toBe('Test Org');
          done();
        }
      });

      service.setCurrentOrganization(mockOrg);
    });

    it('should allow clearing current organization', (done) => {
      const mockOrg: Organization = {
        id: 1,
        name: 'Test Org',
        slug: 'test-org',
        subscriptionTier: 'basic',
        maxUsers: 10,
        active: true,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      service.setCurrentOrganization(mockOrg);
      service.setCurrentOrganization(null);

      service.currentOrganization$.subscribe((org) => {
        expect(org).toBeNull();
        done();
      });
    });
  });

  describe('Multiple concurrent requests', () => {
    it('should handle multiple concurrent requests', () => {
      const mockOrg: Organization = {
        id: 1,
        name: 'Test',
        slug: 'test',
        subscriptionTier: 'basic',
        maxUsers: 10,
        active: true,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      let completedRequests = 0;

      service.getOrganizations().subscribe(() => completedRequests++);
      service.getOrganization(1).subscribe(() => completedRequests++);
      service.getOrganizationUsers(1).subscribe(() => completedRequests++);

      const requests = httpMock.match(() => true);
      expect(requests.length).toBe(3);

      requests[0].flush({ organizations: [mockOrg], total: 1 });
      requests[1].flush(mockOrg);
      requests[2].flush({ users: [], total: 0 });

      expect(completedRequests).toBe(3);
    });
  });
});
