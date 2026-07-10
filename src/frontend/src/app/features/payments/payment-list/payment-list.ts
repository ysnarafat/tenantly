import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
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
        this.showError('Failed to load payments');
        console.error(err);
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
        this.showError('Failed to load tree view');
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
              this.showError(msg);
            } else {
              this.showSuccess(msg);
            }
            this.loadPayments();
            this.loadSummary();
            if (this.activeTab() === 1) this.loadTreePayments();
          },
          error: (err) => {
            this.loading.set(false);
            this.showError(err?.error?.error ?? 'Failed to generate payments');
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
          next: () => {
            this.showSuccess('Payment created');
            this.loadPayments();
            this.loadSummary();
            if (this.activeTab() === 1) this.loadTreePayments();
          },
          error: (err) => this.showError(err?.error?.error ?? 'Failed to create payment'),
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
    ref.afterClosed().subscribe((req: UpdatePaymentRequest | undefined) => {
      if (req) {
        this.paymentService.updatePayment(payment.id, req).subscribe({
          next: () => {
            this.showSuccess('Payment updated');
            this.loadPayments();
            this.loadSummary();
            if (this.activeTab() === 1) this.loadTreePayments();
          },
          error: (err) => this.showError(err?.error?.error ?? 'Failed to update payment'),
        });
      }
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

  private showSuccess(msg: string): void {
    this.snackBar.open(msg, 'Close', { duration: 3000, panelClass: 'snack-success' });
  }

  private showError(msg: string): void {
    this.snackBar.open(msg, 'Close', { duration: 5000, panelClass: 'snack-error' });
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

        <mat-form-field appearance="outline">
          <mat-label>Receipt Number</mat-label>
          <input matInput formControlName="receipt_number" />
        </mat-form-field>

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
    payment_date: [''],
    receipt_number: [''],
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
    this.form.patchValue({
      unit_id: lease.unit_id,
      tenant_id: lease.tenant_id,
      building_id: lease.building_id,
      property_id: lease.property_id,
      amount_due: lease.monthly_rent,
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
        payment_date: val.payment_date || undefined,
        receipt_number: val.receipt_number || undefined,
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
      </div>
      <form [formGroup]="form" class="dialog-form">
        <mat-form-field appearance="outline">
          <mat-label>Amount Paid (BDT)</mat-label>
          <input matInput type="number" formControlName="amount_paid" step="0.01" />
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>Status</mat-label>
          <mat-select formControlName="status">
            <mat-option value="Due">Due</mat-option>
            <mat-option value="Partial">Partial</mat-option>
            <mat-option value="Paid">Paid</mat-option>
            <mat-option value="Overdue">Overdue</mat-option>
          </mat-select>
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
          <mat-label>Receipt Number</mat-label>
          <input matInput formControlName="receipt_number" />
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>Notes</mat-label>
          <textarea matInput formControlName="notes" rows="3"></textarea>
        </mat-form-field>
      </form>
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
      }
      .info-label {
        font-weight: 500;
      }
    `,
  ],
})
export class PaymentUpdateDialog {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<PaymentUpdateDialog>);
  readonly data: PaymentWithDetails = inject(MAT_DIALOG_DATA);

  form: FormGroup = this.fb.group({
    amount_paid: [this.data.amount_paid, [Validators.min(0)]],
    status: [this.data.status],
    payment_method: [this.data.payment_method ?? ''],
    payment_date: [this.data.payment_date ? this.data.payment_date.slice(0, 10) : ''],
    receipt_number: [this.data.receipt_number ?? ''],
    notes: [this.data.notes ?? ''],
  });

  submit(): void {
    const raw = this.form.value;
    const req: UpdatePaymentRequest = {};
    if (raw.amount_paid !== null && raw.amount_paid !== '') req.amount_paid = +raw.amount_paid;
    if (raw.status) req.status = raw.status;
    if (raw.payment_method) req.payment_method = raw.payment_method;
    if (raw.payment_date) req.payment_date = raw.payment_date;
    if (raw.receipt_number) req.receipt_number = raw.receipt_number;
    if (raw.notes) req.notes = raw.notes;
    this.dialogRef.close(req);
  }
}
