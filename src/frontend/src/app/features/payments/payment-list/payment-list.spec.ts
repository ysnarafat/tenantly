import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { of, throwError } from 'rxjs';
import { PageEvent } from '@angular/material/paginator';
import { PaymentList } from './payment-list';
import { PaymentService } from '../../../core/services/payment.service';
import {
  PaymentWithDetails,
  DashboardSummary,
  PaymentListResponse,
  PaymentStatus,
} from '../../../core/models/payment.model';

describe('PaymentList', () => {
  let component: PaymentList;
  let fixture: ComponentFixture<PaymentList>;
  let paymentServiceSpy: jasmine.SpyObj<PaymentService>;
  let dialogSpy: jasmine.SpyObj<MatDialog>;
  let snackBarSpy: jasmine.SpyObj<MatSnackBar>;

  // ── Shared fixtures ────────────────────────────────────────────────────────

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

  const mockListResponse: PaymentListResponse = {
    payments: [mockPayment],
    total: 1,
    page: 1,
    page_size: 20,
    total_pages: 1,
  };

  // ── TestBed setup ──────────────────────────────────────────────────────────

  beforeEach(async () => {
    paymentServiceSpy = jasmine.createSpyObj('PaymentService', [
      'getPayments',
      'getDashboardSummary',
      'createPayment',
      'updatePayment',
    ]);
    paymentServiceSpy.getPayments.and.returnValue(of(mockListResponse));
    paymentServiceSpy.getDashboardSummary.and.returnValue(of(mockSummary));

    dialogSpy = jasmine.createSpyObj('MatDialog', ['open']);
    dialogSpy.open.and.returnValue({ afterClosed: () => of(null) } as any);

    snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);

    await TestBed.configureTestingModule({
      imports: [PaymentList, NoopAnimationsModule, HttpClientTestingModule],
      providers: [
        { provide: PaymentService, useValue: paymentServiceSpy },
        { provide: MatDialog, useValue: dialogSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
      ],
    })
      .overrideProvider(MatDialog, { useValue: dialogSpy })
      .compileComponents();

    fixture = TestBed.createComponent(PaymentList);
    component = fixture.componentInstance;
    fixture.detectChanges(); // triggers ngOnInit
  });

  // ── Initialization ─────────────────────────────────────────────────────────

  describe('Initialization', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should call loadPayments and loadSummary on ngOnInit', () => {
      // Both spies were called during fixture.detectChanges() above
      expect(paymentServiceSpy.getPayments).toHaveBeenCalled();
      expect(paymentServiceSpy.getDashboardSummary).toHaveBeenCalled();
    });

    it('should expose the required displayedColumns', () => {
      expect(component.displayedColumns).toEqual([
        'tenant',
        'unit',
        'period',
        'amount_due',
        'amount_paid',
        'status',
        'actions',
      ]);
    });

    it('should initialise filterYear to the current year', () => {
      expect(component.filterYear).toBe(new Date().getFullYear().toString());
    });

    it('should initialise page signal to 1 and pageSize signal to 20', () => {
      expect(component.page()).toBe(1);
      expect(component.pageSize()).toBe(20);
    });
  });

  // ── loadPayments ───────────────────────────────────────────────────────────

  describe('loadPayments', () => {
    it('should populate the payments signal with data from the service', () => {
      expect(component.payments()).toEqual([mockPayment]);
    });

    it('should update the total signal from the response', () => {
      expect(component.total()).toBe(1);
    });

    it('should set loading to false after a successful fetch', () => {
      expect(component.loading()).toBeFalse();
    });

    it('should set loading to false after a failed fetch', () => {
      paymentServiceSpy.getPayments.and.returnValue(throwError(() => ({ status: 500 })));

      component.loadPayments();

      expect(component.loading()).toBeFalse();
    });

    it('should default payments to an empty array when the response has no payments field', () => {
      const emptyResponse: PaymentListResponse = {
        payments: [],
        total: 0,
        page: 1,
        page_size: 20,
        total_pages: 0,
      };
      paymentServiceSpy.getPayments.and.returnValue(of(emptyResponse));

      component.loadPayments();

      expect(component.payments()).toEqual([]);
      expect(component.total()).toBe(0);
    });

    it('should pass current page and pageSize as filters', () => {
      component.page.set(3);
      component.pageSize.set(10);

      component.loadPayments();

      const callArgs = paymentServiceSpy.getPayments.calls.mostRecent().args[0]!;
      expect(callArgs['page']).toBe(3);
      expect(callArgs['page_size']).toBe(10);
    });

    it('should include status filter when filterStatus is set', () => {
      component.filterStatus = 'Overdue';

      component.loadPayments();

      const callArgs = paymentServiceSpy.getPayments.calls.mostRecent().args[0]!;
      expect(callArgs['status']).toBe('Overdue');
    });

    it('should include month filter as a number when filterMonth is set', () => {
      component.filterMonth = '4';

      component.loadPayments();

      const callArgs = paymentServiceSpy.getPayments.calls.mostRecent().args[0]!;
      expect(callArgs['month']).toBe(4);
    });

    it('should include year filter as a number when filterYear is set', () => {
      component.filterYear = '2025';

      component.loadPayments();

      const callArgs = paymentServiceSpy.getPayments.calls.mostRecent().args[0]!;
      expect(callArgs['year']).toBe(2025);
    });
  });

  // ── loadSummary ────────────────────────────────────────────────────────────

  describe('loadSummary', () => {
    it('should populate the summary signal with data from the service', () => {
      expect(component.summary()).toEqual(mockSummary);
    });

    it('should not throw when getDashboardSummary errors (silent failure)', () => {
      paymentServiceSpy.getDashboardSummary.and.returnValue(
        throwError(() => new Error('network error'))
      );

      expect(() => component.loadSummary()).not.toThrow();
    });
  });

  // ── loading signal ─────────────────────────────────────────────────────────

  describe('loading signal', () => {
    it('should be false after initial load completes', () => {
      expect(component.loading()).toBeFalse();
    });

    it('should be false after an error response', () => {
      paymentServiceSpy.getPayments.and.returnValue(throwError(() => ({ status: 500 })));

      component.loadPayments();

      expect(component.loading()).toBeFalse();
    });
  });

  // ── applyFilters ───────────────────────────────────────────────────────────

  describe('applyFilters', () => {
    it('should reset page signal to 1 before reloading', () => {
      component.page.set(5);

      component.applyFilters();

      expect(component.page()).toBe(1);
    });

    it('should call loadPayments after resetting the page', () => {
      paymentServiceSpy.getPayments.calls.reset();

      component.applyFilters();

      expect(paymentServiceSpy.getPayments).toHaveBeenCalled();
    });
  });

  // ── clearFilters ───────────────────────────────────────────────────────────

  describe('clearFilters', () => {
    it('should clear filterStatus', () => {
      component.filterStatus = 'Overdue';

      component.clearFilters();

      expect(component.filterStatus).toBe('');
    });

    it('should clear filterMonth', () => {
      component.filterMonth = '6';

      component.clearFilters();

      expect(component.filterMonth).toBe('');
    });

    it('should reset filterYear to the current year', () => {
      component.filterYear = '2020';

      component.clearFilters();

      expect(component.filterYear).toBe(new Date().getFullYear().toString());
    });

    it('should reset page signal to 1', () => {
      component.page.set(4);

      component.clearFilters();

      expect(component.page()).toBe(1);
    });

    it('should call loadPayments after clearing filters', () => {
      paymentServiceSpy.getPayments.calls.reset();

      component.clearFilters();

      expect(paymentServiceSpy.getPayments).toHaveBeenCalled();
    });

    it('should reset all filters simultaneously', () => {
      component.filterStatus = 'Due';
      component.filterMonth = '3';
      component.filterYear = '2024';
      component.page.set(7);

      component.clearFilters();

      expect(component.filterStatus).toBe('');
      expect(component.filterMonth).toBe('');
      expect(component.filterYear).toBe(new Date().getFullYear().toString());
      expect(component.page()).toBe(1);
    });
  });

  // ── onPageChange ───────────────────────────────────────────────────────────

  describe('onPageChange', () => {
    it('should update the page signal (pageIndex + 1)', () => {
      const event: PageEvent = { pageIndex: 2, pageSize: 20, length: 100 };

      component.onPageChange(event);

      expect(component.page()).toBe(3);
    });

    it('should update the pageSize signal', () => {
      const event: PageEvent = { pageIndex: 0, pageSize: 50, length: 200 };

      component.onPageChange(event);

      expect(component.pageSize()).toBe(50);
    });

    it('should call loadPayments after updating pagination state', () => {
      paymentServiceSpy.getPayments.calls.reset();
      const event: PageEvent = { pageIndex: 1, pageSize: 20, length: 40 };

      component.onPageChange(event);

      expect(paymentServiceSpy.getPayments).toHaveBeenCalled();
    });

    it('should pass updated page and pageSize to the service', () => {
      const event: PageEvent = { pageIndex: 3, pageSize: 25, length: 200 };

      component.onPageChange(event);

      const callArgs = paymentServiceSpy.getPayments.calls.mostRecent().args[0]!;
      expect(callArgs['page']).toBe(4);
      expect(callArgs['page_size']).toBe(25);
    });
  });

  // ── getStatusClass ─────────────────────────────────────────────────────────

  describe('getStatusClass', () => {
    it('should return "status-paid" for Paid', () => {
      expect(component.getStatusClass('Paid')).toBe('status-paid');
    });

    it('should return "status-due" for Due', () => {
      expect(component.getStatusClass('Due')).toBe('status-due');
    });

    it('should return "status-partial" for Partial', () => {
      expect(component.getStatusClass('Partial')).toBe('status-partial');
    });

    it('should return "status-overdue" for Overdue', () => {
      expect(component.getStatusClass('Overdue')).toBe('status-overdue');
    });

    it('should return an empty string for an unknown status', () => {
      expect(component.getStatusClass('Unknown' as PaymentStatus)).toBe('');
    });
  });

  // ── monthName ──────────────────────────────────────────────────────────────

  describe('monthName', () => {
    it('should return "Jan" for month 1', () => {
      expect(component.monthName(1)).toBe('Jan');
    });

    it('should return "Jun" for month 6', () => {
      expect(component.monthName(6)).toBe('Jun');
    });

    it('should return "Dec" for month 12', () => {
      expect(component.monthName(12)).toBe('Dec');
    });

    it('should return "Apr" for month 4', () => {
      expect(component.monthName(4)).toBe('Apr');
    });

    it('should return "Sep" for month 9', () => {
      expect(component.monthName(9)).toBe('Sep');
    });

    it('should fall back to the number as string for an out-of-range month', () => {
      expect(component.monthName(13)).toBe('13');
    });
  });

  // ── collectionRateColor ────────────────────────────────────────────────────

  describe('collectionRateColor', () => {
    it('should return green (#4caf50) for a rate of 90', () => {
      expect(component.collectionRateColor(90)).toBe('#4caf50');
    });

    it('should return green (#4caf50) for a rate above 90', () => {
      expect(component.collectionRateColor(95)).toBe('#4caf50');
    });

    it('should return green (#4caf50) for a rate of 100', () => {
      expect(component.collectionRateColor(100)).toBe('#4caf50');
    });

    it('should return orange (#ff9800) for a rate of 70', () => {
      expect(component.collectionRateColor(70)).toBe('#ff9800');
    });

    it('should return orange (#ff9800) for a rate between 70 and 89', () => {
      expect(component.collectionRateColor(80)).toBe('#ff9800');
    });

    it('should return red (#f44336) for a rate below 70', () => {
      expect(component.collectionRateColor(69)).toBe('#f44336');
    });

    it('should return red (#f44336) for a rate of 0', () => {
      expect(component.collectionRateColor(0)).toBe('#f44336');
    });
  });

  // ── openCreateDialog ───────────────────────────────────────────────────────

  describe('openCreateDialog', () => {
    it('should open the MatDialog', () => {
      component.openCreateDialog();

      expect(dialogSpy.open).toHaveBeenCalled();
    });

    it('should call createPayment and reload when dialog returns a request', () => {
      const createReq = {
        unit_id: 11,
        tenant_id: 6,
        building_id: 2,
        property_id: 3,
        month: 6,
        year: 2026,
        amount_due: 12000,
      };
      dialogSpy.open.and.returnValue({ afterClosed: () => of(createReq) } as any);
      paymentServiceSpy.createPayment.and.returnValue(of({} as any));
      paymentServiceSpy.getPayments.calls.reset();
      paymentServiceSpy.getDashboardSummary.calls.reset();

      component.openCreateDialog();

      expect(paymentServiceSpy.createPayment).toHaveBeenCalledWith(createReq);
      expect(paymentServiceSpy.getPayments).toHaveBeenCalled();
      expect(paymentServiceSpy.getDashboardSummary).toHaveBeenCalled();
    });

    it('should not call createPayment when dialog is cancelled (returns null)', () => {
      dialogSpy.open.and.returnValue({ afterClosed: () => of(null) } as any);

      component.openCreateDialog();

      expect(paymentServiceSpy.createPayment).not.toHaveBeenCalled();
    });
  });

  // ── openUpdateDialog ───────────────────────────────────────────────────────

  describe('openUpdateDialog', () => {
    it('should open the MatDialog with the payment as dialog data', () => {
      component.openUpdateDialog(mockPayment);

      expect(dialogSpy.open).toHaveBeenCalledWith(
        jasmine.any(Function),
        jasmine.objectContaining({ data: mockPayment })
      );
    });

    it('should call updatePayment and reload when dialog returns an update request', () => {
      const updateReq = { amount_paid: 15000, status: 'Paid' as PaymentStatus };
      dialogSpy.open.and.returnValue({ afterClosed: () => of(updateReq) } as any);
      paymentServiceSpy.updatePayment.and.returnValue(of({} as any));
      paymentServiceSpy.getPayments.calls.reset();
      paymentServiceSpy.getDashboardSummary.calls.reset();

      component.openUpdateDialog(mockPayment);

      expect(paymentServiceSpy.updatePayment).toHaveBeenCalledWith(mockPayment.id, updateReq);
      expect(paymentServiceSpy.getPayments).toHaveBeenCalled();
      expect(paymentServiceSpy.getDashboardSummary).toHaveBeenCalled();
    });

    it('should not call updatePayment when dialog is cancelled (returns undefined)', () => {
      dialogSpy.open.and.returnValue({ afterClosed: () => of(undefined) } as any);

      component.openUpdateDialog(mockPayment);

      expect(paymentServiceSpy.updatePayment).not.toHaveBeenCalled();
    });
  });

  // ── totalPages computed ────────────────────────────────────────────────────

  describe('totalPages computed', () => {
    it('should compute totalPages as ceil(total / pageSize)', () => {
      component.total.set(45);
      component.pageSize.set(20);

      expect(component.totalPages()).toBe(3);
    });

    it('should return 1 when total is 0 to avoid an empty state', () => {
      component.total.set(0);
      component.pageSize.set(20);

      expect(component.totalPages()).toBe(1);
    });

    it('should return 1 when total exactly equals pageSize', () => {
      component.total.set(20);
      component.pageSize.set(20);

      expect(component.totalPages()).toBe(1);
    });
  });

  // ── Data shape ─────────────────────────────────────────────────────────────

  describe('Data shape', () => {
    it('should expose the statuses list with all four PaymentStatus values', () => {
      expect(component.statuses).toEqual(['Paid', 'Due', 'Partial', 'Overdue']);
    });

    it('should expose a months list with 12 entries', () => {
      expect(component.months.length).toBe(12);
    });

    it('should expose a years list starting from the current year', () => {
      expect(component.years[0]).toBe(new Date().getFullYear());
    });
  });
});
