import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { PaymentService, PaymentFilters } from './payment.service';
import {
  Payment,
  PaymentWithDetails,
  PaymentListResponse,
  CreatePaymentRequest,
  UpdatePaymentRequest,
  DashboardSummary,
} from '../models/payment.model';

describe('PaymentService', () => {
  let service: PaymentService;
  let httpMock: HttpTestingController;

  const apiUrl = '/api/v1/payments';
  const baseUrl = '/api/v1';

  // ── Shared fixtures ──────────────────────────────────────────────────────────

  const mockPayment: PaymentWithDetails = {
    id: 1,
    unit_id: 10,
    tenant_id: 5,
    building_id: 2,
    property_id: 3,
    month: 5,
    year: 2026,
    amount_due: 15000,
    amount_paid: 15000,
    status: 'Paid',
    payment_method: 'bKash',
    payment_date: '2026-05-05',
    notes: 'On time',
    receipt_number: 'RCP-001',
    due_date: '2026-05-07',
    created_at: '2026-05-01T00:00:00Z',
    updated_at: '2026-05-05T00:00:00Z',
    property_name: 'Sunrise Apartments',
    building_name: 'Block A',
    building_code: 'BLK-A',
    unit_number: '3B',
    unit_type: 'Residential',
    tenant_name: 'Rahim Uddin',
  };

  const mockListResponse: PaymentListResponse = {
    payments: [mockPayment],
    total: 1,
    page: 1,
    page_size: 20,
    total_pages: 1,
  };

  const mockCreateRequest: CreatePaymentRequest = {
    unit_id: 10,
    tenant_id: 5,
    building_id: 2,
    property_id: 3,
    month: 6,
    year: 2026,
    amount_due: 15000,
    due_date: '2026-06-07',
  };

  const mockUpdateRequest: UpdatePaymentRequest = {
    amount_paid: 15000,
    status: 'Paid',
    payment_method: 'Cash',
    payment_date: '2026-05-10',
    receipt_number: 'RCP-002',
    notes: 'Paid in full',
  };

  const mockSummary: DashboardSummary = {
    total_due: 100000,
    total_paid: 85000,
    total_pending: 10000,
    total_overdue: 5000,
    collection_rate: 85,
    property_count: 4,
    building_count: 8,
    unit_count: 40,
    tenant_count: 38,
  };

  // ── Setup / teardown ─────────────────────────────────────────────────────────

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [PaymentService],
    });
    service = TestBed.inject(PaymentService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  // ── Initialization ───────────────────────────────────────────────────────────

  describe('Initialization', () => {
    it('should be created', () => {
      expect(service).toBeTruthy();
    });
  });

  // ── getPayments ──────────────────────────────────────────────────────────────

  describe('getPayments', () => {
    it('should GET /payments with no query params when called without filters', () => {
      service.getPayments().subscribe((res) => {
        expect(res).toEqual(mockListResponse);
      });

      const req = httpMock.expectOne(apiUrl);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.keys()).toEqual([]);
      req.flush(mockListResponse);
    });

    it('should include status, month, and year params when those filters are provided', () => {
      const filters: PaymentFilters = { status: 'Paid', month: 6, year: 2026 };

      service.getPayments(filters).subscribe();

      const req = httpMock.expectOne((r) => r.url === apiUrl);
      expect(req.request.params.get('status')).toBe('Paid');
      expect(req.request.params.get('month')).toBe('6');
      expect(req.request.params.get('year')).toBe('2026');
      req.flush(mockListResponse);
    });

    it('should include building_id param when building_id filter is provided', () => {
      const filters: PaymentFilters = { building_id: 5 };

      service.getPayments(filters).subscribe();

      const req = httpMock.expectOne((r) => r.url === apiUrl);
      expect(req.request.params.get('building_id')).toBe('5');
      req.flush(mockListResponse);
    });

    it('should include property_id param when property_id filter is provided', () => {
      const filters: PaymentFilters = { property_id: 7 };

      service.getPayments(filters).subscribe();

      const req = httpMock.expectOne((r) => r.url === apiUrl);
      expect(req.request.params.get('property_id')).toBe('7');
      req.flush(mockListResponse);
    });

    it('should include pagination params when page and page_size are provided', () => {
      const filters: PaymentFilters = { page: 2, page_size: 10 };

      service.getPayments(filters).subscribe();

      const req = httpMock.expectOne((r) => r.url === apiUrl);
      expect(req.request.params.get('page')).toBe('2');
      expect(req.request.params.get('page_size')).toBe('10');
      req.flush(mockListResponse);
    });

    it('should omit undefined, null, and empty-string filter values from params', () => {
      const filters: PaymentFilters = {
        status: undefined,
        month: undefined,
        building_id: 3,
      };

      service.getPayments(filters).subscribe();

      const req = httpMock.expectOne((r) => r.url === apiUrl);
      expect(req.request.params.has('status')).toBeFalse();
      expect(req.request.params.has('month')).toBeFalse();
      expect(req.request.params.get('building_id')).toBe('3');
      req.flush(mockListResponse);
    });

    it('should emit an error when the server responds with 500', () => {
      service.getPayments().subscribe({
        next: () => fail('expected an error, not payments'),
        error: (err) => expect(err.status).toBe(500),
      });

      const req = httpMock.expectOne(apiUrl);
      req.flush('Server error', { status: 500, statusText: 'Internal Server Error' });
    });
  });

  // ── getPayment ───────────────────────────────────────────────────────────────

  describe('getPayment', () => {
    it('should GET /payments/:id for a single payment', () => {
      service.getPayment(3).subscribe((res) => {
        expect(res).toEqual(mockPayment);
      });

      const req = httpMock.expectOne(`${apiUrl}/3`);
      expect(req.request.method).toBe('GET');
      req.flush(mockPayment);
    });

    it('should include all detail fields in the response', () => {
      service.getPayment(1).subscribe((res) => {
        expect(res.tenant_name).toBe('Rahim Uddin');
        expect(res.unit_number).toBe('3B');
        expect(res.building_name).toBe('Block A');
        expect(res.property_name).toBe('Sunrise Apartments');
      });

      const req = httpMock.expectOne(`${apiUrl}/1`);
      req.flush(mockPayment);
    });

    it('should emit a 404 error when the payment does not exist', () => {
      service.getPayment(999).subscribe({
        next: () => fail('expected an error, not a payment'),
        error: (err) => expect(err.status).toBe(404),
      });

      const req = httpMock.expectOne(`${apiUrl}/999`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });
  });

  // ── createPayment ────────────────────────────────────────────────────────────

  describe('createPayment', () => {
    it('should POST /payments with the request body', () => {
      const createdPayment: Payment = {
        id: 99,
        ...mockCreateRequest,
        amount_paid: 0,
        status: 'Due',
        created_at: '2026-06-01T00:00:00Z',
        updated_at: '2026-06-01T00:00:00Z',
      };

      service.createPayment(mockCreateRequest).subscribe((res) => {
        expect(res.id).toBe(99);
        expect(res.status).toBe('Due');
      });

      const req = httpMock.expectOne(apiUrl);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockCreateRequest);
      req.flush(createdPayment);
    });

    it('should emit a 400 error when the request payload is invalid', () => {
      const badRequest = {} as CreatePaymentRequest;

      service.createPayment(badRequest).subscribe({
        next: () => fail('expected an error, not a payment'),
        error: (err) => expect(err.status).toBe(400),
      });

      const req = httpMock.expectOne(apiUrl);
      req.flush('Validation failed', { status: 400, statusText: 'Bad Request' });
    });
  });

  // ── updatePayment ────────────────────────────────────────────────────────────

  describe('updatePayment', () => {
    it('should PUT /payments/:id with the update body', () => {
      const updatedPayment: Payment = {
        ...mockPayment,
        amount_paid: 15000,
        status: 'Paid',
        payment_method: 'Cash',
      };

      service.updatePayment(5, mockUpdateRequest).subscribe((res) => {
        expect(res.status).toBe('Paid');
        expect(res.amount_paid).toBe(15000);
      });

      const req = httpMock.expectOne(`${apiUrl}/5`);
      expect(req.request.method).toBe('PUT');
      expect(req.request.body).toEqual(mockUpdateRequest);
      req.flush(updatedPayment);
    });

    it('should support partial updates (only status)', () => {
      const partialUpdate: UpdatePaymentRequest = { status: 'Overdue' };

      service.updatePayment(5, partialUpdate).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/5`);
      expect(req.request.body).toEqual({ status: 'Overdue' });
      req.flush({ ...mockPayment, status: 'Overdue' });
    });

    it('should emit a 404 error when the payment does not exist', () => {
      service.updatePayment(999, mockUpdateRequest).subscribe({
        next: () => fail('expected an error'),
        error: (err) => expect(err.status).toBe(404),
      });

      const req = httpMock.expectOne(`${apiUrl}/999`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });
  });

  // ── bulkCreatePayments ───────────────────────────────────────────────────────

  describe('bulkCreatePayments', () => {
    it('should POST /payments/bulk with an array of requests', () => {
      const req1: CreatePaymentRequest = { ...mockCreateRequest, unit_id: 11 };
      const req2: CreatePaymentRequest = { ...mockCreateRequest, unit_id: 12 };

      service.bulkCreatePayments([req1, req2]).subscribe((res) => {
        expect(res).toBeTruthy();
      });

      const req = httpMock.expectOne(`${apiUrl}/bulk`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual([req1, req2]);
      req.flush({ created: 2 });
    });

    it('should handle bulk create with a single item array', () => {
      service.bulkCreatePayments([mockCreateRequest]).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/bulk`);
      expect(req.request.body.length).toBe(1);
      req.flush({ created: 1 });
    });
  });

  // ── getDashboardSummary ──────────────────────────────────────────────────────

  describe('getDashboardSummary', () => {
    it('should GET /dashboard/summary (not under /payments)', () => {
      service.getDashboardSummary().subscribe((res) => {
        expect(res).toEqual(mockSummary);
      });

      // Must NOT match the payments URL
      httpMock.expectNone(apiUrl);
      const req = httpMock.expectOne(`${baseUrl}/dashboard/summary`);
      expect(req.request.method).toBe('GET');
      req.flush(mockSummary);
    });

    it('should return all summary fields', () => {
      service.getDashboardSummary().subscribe((res) => {
        expect(res.total_due).toBe(100000);
        expect(res.collection_rate).toBe(85);
        expect(res.property_count).toBe(4);
        expect(res.building_count).toBe(8);
        expect(res.unit_count).toBe(40);
        expect(res.tenant_count).toBe(38);
      });

      const req = httpMock.expectOne(`${baseUrl}/dashboard/summary`);
      req.flush(mockSummary);
    });
  });

  // ── getBuildingReport ────────────────────────────────────────────────────────

  describe('getBuildingReport', () => {
    it('should GET /payments/building/:id/report with start_date and end_date params', () => {
      const buildingId = 2;
      const startDate = '2026-01-01';
      const endDate = '2026-03-31';

      service.getBuildingReport(buildingId, startDate, endDate).subscribe((res) => {
        expect(res).toBeTruthy();
      });

      const req = httpMock.expectOne(
        (r) => r.url === `${apiUrl}/building/${buildingId}/report`
      );
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('start_date')).toBe(startDate);
      expect(req.request.params.get('end_date')).toBe(endDate);
      req.flush({ report: [] });
    });

    it('should construct the correct URL for different building IDs', () => {
      service.getBuildingReport(7, '2026-01-01', '2026-12-31').subscribe();

      const req = httpMock.expectOne((r) => r.url.includes('/building/7/report'));
      expect(req.request.url).toBe(`${apiUrl}/building/7/report`);
      req.flush({});
    });
  });

  // ── getPropertyReport ────────────────────────────────────────────────────────

  describe('getPropertyReport', () => {
    it('should GET /payments/property/:id/report with start_date and end_date params', () => {
      const propertyId = 3;
      const startDate = '2026-01-01';
      const endDate = '2026-03-31';

      service.getPropertyReport(propertyId, startDate, endDate).subscribe((res) => {
        expect(res).toBeTruthy();
      });

      const req = httpMock.expectOne(
        (r) => r.url === `${apiUrl}/property/${propertyId}/report`
      );
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('start_date')).toBe(startDate);
      expect(req.request.params.get('end_date')).toBe(endDate);
      req.flush({ report: [] });
    });

    it('should construct the correct URL for different property IDs', () => {
      service.getPropertyReport(12, '2026-04-01', '2026-06-30').subscribe();

      const req = httpMock.expectOne((r) => r.url.includes('/property/12/report'));
      expect(req.request.url).toBe(`${apiUrl}/property/12/report`);
      req.flush({});
    });
  });

  // ── Error handling ───────────────────────────────────────────────────────────

  describe('Error handling', () => {
    it('getPayments should propagate a 500 server error', () => {
      service.getPayments().subscribe({
        next: () => fail('expected error'),
        error: (err) => {
          expect(err.status).toBe(500);
          expect(err.statusText).toBe('Internal Server Error');
        },
      });

      const req = httpMock.expectOne(apiUrl);
      req.flush('Server error', { status: 500, statusText: 'Internal Server Error' });
    });

    it('createPayment should propagate a 400 validation error', () => {
      service.createPayment(mockCreateRequest).subscribe({
        next: () => fail('expected error'),
        error: (err) => expect(err.status).toBe(400),
      });

      const req = httpMock.expectOne(apiUrl);
      req.flush('Bad request', { status: 400, statusText: 'Bad Request' });
    });

    it('updatePayment should propagate a 404 not-found error', () => {
      service.updatePayment(999, mockUpdateRequest).subscribe({
        next: () => fail('expected error'),
        error: (err) => expect(err.status).toBe(404),
      });

      const req = httpMock.expectOne(`${apiUrl}/999`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });

    it('getDashboardSummary should propagate a 403 forbidden error', () => {
      service.getDashboardSummary().subscribe({
        next: () => fail('expected error'),
        error: (err) => expect(err.status).toBe(403),
      });

      const req = httpMock.expectOne(`${baseUrl}/dashboard/summary`);
      req.flush('Forbidden', { status: 403, statusText: 'Forbidden' });
    });
  });

  // ── Concurrent requests ──────────────────────────────────────────────────────

  describe('Concurrent requests', () => {
    it('should handle simultaneous getPayments and getDashboardSummary calls', () => {
      let paymentsCompleted = false;
      let summaryCompleted = false;

      service.getPayments().subscribe(() => (paymentsCompleted = true));
      service.getDashboardSummary().subscribe(() => (summaryCompleted = true));

      const summaryReq = httpMock.expectOne(`${baseUrl}/dashboard/summary`);
      const paymentsReq = httpMock.expectOne(apiUrl);

      paymentsReq.flush(mockListResponse);
      summaryReq.flush(mockSummary);

      expect(paymentsCompleted).toBeTrue();
      expect(summaryCompleted).toBeTrue();
    });
  });
});
