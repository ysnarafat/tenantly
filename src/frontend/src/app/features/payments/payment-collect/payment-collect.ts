import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatInputModule } from '@angular/material/input';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatChipsModule } from '@angular/material/chips';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { forkJoin, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { PaymentService } from '../../../core/services/payment.service';
import {
  PaymentWithDetails,
  UpdatePaymentRequest,
  GenerateMonthlyPaymentsRequest,
  GenerateMonthlyPaymentsResult,
} from '../../../core/models/payment.model';
import { LoadingSpinner } from '../../../shared/components/loading-spinner/loading-spinner';
import { GeneratePaymentsDialog } from '../payment-list/payment-list';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

export interface CollectRow {
  payment: PaymentWithDetails;
  amountPaid: number;
  paymentMethod: string;
  paymentDate: string;
  skip: boolean;
  saving: boolean;
  saved: boolean;
  error: string | null;
}

type SaveOutcome = { row: CollectRow; ok: true } | { row: CollectRow; ok: false; message: string };

export const PAYMENT_METHODS = ['Cash', 'Bank Transfer', 'bKash', 'Nagad', 'Cheque'];

// Suffixes for `PAYMENT_COLLECT.MONTHS.<KEY>` translation keys, indexed by month - 1.
export const MONTH_KEYS = [
  'JAN',
  'FEB',
  'MAR',
  'APR',
  'MAY',
  'JUN',
  'JUL',
  'AUG',
  'SEP',
  'OCT',
  'NOV',
  'DEC',
];

export function todayIso(): string {
  return new Date().toISOString().slice(0, 10);
}

@Component({
  selector: 'app-payment-collect',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatSelectModule,
    MatInputModule,
    MatCheckboxModule,
    MatTooltipModule,
    MatChipsModule,
    TranslateModule,
    LoadingSpinner,
  ],
  templateUrl: './payment-collect.html',
  styleUrl: './payment-collect.scss',
})
export class PaymentCollect implements OnInit {
  private paymentService = inject(PaymentService);
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);

  paymentMethods = PAYMENT_METHODS;
  monthKeys = MONTH_KEYS;
  months = Array.from({ length: 12 }, (_, i) => i + 1);
  years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() + 1 - i);

  month = signal(new Date().getMonth() + 1);
  year = signal(new Date().getFullYear());
  loading = signal(false);
  saving = signal(false);
  loaded = signal(false);
  rows = signal<CollectRow[]>([]);

  pendingCount = computed(() => this.rows().filter((r) => !r.skip && !r.saved).length);
  savedCount = computed(() => this.rows().filter((r) => r.saved).length);
  failedCount = computed(() => this.rows().filter((r) => !!r.error).length);
  totalOutstanding = computed(() =>
    this.rows().reduce((sum, r) => sum + (r.payment.amount_due - r.payment.amount_paid), 0)
  );
  isEmpty = computed(() => this.loaded() && !this.loading() && this.rows().length === 0);
  canSaveAll = computed(
    () => !this.saving() && this.rows().some((r) => !r.skip && !r.saved && r.amountPaid > 0)
  );

  ngOnInit(): void {
    this.load();
  }

  onPeriodChange(): void {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.paymentService.getAllDuePayments(this.month(), this.year()).subscribe({
      next: (payments) => {
        this.rows.set(payments.map((payment) => this.toRow(payment)));
        this.loading.set(false);
        this.loaded.set(true);
      },
      error: (err) => {
        console.error('Error loading due payments:', safeErrorMessage(err));
        notifyError(this.snackBar, 'Failed to load outstanding payments');
        this.loading.set(false);
        this.loaded.set(true);
      },
    });
  }

  private toRow(payment: PaymentWithDetails): CollectRow {
    return {
      payment,
      amountPaid: Math.max(payment.amount_due - payment.amount_paid, 0),
      paymentMethod: '',
      paymentDate: todayIso(),
      skip: false,
      saving: false,
      saved: false,
      error: null,
    };
  }

  toggleSkip(row: CollectRow): void {
    row.skip = !row.skip;
    this.touch();
  }

  // Called from the template on every input change — signals need a new array
  // reference to notify computed()s, since we mutate the row objects in place.
  touch(): void {
    this.rows.update((rows) => [...rows]);
  }

  saveAll(): void {
    const pending = this.rows().filter((r) => !r.skip && !r.saved && r.amountPaid > 0);
    if (pending.length === 0) return;

    pending.forEach((r) => {
      r.saving = true;
      r.error = null;
    });
    this.touch();
    this.saving.set(true);

    const requests = pending.map((row) => {
      const req: UpdatePaymentRequest = {
        amount_paid: row.amountPaid,
        payment_method: row.paymentMethod || undefined,
        payment_date: row.paymentDate || undefined,
      };
      return this.paymentService.updatePayment(row.payment.id, req).pipe(
        map((): SaveOutcome => ({ row, ok: true })),
        catchError((err) =>
          of<SaveOutcome>({
            row,
            ok: false,
            message: err?.error?.error || 'Failed to save',
          })
        )
      );
    });

    forkJoin(requests).subscribe((outcomes) => {
      let succeeded = 0;
      let failed = 0;
      outcomes.forEach((outcome) => {
        outcome.row.saving = false;
        if (outcome.ok) {
          outcome.row.saved = true;
          outcome.row.error = null;
          succeeded++;
        } else {
          outcome.row.error = outcome.message;
          failed++;
        }
      });
      this.touch();
      this.saving.set(false);

      const message =
        failed > 0 ? `${succeeded} recorded, ${failed} failed` : `${succeeded} recorded`;
      if (failed > 0) {
        notifyError(this.snackBar, message);
      } else {
        notifySuccess(this.snackBar, message);
      }
    });
  }

  openGenerateDialog(): void {
    const ref = this.dialog.open(GeneratePaymentsDialog, { width: '420px', maxWidth: '95vw' });
    ref.afterClosed().subscribe((req: GenerateMonthlyPaymentsRequest | undefined) => {
      if (!req) return;

      this.loading.set(true);
      this.paymentService.generateMonthlyPayments(req).subscribe({
        next: (result: GenerateMonthlyPaymentsResult) => {
          notifySuccess(this.snackBar, `Generated ${result.generated} due payment(s)`);
          this.month.set(req.month);
          this.year.set(req.year);
          this.load();
        },
        error: (err) => {
          console.error('Error generating monthly payments:', safeErrorMessage(err));
          notifyError(this.snackBar, 'Failed to generate monthly payments');
          this.loading.set(false);
        },
      });
    });
  }
}
