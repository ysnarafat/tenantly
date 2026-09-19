import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { of, throwError } from 'rxjs';
import { PaymentUpdateDialog } from './payment-list';
import { PaymentService } from '../../../core/services/payment.service';
import {
  PaymentWithDetails,
  PaymentTransaction,
  PaymentTransactionAttachment,
} from '../../../core/models/payment.model';

describe('PaymentUpdateDialog', () => {
  let component: PaymentUpdateDialog;
  let fixture: ComponentFixture<PaymentUpdateDialog>;
  let paymentServiceSpy: jasmine.SpyObj<PaymentService>;
  let dialogRefSpy: jasmine.SpyObj<MatDialogRef<PaymentUpdateDialog>>;

  const mockPayment: PaymentWithDetails = {
    id: 5,
    unit_id: 10,
    tenant_id: 3,
    building_id: 2,
    property_id: 1,
    month: 6,
    year: 2026,
    amount_due: 12000,
    amount_paid: 10000,
    status: 'Partial',
    payment_method: 'Cash',
    payment_date: '2026-06-05',
    receipt_number: 'RCP-001',
    created_at: '2026-06-01T00:00:00Z',
    updated_at: '2026-06-05T00:00:00Z',
    property_name: 'Sunrise Apartments',
    building_name: 'Block A',
    building_code: 'BLK-A',
    unit_number: '3B',
    unit_type: 'Residential',
    tenant_name: 'Rahim Uddin',
  };

  function setup(data: PaymentWithDetails = mockPayment) {
    const paymentServiceSpyObj = jasmine.createSpyObj('PaymentService', [
      'getPaymentTransactions',
      'getPaymentTransactionAttachments',
      'uploadPaymentTransactionAttachment',
      'downloadPaymentTransactionAttachment',
      'deletePaymentTransactionAttachment',
    ]);
    paymentServiceSpyObj.getPaymentTransactions.and.returnValue(of([]));
    paymentServiceSpyObj.getPaymentTransactionAttachments.and.returnValue(of([]));

    const dialogRefSpyObj = jasmine.createSpyObj('MatDialogRef', ['close']);

    TestBed.configureTestingModule({
      imports: [PaymentUpdateDialog, NoopAnimationsModule],
      providers: [
        { provide: PaymentService, useValue: paymentServiceSpyObj },
        { provide: MatDialogRef, useValue: dialogRefSpyObj },
        { provide: MAT_DIALOG_DATA, useValue: data },
        { provide: MatSnackBar, useValue: jasmine.createSpyObj('MatSnackBar', ['open']) },
      ],
    }).compileComponents();

    paymentServiceSpy = paymentServiceSpyObj;
    dialogRefSpy = dialogRefSpyObj;
    fixture = TestBed.createComponent(PaymentUpdateDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  // ── remainingDue / amount_to_add defaults ───────────────────────────────────

  describe('remainingDue', () => {
    it('defaults amount_to_add to the remaining balance (amount_due - amount_paid)', () => {
      setup();
      expect(component.remainingDue).toBe(2000);
      expect(component.form.get('amount_to_add')?.value).toBe(2000);
    });

    it('never lets remainingDue go negative when already overpaid', () => {
      setup({ ...mockPayment, amount_due: 10000, amount_paid: 12000 });
      expect(component.remainingDue).toBe(0);
    });
  });

  // ── amount_to_add max validator ──────────────────────────────────────────────

  describe('amount_to_add validation', () => {
    it('is valid at exactly the remaining due (boundary)', () => {
      setup();
      component.form.get('amount_to_add')?.setValue(2000);
      expect(component.form.get('amount_to_add')?.hasError('max')).toBeFalse();
    });

    it('is invalid above the remaining due — mirrors the backend rejection', () => {
      setup();
      component.form.get('amount_to_add')?.setValue(2000.01);
      expect(component.form.get('amount_to_add')?.hasError('max')).toBeTrue();
      expect(component.form.invalid).toBeTrue();
    });

    it('is invalid when negative', () => {
      setup();
      component.form.get('amount_to_add')?.setValue(-1);
      expect(component.form.get('amount_to_add')?.hasError('min')).toBeTrue();
    });
  });

  // ── submit() ─────────────────────────────────────────────────────────────────

  describe('submit', () => {
    it('closes with a transaction result when a positive amount is entered', () => {
      setup();
      component.form.patchValue({
        amount_to_add: 2000,
        payment_method: 'bKash',
        payment_date: '2026-06-20',
        notes: 'final installment',
      });

      component.submit();

      expect(dialogRefSpy.close).toHaveBeenCalledWith(
        jasmine.objectContaining({
          kind: 'transaction',
          req: jasmine.objectContaining({
            amount: 2000,
            payment_method: 'bKash',
            payment_date: '2026-06-20',
            notes: 'final installment',
          }),
        })
      );
    });

    it('closes with a metadata result when the amount is zero but other fields changed', () => {
      setup();
      component.form.patchValue({ amount_to_add: 0, notes: 'corrected note' });

      component.submit();

      expect(dialogRefSpy.close).toHaveBeenCalledWith(
        jasmine.objectContaining({
          kind: 'metadata',
          req: jasmine.objectContaining({ notes: 'corrected note' }),
        })
      );
    });

    it('closes with undefined when nothing meaningful was entered', () => {
      setup({ ...mockPayment, amount_due: 10000, amount_paid: 10000, payment_method: '' });
      component.form.patchValue({
        amount_to_add: 0,
        payment_method: '',
        payment_date: '',
        notes: '',
      });

      component.submit();

      expect(dialogRefSpy.close).toHaveBeenCalledWith();
    });
  });

  // ── Transaction history / attachments ───────────────────────────────────────

  describe('transaction history', () => {
    const txns: PaymentTransaction[] = [
      {
        id: 1,
        payment_id: 5,
        amount: 10000,
        payment_method: 'Cash',
        payment_date: '2026-06-05',
        receipt_number: 'RCP-001',
        created_at: '2026-06-05T00:00:00Z',
      },
    ];

    it('loads transactions and their attachments on init', () => {
      const attachments: PaymentTransactionAttachment[] = [
        {
          id: 1,
          payment_transaction_id: 1,
          file_name: 'receipt.jpg',
          content_type: 'image/jpeg',
          file_size: 2048,
          created_at: '2026-06-05T00:00:00Z',
        },
      ];
      const paymentServiceSpyObj = jasmine.createSpyObj('PaymentService', [
        'getPaymentTransactions',
        'getPaymentTransactionAttachments',
      ]);
      paymentServiceSpyObj.getPaymentTransactions.and.returnValue(of(txns));
      paymentServiceSpyObj.getPaymentTransactionAttachments.and.returnValue(of(attachments));

      TestBed.configureTestingModule({
        imports: [PaymentUpdateDialog, NoopAnimationsModule],
        providers: [
          { provide: PaymentService, useValue: paymentServiceSpyObj },
          { provide: MatDialogRef, useValue: jasmine.createSpyObj('MatDialogRef', ['close']) },
          { provide: MAT_DIALOG_DATA, useValue: mockPayment },
          { provide: MatSnackBar, useValue: jasmine.createSpyObj('MatSnackBar', ['open']) },
        ],
      }).compileComponents();

      const localFixture = TestBed.createComponent(PaymentUpdateDialog);
      const localComponent = localFixture.componentInstance;
      localFixture.detectChanges();

      expect(localComponent.transactions()).toEqual(txns);
      expect(localComponent.attachmentCount(1)).toBe(1);
      expect(localComponent.attachmentsFor(1)).toEqual(attachments);
    });

    it('toggleAttachments expands and collapses a row', () => {
      setup();
      expect(component.expandedTxnId()).toBeNull();
      component.toggleAttachments(1);
      expect(component.expandedTxnId()).toBe(1);
      component.toggleAttachments(1);
      expect(component.expandedTxnId()).toBeNull();
    });
  });

  // ── formatFileSize ───────────────────────────────────────────────────────────

  describe('formatFileSize', () => {
    it('formats bytes, KB, and MB appropriately', () => {
      setup();
      expect(component.formatFileSize(500)).toBe('500 B');
      expect(component.formatFileSize(2048)).toBe('2.0 KB');
      expect(component.formatFileSize(3 * 1024 * 1024)).toBe('3.0 MB');
    });
  });

  // ── deleteAttachment ─────────────────────────────────────────────────────────

  describe('deleteAttachment', () => {
    it('removes the attachment from state on success', () => {
      setup();
      const attachment: PaymentTransactionAttachment = {
        id: 9,
        payment_transaction_id: 1,
        file_name: 'receipt.jpg',
        content_type: 'image/jpeg',
        file_size: 1024,
        created_at: '2026-06-05T00:00:00Z',
      };
      component.attachmentsByTxn.set({ 1: [attachment] });
      paymentServiceSpy.deletePaymentTransactionAttachment.and.returnValue(
        of({ message: 'deleted' })
      );

      component.deleteAttachment(1, attachment);

      expect(component.attachmentsFor(1)).toEqual([]);
    });

    it('leaves state unchanged when the delete request fails', () => {
      setup();
      const attachment: PaymentTransactionAttachment = {
        id: 9,
        payment_transaction_id: 1,
        file_name: 'receipt.jpg',
        content_type: 'image/jpeg',
        file_size: 1024,
        created_at: '2026-06-05T00:00:00Z',
      };
      component.attachmentsByTxn.set({ 1: [attachment] });
      paymentServiceSpy.deletePaymentTransactionAttachment.and.returnValue(
        throwError(() => ({ status: 500 }))
      );

      component.deleteAttachment(1, attachment);

      expect(component.attachmentsFor(1)).toEqual([attachment]);
    });
  });
});
