import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import {
  FormsModule,
  ReactiveFormsModule,
  FormBuilder,
  FormGroup,
  Validators,
  FormControl,
} from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import {
  MatDialog,
  MatDialogModule,
  MatDialogRef,
  MAT_DIALOG_DATA,
} from '@angular/material/dialog';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatTabsModule } from '@angular/material/tabs';
import { PaymentService, PaymentFilters } from '../../../core/services/payment.service';
import { LeaseService } from '../../../core/services/lease.service';
import { LeaseWithDetails } from '../../../core/models/lease.model';
import {
  PaymentWithDetails,
  UpdatePaymentRequest,
  CreatePaymentTransactionRequest,
  PaymentTransaction,
  PaymentTransactionAttachment,
  MAX_ATTACHMENT_FILE_SIZE,
  ALLOWED_ATTACHMENT_CONTENT_TYPES,
  PaymentStatus,
  DashboardSummary,
  CreatePaymentRequest,
  LeaseSearchResult,
  GenerateMonthlyPaymentsRequest,
  GenerateMonthlyPaymentsResult,
} from '../../../core/models/payment.model';
import { debounceTime, switchMap } from 'rxjs/operators';
import { of } from 'rxjs';
import { DataTable } from '../../../shared/components/data-table/data-table';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

interface BuildingNode {
  building_id: number;
  building_name: string;
  building_code: string;
  payments: PaymentWithDetails[];
  totalDue: number;
  totalPaid: number;
  collectionRate: number;
}

interface PropertyNode {
  property_id: number;
  property_name: string;
  buildings: BuildingNode[];
  totalDue: number;
  totalPaid: number;
  paidCount: number;
  dueCount: number;
  overdueCount: number;
  partialCount: number;
  collectionRate: number;
}

@Component({
  selector: 'app-payment-list',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    FormsModule,
    ReactiveFormsModule,
    TranslateModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatTableModule,
    MatPaginatorModule,
    MatSelectModule,
    MatFormFieldModule,
    MatInputModule,
    MatDialogModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatProgressBarModule,
    MatSnackBarModule,
    MatDividerModule,
    MatTooltipModule,
    MatTabsModule,
    DataTable,
  ],
  templateUrl: './payment-list.html',
  styleUrls: ['./payment-list.scss'],
})
export class PaymentList implements OnInit {
  private paymentService = inject(PaymentService);
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);

  payments = signal<PaymentWithDetails[]>([]);
  summary = signal<DashboardSummary | null>(null);
  loading = signal(false);
  total = signal(0);
  page = signal(1);
  pageSize = signal(20);
  activeTab = signal(0);

  filterStatus = '';
  filterMonth = '';
  filterYear = new Date().getFullYear();

  get hasActiveFilters(): boolean {
    return !!(
      this.filterStatus ||
      this.filterMonth ||
      this.filterYear !== new Date().getFullYear()
    );
  }

  treePayments = signal<PaymentWithDetails[]>([]);
  treeLoading = signal(false);
  treeFilterMonth: number | '' = '';
  treeFilterYear = new Date().getFullYear();

  treeData = computed<PropertyNode[]>(() => {
    const propMap = new Map<number, PropertyNode>();
    for (const p of this.treePayments()) {
      if (!propMap.has(p.property_id)) {
        propMap.set(p.property_id, {
          property_id: p.property_id,
          property_name: p.property_name,
          buildings: [],
          totalDue: 0,
          totalPaid: 0,
          paidCount: 0,
          dueCount: 0,
          overdueCount: 0,
          partialCount: 0,
          collectionRate: 0,
        });
      }
      const prop = propMap.get(p.property_id)!;
      let bldg = prop.buildings.find((b) => b.building_id === p.building_id);
      if (!bldg) {
        bldg = {
          building_id: p.building_id,
          building_name: p.building_name,
          building_code: p.building_code,
          payments: [],
          totalDue: 0,
          totalPaid: 0,
          collectionRate: 0,
        };
        prop.buildings.push(bldg);
      }
      bldg.payments.push(p);
      bldg.totalDue += p.amount_due;
      bldg.totalPaid += p.amount_paid;
      prop.totalDue += p.amount_due;
      prop.totalPaid += p.amount_paid;
      if (p.status === 'Paid') prop.paidCount++;
      else if (p.status === 'Overdue') prop.overdueCount++;
      else if (p.status === 'Partial') prop.partialCount++;
      else prop.dueCount++;
    }
    return Array.from(propMap.values()).map((prop) => ({
      ...prop,
      collectionRate: prop.totalDue > 0 ? (prop.totalPaid / prop.totalDue) * 100 : 0,
      buildings: prop.buildings.map((b) => ({
        ...b,
        collectionRate: b.totalDue > 0 ? (b.totalPaid / b.totalDue) * 100 : 0,
      })),
    }));
  });

  displayedColumns = ['tenant', 'unit', 'period', 'amount_due', 'amount_paid', 'status', 'actions'];

  months = [
    { value: 1, label: 'January' },
    { value: 2, label: 'February' },
    { value: 3, label: 'March' },
    { value: 4, label: 'April' },
    { value: 5, label: 'May' },
    { value: 6, label: 'June' },
    { value: 7, label: 'July' },
    { value: 8, label: 'August' },
    { value: 9, label: 'September' },
    { value: 10, label: 'October' },
    { value: 11, label: 'November' },
    { value: 12, label: 'December' },
  ];

  years = Array.from({ length: 10 }, (_, i) => new Date().getFullYear() - i);
  statuses: PaymentStatus[] = ['Paid', 'Due', 'Partial', 'Overdue'];

  totalPages = computed(() => Math.ceil(this.total() / this.pageSize()) || 1);

  ngOnInit(): void {
    this.loadSummary();
    this.loadPayments();
  }

  loadSummary(): void {
    this.paymentService.getDashboardSummary().subscribe({
      next: (s) => this.summary.set(s),
      error: () => {},
    });
  }

  loadPayments(): void {
    this.loading.set(true);
    const filters: PaymentFilters = {
      page: this.page(),
      page_size: this.pageSize(),
    };
    if (this.filterStatus) filters['status'] = this.filterStatus;
    if (this.filterMonth) filters['month'] = +this.filterMonth;
    if (this.filterYear) filters['year'] = +this.filterYear;

    this.paymentService.getPayments(filters).subscribe({
      next: (res) => {
        this.payments.set(res.payments ?? []);
        this.total.set(res.total ?? 0);
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);
        notifyError(this.snackBar, 'Failed to load payments');
        console.error(safeErrorMessage(err));
      },
    });
  }

  loadTreePayments(): void {
    this.treeLoading.set(true);
    const filters: PaymentFilters = { page: 1, page_size: 500 };
    if (this.treeFilterMonth) filters['month'] = +this.treeFilterMonth;
    if (this.treeFilterYear) filters['year'] = +this.treeFilterYear;
    this.paymentService.getPayments(filters).subscribe({
      next: (res) => {
        this.treePayments.set(res.payments ?? []);
        this.treeLoading.set(false);
      },
      error: () => {
        this.treeLoading.set(false);
        notifyError(this.snackBar, 'Failed to load tree view');
      },
    });
  }

  onTabChange(index: number): void {
    this.activeTab.set(index);
    if (index === 1 && this.treePayments().length === 0) this.loadTreePayments();
  }

  onPageChange(event: PageEvent): void {
    this.page.set(event.pageIndex + 1);
    this.pageSize.set(event.pageSize);
    this.loadPayments();
  }

  applyFilters(): void {
    this.page.set(1);
    this.loadPayments();
  }

  clearFilters(): void {
    this.filterStatus = '';
    this.filterMonth = '';
    this.filterYear = new Date().getFullYear();
    this.page.set(1);
    this.loadPayments();
  }

  openGenerateDialog(): void {
    const ref = this.dialog.open(GeneratePaymentsDialog, { width: '420px', maxWidth: '95vw' });
    ref.afterClosed().subscribe((req: GenerateMonthlyPaymentsRequest | undefined) => {
      if (req) {
        this.loading.set(true);
        this.paymentService.generateMonthlyPayments(req).subscribe({
          next: (result: GenerateMonthlyPaymentsResult) => {
            this.loading.set(false);
            const msg = `Generated ${result.generated} · Skipped ${result.skipped} · Failed ${result.failed}`;
            if (result.failed > 0) {
              notifyError(this.snackBar, msg);
            } else {
              notifySuccess(this.snackBar, msg);
            }
            this.loadPayments();
            this.loadSummary();
            if (this.activeTab() === 1) this.loadTreePayments();
          },
          error: (err) => {
            this.loading.set(false);
            notifyError(this.snackBar, err?.error?.error ?? 'Failed to generate payments');
          },
        });
      }
    });
  }

  openCreateDialog(): void {
    const ref = this.dialog.open(PaymentCreateDialog, { width: '560px', maxWidth: '95vw' });
    ref.afterClosed().subscribe((req: CreatePaymentRequest | undefined) => {
      if (req) {
        this.paymentService.createPayment(req).subscribe({
          next: (payment) => {
            notifySuccess(this.snackBar, `Payment created — Receipt ${payment.receipt_number}`);
            this.loadPayments();
            this.loadSummary();
            if (this.activeTab() === 1) this.loadTreePayments();
          },
          error: (err) =>
            notifyError(this.snackBar, err?.error?.error ?? 'Failed to create payment'),
        });
      }
    });
  }

  openUpdateDialog(payment: PaymentWithDetails): void {
    const ref = this.dialog.open(PaymentUpdateDialog, {
      width: '480px',
      maxWidth: '95vw',
      data: payment,
    });
    ref.afterClosed().subscribe((result: PaymentUpdateResult | undefined) => {
      if (!result) return;
      // A positive amount records a new installment (accumulates on top of
      // whatever's already paid); otherwise it's a metadata-only correction
      // (payment method/date/notes) with no money involved.
      const obs =
        result.kind === 'transaction'
          ? this.paymentService.recordPaymentTransaction(payment.id, result.req)
          : this.paymentService.updatePayment(payment.id, result.req);
      obs.subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Payment updated');
          this.loadPayments();
          this.loadSummary();
          if (this.activeTab() === 1) this.loadTreePayments();
        },
        error: (err) => notifyError(this.snackBar, err?.error?.error ?? 'Failed to update payment'),
      });
    });
  }

  downloadReceipt(payment: PaymentWithDetails): void {
    this.paymentService.downloadReceipt(payment.id).subscribe({
      next: (blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `receipt-${payment.receipt_number || payment.id}.pdf`;
        a.click();
        URL.revokeObjectURL(url);
      },
      error: () => notifyError(this.snackBar, 'Failed to download receipt'),
    });
  }

  getStatusClass(status: PaymentStatus): string {
    const map: Record<PaymentStatus, string> = {
      Paid: 'status-paid',
      Due: 'status-due',
      Partial: 'status-partial',
      Overdue: 'status-overdue',
    };
    return map[status] ?? '';
  }

  monthName(month: number): string {
    return this.months.find((m) => m.value === month)?.label.slice(0, 3) ?? month.toString();
  }

  collectionRateColor(rate: number): string {
    if (rate >= 90) return '#4caf50';
    if (rate >= 70) return '#ff9800';
    return '#f44336';
  }
}

// ── Create Dialog ──────────────────────────────────────────────────────────────
export interface PaymentCreatePrefill {
  unit_id: number;
  unit_number: string;
  building_name: string;
  property_name: string;
  lease: LeaseSearchResult | null;
}

@Component({
  selector: 'app-payment-create-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatDialogModule,
    MatAutocompleteModule,
    MatProgressSpinnerModule,
    MatIconModule,
  ],
  template: `
    <h2 mat-dialog-title>New Payment</h2>
    @if (prefill) {
      <div class="unit-context-bar">
        <mat-icon>apartment</mat-icon>
        <span>{{ prefill.property_name }}</span>
        <span class="ctx-sep">›</span>
        <span>{{ prefill.building_name }}</span>
        <span class="ctx-sep">›</span>
        <strong>Unit {{ prefill.unit_number }}</strong>
      </div>
    }
    <mat-dialog-content>
      <form [formGroup]="form" class="dialog-form">
        <!-- Lease search -->
        <mat-form-field appearance="outline">
          <mat-label>Tenant / Lease</mat-label>
          <input
            matInput
            [formControl]="leaseSearch"
            [matAutocomplete]="leaseAuto"
            placeholder="Type tenant name or unit…"
          />
          <mat-autocomplete
            #leaseAuto="matAutocomplete"
            [displayWith]="leaseDisplay"
            (optionSelected)="onLeaseSelected($event.option.value)"
          >
            @for (l of filteredLeases(); track l.lease_id) {
              <mat-option [value]="l">
                <span class="lease-option-main">{{ l.tenant_name }}</span>
                <span class="lease-option-sub">
                  {{ l.unit_number }} · {{ l.building_name }} · ৳{{
                    l.monthly_rent | number: '1.0-0'
                  }}/mo
                </span>
              </mat-option>
            }
            @if (filteredLeases().length === 0 && leaseSearch.value) {
              <mat-option disabled>No active leases found</mat-option>
            }
          </mat-autocomplete>
          @if (form.get('tenant_id')?.hasError('required') && form.get('tenant_id')?.touched) {
            <mat-error>Please select a lease</mat-error>
          }
        </mat-form-field>

        <!-- Selected lease summary -->
        @if (selectedLease()) {
          <div class="lease-summary">
            <span>{{ selectedLease()!.property_name }}</span>
            <span class="sep">›</span>
            <span>{{ selectedLease()!.building_name }}</span>
            <span class="sep">›</span>
            <span>Unit {{ selectedLease()!.unit_number }}</span>
            <span class="sep">›</span>
            <span class="rent">৳{{ selectedLease()!.monthly_rent | number: '1.0-0' }}/mo</span>
          </div>
          @if (selectedLease()!.outstanding_balance > 0) {
            <div class="balance-hint balance-hint--due">
              Carrying forward outstanding balance of ৳{{
                selectedLease()!.outstanding_balance | number: '1.0-0'
              }}
              from prior periods
            </div>
          } @else {
            <div class="balance-hint balance-hint--clear">New month — full rent due</div>
          }
        }

        <!-- Search loading indicator -->
        @if (searching()) {
          <div class="search-loading">
            <mat-spinner diameter="16"></mat-spinner>
            <span>Searching leases...</span>
          </div>
        }

        <div class="row-2">
          <mat-form-field appearance="outline">
            <mat-label>Month</mat-label>
            <mat-select formControlName="month">
              @for (m of months; track m.value) {
                <mat-option [value]="m.value">{{ m.label }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>Year</mat-label>
            <mat-select formControlName="year">
              @for (y of years; track y) {
                <mat-option [value]="y">{{ y }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
        </div>

        <mat-form-field appearance="outline">
          <mat-label>Amount Due (BDT)</mat-label>
          <input matInput type="number" formControlName="amount_due" step="0.01" min="0.01" />
          <mat-error>Amount must be greater than 0</mat-error>
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Due Date</mat-label>
          <input matInput type="date" formControlName="due_date" />
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Payment Method</mat-label>
          <mat-select formControlName="payment_method">
            <mat-option value="">— None —</mat-option>
            <mat-option value="Cash">Cash</mat-option>
            <mat-option value="Bank Transfer">Bank Transfer</mat-option>
            <mat-option value="bKash">bKash</mat-option>
            <mat-option value="Nagad">Nagad</mat-option>
            <mat-option value="Cheque">Cheque</mat-option>
          </mat-select>
        </mat-form-field>

        <div class="row-2">
          <mat-form-field appearance="outline">
            <mat-label>Amount Paid (BDT)</mat-label>
            <input matInput type="number" formControlName="amount_paid" step="0.01" min="0" />
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>Payment Date</mat-label>
            <input matInput type="date" formControlName="payment_date" />
          </mat-form-field>
        </div>

        <p class="receipt-note">A receipt number will be generated automatically on save.</p>

        <mat-form-field appearance="outline">
          <mat-label>Notes</mat-label>
          <textarea matInput formControlName="notes" rows="2"></textarea>
        </mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Cancel</button>
      <button
        mat-raised-button
        color="primary"
        [disabled]="form.invalid || !selectedLease()"
        (click)="submit()"
      >
        Create
      </button>
    </mat-dialog-actions>
  `,
  styles: [
    `
      .unit-context-bar {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 13px;
        color: #555;
        background: var(--bg-secondary, #e8f5e9);
        border-radius: 6px;
        padding: 6px 16px 8px;
        margin: -8px 16px 0;
      }
      .unit-context-bar mat-icon {
        font-size: 16px;
        width: 16px;
        height: 16px;
        color: #2196f3;
      }
      .ctx-sep {
        color: #aaa;
      }
      .dialog-form {
        display: flex;
        flex-direction: column;
        gap: 6px;
        width: 100%;
      }
      .row-2 {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
      }
      @media (max-width: 480px) {
        .row-2 {
          grid-template-columns: 1fr;
        }
      }
      .lease-option-main {
        display: block;
        font-weight: 500;
      }
      .lease-option-sub {
        display: block;
        font-size: 12px;
        color: #888;
      }
      .lease-summary {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 13px;
        color: #555;
        background: var(--bg-secondary, #f5f5f5);
        border-radius: 6px;
        padding: 8px 10px;
        margin-top: -4px;
      }
      .lease-summary .rent {
        margin-left: auto;
        font-weight: 600;
        color: #2196f3;
      }
      .balance-hint {
        font-size: 12px;
        margin-top: -6px;
        padding: 4px 10px;
      }
      .balance-hint--due {
        color: #e65100;
      }
      .balance-hint--clear {
        color: #888;
      }
      .sep {
        color: #aaa;
      }
      .search-loading {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 12px;
        color: #999;
        padding: 6px 10px;
        margin-top: -4px;
      }
      .receipt-note {
        font-size: 12px;
        color: #888;
        margin: -8px 0 0;
      }
    `,
  ],
})
export class PaymentCreateDialog implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<PaymentCreateDialog>);
  private paymentService = inject(PaymentService);
  readonly prefill: PaymentCreatePrefill | null = inject(MAT_DIALOG_DATA, { optional: true });

  filteredLeases = signal<LeaseSearchResult[]>([]);
  selectedLease = signal<LeaseSearchResult | null>(null);
  searching = signal(false);
  leaseSearch = new FormControl('');

  months = [
    { value: 1, label: 'January' },
    { value: 2, label: 'February' },
    { value: 3, label: 'March' },
    { value: 4, label: 'April' },
    { value: 5, label: 'May' },
    { value: 6, label: 'June' },
    { value: 7, label: 'July' },
    { value: 8, label: 'August' },
    { value: 9, label: 'September' },
    { value: 10, label: 'October' },
    { value: 11, label: 'November' },
    { value: 12, label: 'December' },
  ];
  years = Array.from({ length: 10 }, (_, i) => new Date().getFullYear() - i);

  form: FormGroup = this.fb.group({
    unit_id: [null, [Validators.required, Validators.min(1)]],
    tenant_id: [null, [Validators.required, Validators.min(1)]],
    building_id: [null, [Validators.required, Validators.min(1)]],
    property_id: [null, [Validators.required, Validators.min(1)]],
    month: [new Date().getMonth() + 1, Validators.required],
    year: [new Date().getFullYear(), Validators.required],
    amount_due: [null, [Validators.required, Validators.min(0.01)]],
    due_date: [''],
    payment_method: [''],
    amount_paid: [null, [Validators.min(0)]],
    payment_date: [new Date().toISOString().slice(0, 10)],
    notes: [''],
  });

  ngOnInit(): void {
    // Lease pre-fetched by caller — skip search, apply immediately and lock the field
    if (this.prefill?.lease) {
      this.onLeaseSelected(this.prefill.lease);
      this.leaseSearch.setValue(this.leaseDisplay(this.prefill.lease), { emitEvent: false });
      this.leaseSearch.disable({ emitEvent: false });
      return;
    }

    // Normal search flow
    this.leaseSearch.valueChanges
      .pipe(
        debounceTime(300),
        switchMap((query) => {
          if (!query || typeof query !== 'string' || !query.trim()) {
            return of({ results: [], total: 0 });
          }
          this.searching.set(true);
          return this.paymentService.searchLeases(query.trim());
        })
      )
      .subscribe({
        next: (res) => {
          this.filteredLeases.set(res.results);
          this.searching.set(false);
        },
        error: () => {
          this.filteredLeases.set([]);
          this.searching.set(false);
        },
      });

    // Clear selected lease if user edits the search field manually
    this.leaseSearch.valueChanges.subscribe((v) => {
      if (typeof v === 'string') {
        this.selectedLease.set(null);
        this.form.patchValue({
          unit_id: null,
          tenant_id: null,
          building_id: null,
          property_id: null,
        });
      }
    });
  }

  leaseDisplay = (lease: LeaseSearchResult | string | null): string => {
    if (!lease || typeof lease === 'string') return typeof lease === 'string' ? lease : '';
    return `${lease.tenant_name} — Unit ${lease.unit_number}`;
  };

  onLeaseSelected(lease: LeaseSearchResult): void {
    this.selectedLease.set(lease);
    const amountDue =
      lease.outstanding_balance > 0 ? lease.outstanding_balance : lease.monthly_rent;
    this.form.patchValue({
      unit_id: lease.unit_id,
      tenant_id: lease.tenant_id,
      building_id: lease.building_id,
      property_id: lease.property_id,
      amount_due: amountDue,
    });
  }

  submit(): void {
    if (this.form.valid && this.selectedLease()) {
      const val = this.form.value;
      const req: CreatePaymentRequest = {
        unit_id: val.unit_id,
        tenant_id: val.tenant_id,
        building_id: val.building_id,
        property_id: val.property_id,
        month: val.month,
        year: val.year,
        amount_due: val.amount_due,
        due_date: val.due_date || undefined,
        payment_method: val.payment_method || undefined,
        amount_paid: val.amount_paid != null ? val.amount_paid : undefined,
        // Only meaningful once money has actually been recorded — sending
        // today's date alongside a $0/unset amount_paid would misleadingly
        // mark a still-unpaid Due record as "paid today".
        payment_date: val.amount_paid > 0 ? val.payment_date || undefined : undefined,
        notes: val.notes || undefined,
      };
      this.dialogRef.close(req);
    }
  }
}

// ── Generate Monthly Payments Dialog ──────────────────────────────────────────
@Component({
  selector: 'app-generate-payments-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
  ],
  template: `
    <h2 mat-dialog-title>Generate Monthly Payments</h2>
    <mat-dialog-content>
      <p class="info-text">
        Creates a <strong>Due</strong> payment record for every active lease in the selected month.
        Leases that already have a payment record are skipped.
      </p>
      <form [formGroup]="form" class="dialog-form">
        <div class="row-2">
          <mat-form-field appearance="outline">
            <mat-label>Month</mat-label>
            <mat-select formControlName="month">
              @for (m of months; track m.value) {
                <mat-option [value]="m.value">{{ m.label }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>Year</mat-label>
            <mat-select formControlName="year">
              @for (y of years; track y) {
                <mat-option [value]="y">{{ y }}</mat-option>
              }
            </mat-select>
          </mat-form-field>
        </div>
        <mat-form-field appearance="outline">
          <mat-label>Due Day of Month</mat-label>
          <input matInput type="number" formControlName="due_day_of_month" min="1" max="28" />
          <mat-hint>Day of the month rent is due (1–28, default 7)</mat-hint>
        </mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Cancel</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid" (click)="submit()">
        <mat-icon>bolt</mat-icon>
        Generate
      </button>
    </mat-dialog-actions>
  `,
  styles: [
    `
      .info-text {
        font-size: 13px;
        color: var(--text-secondary, #666);
        margin: 0 0 16px;
        line-height: 1.5;
      }
      .dialog-form {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }
      .row-2 {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
      }
      @media (max-width: 480px) {
        .row-2 {
          grid-template-columns: 1fr;
        }
      }
    `,
  ],
})
export class GeneratePaymentsDialog {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<GeneratePaymentsDialog>);

  months = [
    { value: 1, label: 'January' },
    { value: 2, label: 'February' },
    { value: 3, label: 'March' },
    { value: 4, label: 'April' },
    { value: 5, label: 'May' },
    { value: 6, label: 'June' },
    { value: 7, label: 'July' },
    { value: 8, label: 'August' },
    { value: 9, label: 'September' },
    { value: 10, label: 'October' },
    { value: 11, label: 'November' },
    { value: 12, label: 'December' },
  ];
  years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() + 1 - i);

  form: FormGroup = this.fb.group({
    month: [new Date().getMonth() + 1, Validators.required],
    year: [new Date().getFullYear(), Validators.required],
    due_day_of_month: [7, [Validators.min(1), Validators.max(28)]],
  });

  submit(): void {
    if (this.form.valid) {
      const { month, year, due_day_of_month } = this.form.value;
      const req: GenerateMonthlyPaymentsRequest = { month, year, due_day_of_month };
      this.dialogRef.close(req);
    }
  }
}

// ── Update Dialog ──────────────────────────────────────────────────────────────
// Recording money received always creates a new transaction (accumulates on
// top of prior installments); everything else (payment method/date/notes
// with no amount) is a metadata-only correction with no money involved.
export type PaymentUpdateResult =
  | { kind: 'transaction'; req: CreatePaymentTransactionRequest }
  | { kind: 'metadata'; req: UpdatePaymentRequest };

@Component({
  selector: 'app-payment-update-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatTooltipModule,
    MatProgressSpinnerModule,
  ],
  template: `
    <h2 mat-dialog-title>Update Payment</h2>
    <mat-dialog-content>
      <div class="payment-info">
        <span class="info-label">Tenant:</span> {{ data.tenant_name }}
        &nbsp;|&nbsp;
        <span class="info-label">Unit:</span> {{ data.unit_number }}
        &nbsp;|&nbsp;
        <span class="info-label">Period:</span> {{ data.month }}/{{ data.year }}
      </div>
      <div class="amount-info">
        <span
          >Due: <strong>৳{{ data.amount_due | number: '1.2-2' }}</strong></span
        >
        &nbsp;&nbsp;
        <span
          >Paid: <strong>৳{{ data.amount_paid | number: '1.2-2' }}</strong></span
        >
        &nbsp;&nbsp;
        <span
          >Status:
          <span class="inline-badge" [class]="'status-' + computedStatus().toLowerCase()">{{
            computedStatus()
          }}</span></span
        >
      </div>
      <p class="status-note">
        Status is calculated automatically from the amount paid — it can't be set directly.
      </p>
      <form [formGroup]="form" class="dialog-form">
        <mat-form-field appearance="outline">
          <mat-label>Amount to Record (BDT)</mat-label>
          <input
            matInput
            type="number"
            formControlName="amount_to_add"
            step="0.01"
            [max]="remainingDue"
          />
          @if (form.get('amount_to_add')?.hasError('max')) {
            <mat-error
              >Cannot exceed the remaining due of ৳{{ remainingDue | number: '1.2-2' }}</mat-error
            >
          } @else {
            <mat-hint>Adds a new installment on top of the amount already paid</mat-hint>
          }
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>Payment Method</mat-label>
          <mat-select formControlName="payment_method">
            <mat-option value="Cash">Cash</mat-option>
            <mat-option value="Bank Transfer">Bank Transfer</mat-option>
            <mat-option value="bKash">bKash</mat-option>
            <mat-option value="Nagad">Nagad</mat-option>
            <mat-option value="Cheque">Cheque</mat-option>
          </mat-select>
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>Payment Date</mat-label>
          <input matInput type="date" formControlName="payment_date" />
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>Notes</mat-label>
          <textarea matInput formControlName="notes" rows="3"></textarea>
        </mat-form-field>
        <p class="receipt-note">
          @if (data.receipt_number) {
            Receipt No: <strong>{{ data.receipt_number }}</strong>
          } @else {
            A receipt number will be generated automatically once this payment is paid.
          }
        </p>
      </form>

      <div class="history-section">
        <h3 class="history-title">Payment History</h3>
        @if (loadingTransactions()) {
          <mat-spinner diameter="20"></mat-spinner>
        } @else if (transactions().length === 0) {
          <p class="history-empty">No installments recorded yet.</p>
        } @else {
          @for (txn of transactions(); track txn.id) {
            <div class="txn-row">
              <div class="txn-summary">
                <span class="txn-amount">৳{{ txn.amount | number: '1.2-2' }}</span>
                <span class="txn-meta"
                  >{{ txn.payment_date }}
                  @if (txn.payment_method) {
                    · {{ txn.payment_method }}
                  }
                  @if (txn.receipt_number) {
                    · {{ txn.receipt_number }}
                  }
                </span>
              </div>
              <div class="txn-actions">
                <button
                  mat-icon-button
                  type="button"
                  matTooltip="Attachments"
                  (click)="toggleAttachments(txn.id)"
                >
                  <mat-icon>attach_file</mat-icon>
                  @if (attachmentCount(txn.id) > 0) {
                    <span class="attachment-count">{{ attachmentCount(txn.id) }}</span>
                  }
                </button>
                <button
                  mat-icon-button
                  type="button"
                  matTooltip="Attach a file"
                  [disabled]="uploadingTxnId() === txn.id"
                  (click)="fileInput.click()"
                >
                  <mat-icon>upload_file</mat-icon>
                </button>
                <input
                  #fileInput
                  type="file"
                  hidden
                  accept="image/jpeg,image/png,image/gif,image/webp,application/pdf"
                  (change)="onFileSelected(txn.id, $event, fileInput)"
                />
              </div>
            </div>
            @if (expandedTxnId() === txn.id) {
              <div class="attachment-list">
                @for (att of attachmentsFor(txn.id); track att.id) {
                  <div class="attachment-row">
                    <mat-icon class="attachment-icon">{{
                      att.content_type === 'application/pdf' ? 'picture_as_pdf' : 'image'
                    }}</mat-icon>
                    <span class="attachment-name">{{ att.file_name }}</span>
                    <span class="attachment-size">{{ formatFileSize(att.file_size) }}</span>
                    <button
                      mat-icon-button
                      type="button"
                      matTooltip="Download"
                      (click)="downloadAttachment(txn.id, att)"
                    >
                      <mat-icon>download</mat-icon>
                    </button>
                    <button
                      mat-icon-button
                      type="button"
                      matTooltip="Delete"
                      (click)="deleteAttachment(txn.id, att)"
                    >
                      <mat-icon>delete</mat-icon>
                    </button>
                  </div>
                } @empty {
                  <p class="history-empty">No attachments for this installment yet.</p>
                }
              </div>
            }
          }
        }
      </div>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Cancel</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid" (click)="submit()">
        Update
      </button>
    </mat-dialog-actions>
  `,
  styles: [
    `
      .dialog-form {
        display: flex;
        flex-direction: column;
        gap: 4px;
        width: 100%;
        margin-top: 12px;
      }
      .payment-info {
        font-size: 14px;
        color: var(--text-secondary, #666);
        margin-bottom: 8px;
      }
      .amount-info {
        font-size: 14px;
        margin-bottom: 12px;
        display: flex;
        align-items: center;
      }
      .info-label {
        font-weight: 500;
      }
      .status-note {
        font-size: 12px;
        color: var(--text-secondary, #666);
        margin: 0 0 12px;
      }
      .receipt-note {
        font-size: 12px;
        color: var(--text-secondary, #666);
        margin: 4px 0 0;
      }
      .inline-badge {
        display: inline-block;
        padding: 2px 8px;
        border-radius: 4px;
        font-size: 11px;
        font-weight: 600;
        letter-spacing: 0.3px;
        white-space: nowrap;
        margin-left: 4px;

        &.status-paid {
          background: rgba(76, 175, 80, 0.12);
          color: var(--color-paid, #4caf50);
        }
        &.status-due {
          background: rgba(144, 164, 174, 0.12);
          color: var(--color-due, #90a4ae);
        }
        &.status-partial {
          background: rgba(255, 152, 0, 0.12);
          color: var(--color-pending, #ff9800);
        }
        &.status-overdue {
          background: rgba(244, 67, 54, 0.12);
          color: var(--color-overdue, #f44336);
        }
      }
      .history-section {
        margin-top: 16px;
        padding-top: 12px;
        border-top: 1px solid var(--border-color, #e0e0e0);
      }
      .history-title {
        font-size: 13px;
        font-weight: 600;
        margin: 0 0 8px;
        color: var(--text-secondary, #666);
      }
      .history-empty {
        font-size: 12px;
        color: var(--text-secondary, #666);
        margin: 4px 0;
      }
      .txn-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 4px 0;
        border-bottom: 1px solid var(--border-color, #f0f0f0);
      }
      .txn-summary {
        display: flex;
        flex-direction: column;
        font-size: 13px;
      }
      .txn-amount {
        font-weight: 600;
      }
      .txn-meta {
        font-size: 11px;
        color: var(--text-secondary, #666);
      }
      .txn-actions {
        display: flex;
        align-items: center;
        position: relative;
      }
      .attachment-count {
        position: absolute;
        top: 2px;
        right: 2px;
        background: var(--color-primary, #1e88e5);
        color: #fff;
        font-size: 9px;
        line-height: 1;
        border-radius: 8px;
        padding: 2px 4px;
        min-width: 12px;
        text-align: center;
      }
      .attachment-list {
        padding: 4px 0 8px 8px;
      }
      .attachment-row {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 12px;
        padding: 2px 0;
      }
      .attachment-icon {
        font-size: 18px;
        width: 18px;
        height: 18px;
        color: var(--text-secondary, #666);
      }
      .attachment-name {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .attachment-size {
        color: var(--text-secondary, #666);
        font-size: 11px;
      }
    `,
  ],
})
export class PaymentUpdateDialog {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<PaymentUpdateDialog>);
  private paymentService = inject(PaymentService);
  private snackBar = inject(MatSnackBar);
  readonly data: PaymentWithDetails = inject(MAT_DIALOG_DATA);

  transactions = signal<PaymentTransaction[]>([]);
  loadingTransactions = signal(false);
  attachmentsByTxn = signal<Record<number, PaymentTransactionAttachment[]>>({});
  expandedTxnId = signal<number | null>(null);
  uploadingTxnId = signal<number | null>(null);

  constructor() {
    this.loadTransactions();
  }

  // The backend rejects an installment that would push amount_paid past
  // amount_due — recording a genuine advance/next-month payment belongs on
  // its own payment period instead.
  readonly remainingDue = Math.max(this.data.amount_due - this.data.amount_paid, 0);

  form: FormGroup = this.fb.group({
    // Defaults to the remaining balance, not the amount already paid — this
    // field is money to add now, via a new transaction, not the new total.
    amount_to_add: [this.remainingDue, [Validators.min(0), Validators.max(this.remainingDue)]],
    payment_method: [this.data.payment_method ?? ''],
    payment_date: [
      this.data.payment_date
        ? this.data.payment_date.slice(0, 10)
        : new Date().toISOString().slice(0, 10),
    ],
    notes: [this.data.notes ?? ''],
  });

  // Mirrors the backend's derivation (PaymentRepository.Update/Create) so the
  // dialog previews the status the server will actually compute, instead of
  // letting the user pick one that might not match.
  computedStatus(): PaymentStatus {
    const amountToAdd = Number(this.form.get('amount_to_add')?.value) || 0;
    const amountPaid = this.data.amount_paid + amountToAdd;
    const amountDue = this.data.amount_due;
    if (amountDue > 0 && amountPaid >= amountDue) return 'Paid';
    if (amountPaid > 0) return 'Partial';
    if (this.data.due_date && new Date(this.data.due_date) < this.todayMidnight()) return 'Overdue';
    return 'Due';
  }

  private todayMidnight(): Date {
    const d = new Date();
    d.setHours(0, 0, 0, 0);
    return d;
  }

  submit(): void {
    const raw = this.form.value;
    const amount = raw.amount_to_add !== null && raw.amount_to_add !== '' ? +raw.amount_to_add : 0;

    if (amount > 0) {
      const req: CreatePaymentTransactionRequest = { amount };
      if (raw.payment_method) req.payment_method = raw.payment_method;
      if (raw.payment_date) req.payment_date = raw.payment_date;
      if (raw.notes) req.notes = raw.notes;
      this.dialogRef.close({ kind: 'transaction', req } as PaymentUpdateResult);
      return;
    }

    // No money recorded — treat any filled fields as a metadata-only
    // correction (e.g. fixing a typo'd payment method or note).
    const req: UpdatePaymentRequest = {};
    if (raw.payment_method) req.payment_method = raw.payment_method;
    if (raw.payment_date) req.payment_date = raw.payment_date;
    if (raw.notes) req.notes = raw.notes;
    if (Object.keys(req).length === 0) {
      this.dialogRef.close();
      return;
    }
    this.dialogRef.close({ kind: 'metadata', req } as PaymentUpdateResult);
  }

  private loadTransactions(): void {
    this.loadingTransactions.set(true);
    this.paymentService.getPaymentTransactions(this.data.id).subscribe({
      next: (txns) => {
        this.transactions.set(txns);
        this.loadingTransactions.set(false);
        // Loaded eagerly (rather than only on expand) so the attachment
        // count badge is accurate before the user opens any row.
        txns.forEach((txn) => this.loadAttachments(txn.id));
      },
      error: () => {
        this.loadingTransactions.set(false);
      },
    });
  }

  private loadAttachments(transactionId: number): void {
    this.paymentService.getPaymentTransactionAttachments(this.data.id, transactionId).subscribe({
      next: (atts) => {
        this.attachmentsByTxn.update((map) => ({ ...map, [transactionId]: atts }));
      },
    });
  }

  attachmentsFor(transactionId: number): PaymentTransactionAttachment[] {
    return this.attachmentsByTxn()[transactionId] ?? [];
  }

  attachmentCount(transactionId: number): number {
    return this.attachmentsFor(transactionId).length;
  }

  toggleAttachments(transactionId: number): void {
    this.expandedTxnId.set(this.expandedTxnId() === transactionId ? null : transactionId);
  }

  onFileSelected(transactionId: number, event: Event, fileInput: HTMLInputElement): void {
    const file = (event.target as HTMLInputElement).files?.[0];
    fileInput.value = ''; // allow re-selecting the same file later
    if (!file) return;

    if (file.size > MAX_ATTACHMENT_FILE_SIZE) {
      notifyError(this.snackBar, 'File exceeds the maximum allowed size of 10MB');
      return;
    }
    if (!ALLOWED_ATTACHMENT_CONTENT_TYPES.includes(file.type)) {
      notifyError(this.snackBar, 'Only images and PDF files are allowed');
      return;
    }

    this.uploadingTxnId.set(transactionId);
    this.paymentService
      .uploadPaymentTransactionAttachment(this.data.id, transactionId, file)
      .subscribe({
        next: (attachment) => {
          this.attachmentsByTxn.update((map) => ({
            ...map,
            [transactionId]: [...(map[transactionId] ?? []), attachment],
          }));
          this.expandedTxnId.set(transactionId);
          this.uploadingTxnId.set(null);
          notifySuccess(this.snackBar, 'Attachment uploaded');
        },
        error: (err) => {
          this.uploadingTxnId.set(null);
          notifyError(this.snackBar, err?.error?.error ?? 'Failed to upload attachment');
        },
      });
  }

  downloadAttachment(transactionId: number, attachment: PaymentTransactionAttachment): void {
    this.paymentService
      .downloadPaymentTransactionAttachment(this.data.id, transactionId, attachment.id)
      .subscribe({
        next: (blob) => {
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = attachment.file_name;
          a.click();
          URL.revokeObjectURL(url);
        },
        error: () => notifyError(this.snackBar, 'Failed to download attachment'),
      });
  }

  deleteAttachment(transactionId: number, attachment: PaymentTransactionAttachment): void {
    this.paymentService
      .deletePaymentTransactionAttachment(this.data.id, transactionId, attachment.id)
      .subscribe({
        next: () => {
          this.attachmentsByTxn.update((map) => ({
            ...map,
            [transactionId]: (map[transactionId] ?? []).filter((a) => a.id !== attachment.id),
          }));
          notifySuccess(this.snackBar, 'Attachment deleted');
        },
        error: () => notifyError(this.snackBar, 'Failed to delete attachment'),
      });
  }

  formatFileSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
}
