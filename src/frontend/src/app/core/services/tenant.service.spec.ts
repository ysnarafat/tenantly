import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TenantService } from './tenant.service';
import { environment } from '../../../environments/environment';
import { CreateTenantRequest, Tenant } from '../models/tenant.model';

describe('TenantService', () => {
  let service: TenantService;
  let httpMock: HttpTestingController;

  const mockTenant: Tenant = {
    id: 1,
    name: 'John Doe',
    tenant_type: 'Individual',
    email: 'john@example.com',
    phone_number: '1234567890',
    nid_last_four: 'D123',
    address: '123 Main St',
    active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  const createRequest: CreateTenantRequest = {
    name: 'John Doe',
    tenant_type: 'Individual',
    email: 'john@example.com',
    phone_number: '1234567890',
    nid_number: 'NID123',
    address: '123 Main St',
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [TenantService],
    });
    service = TestBed.inject(TenantService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should create a tenant', () => {
    service.createTenant(createRequest).subscribe((tenant) => {
      expect(tenant).toEqual(mockTenant);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/tenants`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(createRequest);
    req.flush(mockTenant);
  });

  it('should handle error when creating tenant fails', () => {
    const errorMsg = 'Email already exists';

    service.createTenant(createRequest).subscribe({
      next: () => fail('should have failed with 409 error'),
      error: (error) => {
        expect(error.status).toBe(409);
        expect(error.error).toBe(errorMsg);
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/tenants`);
    expect(req.request.method).toBe('POST');
    req.flush(errorMsg, { status: 409, statusText: 'Conflict' });
  });
});
