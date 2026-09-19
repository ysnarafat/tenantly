import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialog, MatDialogRef } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { of, throwError } from 'rxjs';
import { PaymentCollect, todayIso } from './payment-collect';
import { PaymentService } from '../../../core/services/payment.service';
import {
  PaymentWithDetails,
  GenerateMonthlyPaymentsRequest,
} from '../../../core/models/payment.model';

function dialogRefStub(result: unknown): MatDialogRef<unknown> {
  return { afterClosed: () => of(result) } as unknown as MatDialogRef<unknown>;
}

function makePayment(overrides: Partial<PaymentWithDetails> = {}): PaymentWithDetails {
  return {
    id: 1,
    unit_id: 10,
    tenant_id: 5,
    building_id: 2,
    property_id: 3,
    month: 6,
    year: 2026,
    amount_due: 10000,
    amount_paid: 0,
    status: 'Due',
    created_at: '2026-06-01T00:00:00Z',
    updated_at: '2026-06-01T00:00:00Z',
    property_name: 'Sunrise Apartments',
    building_name: 'Block A',
    building_code: 'BLK-A',
    unit_number: '3B',
    unit_type: 'Residential',
    tenant_name: 'Rahim Uddin',
    ...overrides,
  };
}

describe('PaymentCollect', () => {
  let component: PaymentCollect;
  let fixture: ComponentFixture<PaymentCollect>;
  let paymentService: jasmine.SpyObj<PaymentService>;
  let dialog: jasmine.SpyObj<MatDialog>;
  let snackBar: jasmine.SpyObj<MatSnackBar>;

  beforeEach(async () => {
    const paymentServiceSpy = jasmine.createSpyObj('PaymentService', [
      'getAllDuePayments',
      'recordPaymentTransaction',
      'generateMonthlyPayments',
    ]);
    const dialogSpy = jasmine.createSpyObj('MatDialog', ['open']);
    const snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);

    await TestBed.configureTestingModule({
      imports: [PaymentCollect, TranslateModule.forRoot()],
      providers: [
        { provide: PaymentService, useValue: paymentServiceSpy },
        { provide: MatDialog, useValue: dialogSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
      ],
    }).compileComponents();

    paymentService = TestBed.inject(PaymentService) as jasmine.SpyObj<PaymentService>;
    dialog = TestBed.inject(MatDialog) as jasmine.SpyObj<MatDialog>;
    snackBar = TestBed.inject(MatSnackBar) as jasmine.SpyObj<MatSnackBar>;

    paymentService.getAllDuePayments.and.returnValue(of([]));

    fixture = TestBed.createComponent(PaymentCollect);
    component = fixture.componentInstance;
  });

  describe('Loading', () => {
    it('fetches due payments for the current month/year on init', () => {
      const now = new Date();
      fixture.detectChanges();

      expect(paymentService.getAllDuePayments).toHaveBeenCalledWith(
        now.getMonth() + 1,
        now.getFullYear()
      );
    });

    it('maps each payment to a row defaulted to the remaining owed amount, today, and no method', () => {
      const payment = makePayment({ amount_due: 10000, amount_paid: 4000 });
      paymentService.getAllDuePayments.and.returnValue(of([payment]));

      fixture.detectChanges();

      const [row] = component.rows();
      expect(row.amountPaid).toBe(6000);
      expect(row.paymentMethod).toBe('');
      expect(row.paymentDate).toBe(todayIso());
      expect(row.skip).toBeFalse();
      expect(row.saved).toBeFalse();
    });

    it('never lets the remaining-owed default go negative', () => {
      // Defensive: a payment somehow already overpaid shouldn't produce a
      // negative amountPaid default, which would fail the >0 save filter oddly.
      const payment = makePayment({ amount_due: 10000, amount_paid: 12000 });
      paymentService.getAllDuePayments.and.returnValue(of([payment]));

      fixture.detectChanges();

      expect(component.rows()[0].amountPaid).toBe(0);
    });

    it('is empty once loaded with zero outstanding payments', () => {
      paymentService.getAllDuePayments.and.returnValue(of([]));

      fixture.detectChanges();

      expect(component.isEmpty()).toBeTrue();
      expect(component.loading()).toBeFalse();
    });

    it('is not empty while still loading, even with no rows yet', () => {
      expect(component.isEmpty()).toBeFalse();
    });

    it('shows an error toast and still resolves loading/loaded on failure', () => {
      paymentService.getAllDuePayments.and.returnValue(throwError(() => ({ status: 500 })));

      fixture.detectChanges();

      expect(snackBar.open).toHaveBeenCalled();
      expect(component.loading()).toBeFalse();
      expect(component.loaded()).toBeTrue();
      expect(component.rows()).toEqual([]);
    });

    it('reloads with the newly selected month/year when the period changes', () => {
      fixture.detectChanges();
      paymentService.getAllDuePayments.calls.reset();

      component.month.set(3);
      component.year.set(2027);
      component.onPeriodChange();

      expect(paymentService.getAllDuePayments).toHaveBeenCalledWith(3, 2027);
    });
  });

  describe('Skip', () => {
    it('excludes a skipped row from the pending count', () => {
      paymentService.getAllDuePayments.and.returnValue(of([makePayment()]));
      fixture.detectChanges();

      expect(component.pendingCount()).toBe(1);
      component.toggleSkip(component.rows()[0]);

      expect(component.rows()[0].skip).toBeTrue();
      expect(component.pendingCount()).toBe(0);
    });

    it('toggles back off on a second call', () => {
      paymentService.getAllDuePayments.and.returnValue(of([makePayment()]));
      fixture.detectChanges();

      const row = component.rows()[0];
      component.toggleSkip(row);
      component.toggleSkip(row);

      expect(row.skip).toBeFalse();
    });
  });

  describe('canSaveAll', () => {
    it('is false with no rows', () => {
      fixture.detectChanges();
      expect(component.canSaveAll()).toBeFalse();
    });

    it('is true when at least one non-skipped row has a positive amount', () => {
      paymentService.getAllDuePayments.and.returnValue(of([makePayment()]));
      fixture.detectChanges();

      expect(component.canSaveAll()).toBeTrue();
    });

    it('is false when every row is skipped', () => {
      paymentService.getAllDuePayments.and.returnValue(of([makePayment()]));
      fixture.detectChanges();
      component.toggleSkip(component.rows()[0]);

      expect(component.canSaveAll()).toBeFalse();
    });

    it('is false when the only row has a zero amount', () => {
      paymentService.getAllDuePayments.and.returnValue(
        of([makePayment({ amount_due: 5000, amount_paid: 5000 })])
      );
      fixture.detectChanges();

      expect(component.canSaveAll()).toBeFalse();
    });
  });

  describe('saveAll', () => {
    it('sends amount/payment_method/payment_date for every pending row and marks them saved', () => {
      const payment = makePayment({ id: 1 });
      paymentService.getAllDuePayments.and.returnValue(of([payment]));
      paymentService.recordPaymentTransaction.and.returnValue(of({ ...payment, status: 'Paid' }));
      fixture.detectChanges();

      const row = component.rows()[0];
      row.amountPaid = 10000;
      row.paymentMethod = 'bKash';
      row.paymentDate = '2026-06-15';

      component.saveAll();

      expect(paymentService.recordPaymentTransaction).toHaveBeenCalledWith(1, {
        amount: 10000,
        payment_method: 'bKash',
        payment_date: '2026-06-15',
      });
      expect(component.rows()[0].saved).toBeTrue();
      expect(component.rows()[0].saving).toBeFalse();
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('does not call recordPaymentTransaction for skipped, already-saved, or zero-amount rows', () => {
      const skipped = makePayment({ id: 1 });
      const zero = makePayment({ id: 2, amount_due: 5000, amount_paid: 5000 });
      const normal = makePayment({ id: 3 });
      paymentService.getAllDuePayments.and.returnValue(of([skipped, zero, normal]));
      paymentService.recordPaymentTransaction.and.returnValue(of(normal));
      fixture.detectChanges();

      component.toggleSkip(component.rows()[0]);
      component.saveAll();

      expect(paymentService.recordPaymentTransaction).toHaveBeenCalledTimes(1);
      expect(paymentService.recordPaymentTransaction).toHaveBeenCalledWith(3, jasmine.any(Object));
    });

    it('keeps failed rows unsaved with an error message while succeeded rows are marked saved', () => {
      const ok = makePayment({ id: 1 });
      const bad = makePayment({ id: 2 });
      paymentService.getAllDuePayments.and.returnValue(of([ok, bad]));
      paymentService.recordPaymentTransaction.and.callFake((id: number) =>
        id === 1
          ? of({ ...ok, status: 'Paid' as const })
          : throwError(() => ({ error: { error: 'unit locked' } }))
      );
      fixture.detectChanges();

      component.saveAll();

      const [rowOk, rowBad] = component.rows();
      expect(rowOk.saved).toBeTrue();
      expect(rowOk.error).toBeNull();
      expect(rowBad.saved).toBeFalse();
      expect(rowBad.error).toBe('unit locked');
      expect(rowBad.saving).toBeFalse();
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('falls back to a generic message when the server error has no detail', () => {
      paymentService.getAllDuePayments.and.returnValue(of([makePayment({ id: 1 })]));
      paymentService.recordPaymentTransaction.and.returnValue(throwError(() => ({})));
      fixture.detectChanges();

      component.saveAll();

      expect(component.rows()[0].error).toBe('Failed to save');
    });

    it('is a no-op when there is nothing pending to save', () => {
      fixture.detectChanges();
      component.saveAll();

      expect(paymentService.recordPaymentTransaction).not.toHaveBeenCalled();
    });

    it('re-saving after a partial failure only retries the still-failed rows', () => {
      const ok = makePayment({ id: 1 });
      const bad = makePayment({ id: 2 });
      paymentService.getAllDuePayments.and.returnValue(of([ok, bad]));
      paymentService.recordPaymentTransaction.and.callFake((id: number) =>
        id === 1 ? of({ ...ok, status: 'Paid' as const }) : throwError(() => ({}))
      );
      fixture.detectChanges();
      component.saveAll();
      paymentService.recordPaymentTransaction.calls.reset();

      paymentService.recordPaymentTransaction.and.returnValue(
        of({ ...bad, status: 'Paid' as const })
      );
      component.saveAll();

      expect(paymentService.recordPaymentTransaction).toHaveBeenCalledTimes(1);
      expect(paymentService.recordPaymentTransaction).toHaveBeenCalledWith(2, jasmine.any(Object));
      expect(component.rows()[1].saved).toBeTrue();
    });
  });

  describe('openGenerateDialog', () => {
    it('generates payments for the chosen period and reloads on confirmation', () => {
      fixture.detectChanges();
      const req: GenerateMonthlyPaymentsRequest = { month: 7, year: 2026, due_day_of_month: 7 };
      dialog.open.and.returnValue(dialogRefStub(req));
      paymentService.generateMonthlyPayments.and.returnValue(
        of({ generated: 12, skipped: 0, failed: 0 })
      );
      paymentService.getAllDuePayments.calls.reset();

      component.openGenerateDialog();

      expect(paymentService.generateMonthlyPayments).toHaveBeenCalledWith(req);
      expect(component.month()).toBe(7);
      expect(component.year()).toBe(2026);
      expect(paymentService.getAllDuePayments).toHaveBeenCalledWith(7, 2026);
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('does nothing when the dialog is dismissed without a result', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(undefined));

      component.openGenerateDialog();

      expect(paymentService.generateMonthlyPayments).not.toHaveBeenCalled();
    });

    it('shows an error toast and stops loading when generation fails', () => {
      fixture.detectChanges();
      const req: GenerateMonthlyPaymentsRequest = { month: 7, year: 2026, due_day_of_month: 7 };
      dialog.open.and.returnValue(dialogRefStub(req));
      paymentService.generateMonthlyPayments.and.returnValue(throwError(() => ({})));

      component.openGenerateDialog();

      expect(component.loading()).toBeFalse();
      expect(snackBar.open).toHaveBeenCalled();
    });
  });
});
