import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule, JsonPipe } from '@angular/common';
import {
  FormsModule,
  ReactiveFormsModule,
  FormBuilder,
  FormGroup,
  Validators,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatTabsModule } from '@angular/material/tabs';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatTableModule } from '@angular/material/table';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatGridListModule } from '@angular/material/grid-list';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatDialog } from '@angular/material/dialog';
import { TranslateModule } from '@ngx-translate/core';
import { PermissionService } from '../../core/services/permission.service';
import {
  ReportService,
  DashboardMetrics,
  CollectionSummaryReport,
  PaymentAnalysisReport,
  FinancialLedgerReport,
  TenantSummaryReport,
  PropertyAnalyticsReport,
} from '../../core/services/report.service';
import { PaymentService } from '../../core/services/payment.service';
import { BuildingService } from '../../core/services/building.service';
import { PropertyService } from '../../core/services/property.service';
import { Building, Property } from '../../core/models';

export interface ReportTemplate {
  id: string;
  name: string;
  description: string;
  icon: string;
  category: 'financial' | 'operational' | 'tenant' | 'collections';
  requires: string[];
}

export interface QuickMetric {
  label: string;
  value: number | string;
  trend?: number;
  unit?: string;
  icon: string;
}

function dateRangeValidator(fg: AbstractControl): ValidationErrors | null {
  const start = fg.get('startDate')?.value;
  const end = fg.get('endDate')?.value;
  if (start && end && new Date(end) < new Date(start)) {
    return { endBeforeStart: true };
  }
  return null;
}

@Component({
  selector: 'app-report-analysis',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatMenuModule,
    MatTabsModule,
    MatDatepickerModule,
    MatNativeDateModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatTableModule,
    MatProgressBarModule,
    MatChipsModule,
    MatGridListModule,
    MatProgressSpinnerModule,
    MatTooltipModule,
    TranslateModule,
    JsonPipe,
  ],
  templateUrl: './report-analysis.html',
  styleUrls: ['./report-analysis.scss'],
})
export class ReportAnalysis implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  private cancelPending$ = new Subject<void>();
  private fb = inject(FormBuilder);
  private permissionService = inject(PermissionService);
  private snackBar = inject(MatSnackBar);
  private dialog = inject(MatDialog);
  private reportService = inject(ReportService);
  private paymentService = inject(PaymentService);
  private buildingService = inject(BuildingService);
  private propertyService = inject(PropertyService);

  reportForm!: FormGroup;
  selectedReport: ReportTemplate | null = null;
  loading = false;
  generatingReport = false;
  loadingMetrics = false;

  activeTab = 0;

  // Report data
  dashboardMetrics: DashboardMetrics | null = null;
  collectionReport: CollectionSummaryReport | null = null;
  paymentReport: PaymentAnalysisReport | null = null;
  ledgerReport: FinancialLedgerReport | null = null;
  tenantReport: TenantSummaryReport | null = null;
  propertyAnalyticsReport: PropertyAnalyticsReport | null = null;
  buildingReport: unknown | null = null;

  buildings: Building[] = [];
  properties: Property[] = [];

  reports: ReportTemplate[] = [
    {
      id: 'ledger',
      name: 'Financial Ledger',
      description: 'Complete transaction history with balances',
      icon: 'receipt_long',
      category: 'financial',
      requires: ['org_id'],
    },
    {
      id: 'collection_summary',
      name: 'Collection Summary',
      description: 'Collection rates, aging analysis, trends',
      icon: 'trending_up',
      category: 'collections',
      requires: ['org_id', 'date_range'],
    },
    {
      id: 'property_analytics',
      name: 'Property Analytics',
      description: 'Performance by property, occupancy, revenue',
      icon: 'domain',
      category: 'operational',
      requires: ['org_id', 'date_range'],
    },
    {
      id: 'tenant_report',
      name: 'Tenant Report',
      description: 'Tenant information, lease details, contact',
      icon: 'people',
      category: 'tenant',
      requires: ['org_id'],
    },
    {
      id: 'building_performance',
      name: 'Building Performance',
      description: 'Unit occupancy, maintenance, revenue by building',
      icon: 'apartment',
      category: 'operational',
      requires: ['org_id', 'date_range'],
    },
    {
      id: 'payment_analysis',
      name: 'Payment Analysis',
      description: 'Payment methods, channels, trends',
      icon: 'payments',
      category: 'financial',
      requires: ['org_id', 'date_range'],
    },
  ];

  get quickMetrics(): QuickMetric[] {
    const c = this.dashboardMetrics?.collection_summary;
    const p = this.dashboardMetrics?.payment_analysis;
    return [
      {
        label: 'Total Revenue',
        value: c ? `৳ ${c.total_collected.toLocaleString()}` : '—',
        icon: 'attach_money',
      },
      {
        label: 'Collection Rate',
        value: c ? `${c.collection_rate.toFixed(1)}%` : '—',
        icon: 'percent',
      },
      {
        label: 'Outstanding Due',
        value: c ? `৳ ${(c.total_due - c.total_collected).toLocaleString()}` : '—',
        icon: 'warning',
      },
      {
        label: 'Total Payments',
        value: p ? p.total_payments.toString() : '—',
        icon: 'payments',
      },
    ];
  }

  ngOnInit() {
    this.initForm();
    this.loadDashboardMetrics();
    this.loadBuildings();
    this.loadProperties();
  }

  loadBuildings() {
    this.buildingService.getBuildings({ active: true }).subscribe({
      next: (res) => {
        this.buildings = res.buildings ?? [];
      },
      error: () => {},
    });
  }

  loadProperties() {
    this.propertyService.getProperties({ active: true }).subscribe({
      next: (res) => {
        this.properties = res.properties ?? [];
      },
      error: () => {},
    });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  initForm() {
    this.reportForm = this.fb.group(
      {
        reportType: ['ledger', Validators.required],
        propertyFilter: [''],
        buildingFilter: [''],
        startDate: [''],
        endDate: [''],
        exportFormat: ['pdf'],
      },
      { validators: dateRangeValidator }
    );
  }

  loadDashboardMetrics() {
    this.loadingMetrics = true;
    this.reportService.getDashboardMetrics().subscribe({
      next: (metrics) => {
        this.dashboardMetrics = metrics;
        this.collectionReport = metrics.collection_summary;
        this.paymentReport = metrics.payment_analysis;
        this.loadingMetrics = false;
      },
      error: (error) => {
        console.error('Error loading dashboard metrics:', error);
        this.snackBar.open('Error loading report data', 'Close', { duration: 3000 });
        this.loadingMetrics = false;
      },
    });
  }

  loadLedgerReport() {
    const filters: Record<string, any> = {};
    if (this.reportForm.get('startDate')?.value) {
      const d = new Date(this.reportForm.get('startDate')!.value);
      filters['month'] = d.getMonth() + 1;
      filters['year'] = d.getFullYear();
    }

    this.generatingReport = true;
    this.reportService
      .getFinancialLedger(1, 50, filters)
      .pipe(takeUntil(this.cancelPending$), takeUntil(this.destroy$))
      .subscribe({
        next: (report) => {
          this.ledgerReport = report;
          this.generatingReport = false;
          this.snackBar.open('Ledger report generated', 'Close', { duration: 2000 });
        },
        error: (error) => {
          console.error('Error generating ledger report:', error);
          this.snackBar.open('Error generating report', 'Close', { duration: 3000 });
          this.generatingReport = false;
        },
      });
  }

  loadTenantSummary() {
    this.generatingReport = true;
    this.reportService
      .getTenantSummary()
      .pipe(takeUntil(this.cancelPending$), takeUntil(this.destroy$))
      .subscribe({
        next: (report) => {
          this.tenantReport = report;
          this.generatingReport = false;
          this.snackBar.open('Tenant report generated', 'Close', { duration: 2000 });
        },
        error: (error) => {
          console.error('Error generating tenant report:', error);
          this.snackBar.open('Error generating report', 'Close', { duration: 3000 });
          this.generatingReport = false;
        },
      });
  }

  loadPropertyAnalytics() {
    const startDate = this.reportForm.get('startDate')?.value;
    const endDate = this.reportForm.get('endDate')?.value;
    const start = startDate ? new Date(startDate).toISOString().split('T')[0] : undefined;
    const end = endDate ? new Date(endDate).toISOString().split('T')[0] : undefined;

    this.generatingReport = true;
    this.reportService
      .getPropertyAnalytics(start, end)
      .pipe(takeUntil(this.cancelPending$), takeUntil(this.destroy$))
      .subscribe({
        next: (report) => {
          this.propertyAnalyticsReport = report;
          this.generatingReport = false;
          this.snackBar.open('Property analytics generated', 'Close', { duration: 2000 });
        },
        error: (error) => {
          console.error('Error generating property analytics:', error);
          this.snackBar.open('Error generating report', 'Close', { duration: 3000 });
          this.generatingReport = false;
        },
      });
  }

  loadBuildingPerformance() {
    const buildingIdStr = this.reportForm.get('buildingFilter')?.value;
    const buildingId = parseInt(buildingIdStr, 10);
    if (!buildingId) {
      this.snackBar.open('Please select a building first', 'Close', { duration: 3000 });
      return;
    }

    const startDate = this.reportForm.get('startDate')?.value;
    const endDate = this.reportForm.get('endDate')?.value;
    const start = startDate
      ? new Date(startDate).toISOString().split('T')[0]
      : new Date(new Date().setDate(1)).toISOString().split('T')[0];
    const end = endDate
      ? new Date(endDate).toISOString().split('T')[0]
      : new Date().toISOString().split('T')[0];

    this.generatingReport = true;
    this.paymentService
      .getBuildingReport(buildingId, start, end)
      .pipe(takeUntil(this.cancelPending$), takeUntil(this.destroy$))
      .subscribe({
        next: (report) => {
          this.buildingReport = report;
          this.generatingReport = false;
          this.snackBar.open('Building performance report generated', 'Close', { duration: 2000 });
        },
        error: (error) => {
          console.error('Error generating building report:', error);
          this.snackBar.open('Error generating report', 'Close', { duration: 3000 });
          this.generatingReport = false;
        },
      });
  }

  loadCollectionSummary() {
    const startDate = this.reportForm.get('startDate')?.value;
    const endDate = this.reportForm.get('endDate')?.value;

    const start = startDate ? new Date(startDate).toISOString().split('T')[0] : undefined;
    const end = endDate ? new Date(endDate).toISOString().split('T')[0] : undefined;

    this.generatingReport = true;
    this.reportService
      .getCollectionSummary(start, end)
      .pipe(takeUntil(this.cancelPending$), takeUntil(this.destroy$))
      .subscribe({
        next: (report) => {
          this.collectionReport = report;
          this.generatingReport = false;
          this.snackBar.open('Report generated', 'Close', { duration: 2000 });
        },
        error: (error) => {
          console.error('Error generating report:', error);
          this.snackBar.open('Error generating report', 'Close', { duration: 3000 });
          this.generatingReport = false;
        },
      });
  }

  loadPaymentAnalysis() {
    const startDate = this.reportForm.get('startDate')?.value;
    const endDate = this.reportForm.get('endDate')?.value;

    const start = startDate ? new Date(startDate).toISOString().split('T')[0] : undefined;
    const end = endDate ? new Date(endDate).toISOString().split('T')[0] : undefined;

    this.generatingReport = true;
    this.reportService
      .getPaymentAnalysis(start, end)
      .pipe(takeUntil(this.cancelPending$), takeUntil(this.destroy$))
      .subscribe({
        next: (report) => {
          this.paymentReport = report;
          this.generatingReport = false;
          this.snackBar.open('Report generated', 'Close', { duration: 2000 });
        },
        error: (error) => {
          console.error('Error generating report:', error);
          this.snackBar.open('Error generating report', 'Close', { duration: 3000 });
          this.generatingReport = false;
        },
      });
  }

  selectReport(report: ReportTemplate) {
    this.selectedReport = report;
    this.reportForm.patchValue({ reportType: report.id });
  }

  generateReport() {
    if (this.reportForm.hasError('endBeforeStart')) {
      this.snackBar.open('End date must be after start date', 'Close', { duration: 3000 });
      return;
    }
    if (!this.reportForm.valid) {
      this.snackBar.open('Please fill required fields', 'Close', { duration: 3000 });
      return;
    }

    this.cancelPending$.next();
    const reportType = this.reportForm.get('reportType')?.value;

    switch (reportType) {
      case 'ledger':
        this.loadLedgerReport();
        break;
      case 'collection_summary':
        this.loadCollectionSummary();
        break;
      case 'payment_analysis':
        this.loadPaymentAnalysis();
        break;
      case 'tenant_report':
        this.loadTenantSummary();
        break;
      case 'property_analytics':
        this.loadPropertyAnalytics();
        break;
      case 'building_performance':
        this.loadBuildingPerformance();
        break;
      default:
        this.snackBar.open('Report type not yet implemented', 'Close', { duration: 3000 });
    }
  }

  exportReport(format: 'pdf' | 'csv' | 'xlsx') {
    if (format !== 'csv') {
      this.snackBar.open(`${format.toUpperCase()} export coming soon`, 'Close', { duration: 2000 });
      return;
    }

    const reportType = this.selectedReport?.id ?? this.reportForm.get('reportType')?.value;
    let csv = '';
    let filename = 'report.csv';

    if (reportType === 'ledger' && this.ledgerReport?.payments?.length) {
      filename = 'financial-ledger.csv';
      csv = this.toCsv(
        ['Tenant', 'Unit', 'Building', 'Period', 'Due (৳)', 'Paid (৳)', 'Status'],
        this.ledgerReport.payments.map((p: any) => [
          p.tenant_name,
          p.unit_number,
          p.building_name,
          `${p.month}/${p.year}`,
          p.amount_due,
          p.amount_paid,
          p.status,
        ])
      );
    } else if (reportType === 'tenant_report' && this.tenantReport?.tenants?.length) {
      filename = 'tenant-report.csv';
      csv = this.toCsv(
        [
          'Tenant',
          'Phone',
          'Email',
          'Unit',
          'Building',
          'Property',
          'Monthly Rent',
          'Total Due',
          'Total Paid',
          'Balance',
          'Lease Active',
        ],
        this.tenantReport.tenants.map((t) => [
          t.tenant_name,
          t.phone_number,
          t.email,
          t.unit_number,
          t.building_name,
          t.property_name,
          t.monthly_rent,
          t.total_due,
          t.total_paid,
          t.balance_due,
          t.lease_active ? 'Yes' : 'No',
        ])
      );
    } else if (reportType === 'building_performance' && this.buildingReport) {
      filename = 'building-performance.csv';
      const payments = (this.buildingReport as any).payments ?? [];
      csv = this.toCsv(
        ['Tenant', 'Unit', 'Period', 'Due (৳)', 'Paid (৳)', 'Status'],
        payments.map((p: any) => [
          p.tenant_name,
          p.unit_number,
          `${p.month}/${p.year}`,
          p.amount_due,
          p.amount_paid,
          p.status,
        ])
      );
    } else if (reportType === 'collection_summary' && this.collectionReport) {
      filename = 'collection-summary.csv';
      csv = this.toCsv(
        ['Metric', 'Value'],
        [
          ['Collection Rate (%)', this.collectionReport.collection_rate.toFixed(2)],
          ['Total Due', this.collectionReport.total_due],
          ['Total Collected', this.collectionReport.total_collected],
          ['Total Pending', this.collectionReport.total_pending],
          ['Total Overdue', this.collectionReport.total_overdue],
          ['Period', this.collectionReport.report_period],
        ]
      );
    } else if (
      reportType === 'property_analytics' &&
      this.propertyAnalyticsReport?.properties?.length
    ) {
      filename = 'property-analytics.csv';
      csv = this.toCsv(
        ['Property', 'Code', 'Type'],
        this.propertyAnalyticsReport.properties.map((p) => [
          p.property_name,
          p.property_code,
          p.property_type,
        ])
      );
    } else {
      this.snackBar.open('Generate a report first before exporting', 'Close', { duration: 3000 });
      return;
    }

    this.downloadCsv(csv, filename);
    this.snackBar.open('CSV downloaded', 'Close', { duration: 2000 });
  }

  private toCsv(headers: string[], rows: any[][]): string {
    const escape = (v: any) => {
      const s = String(v ?? '');
      return s.includes(',') || s.includes('"') || s.includes('\n')
        ? `"${s.replace(/"/g, '""')}"`
        : s;
    };
    return [headers, ...rows].map((row) => row.map(escape).join(',')).join('\r\n');
  }

  private downloadCsv(csv: string, filename: string): void {
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  }

  getReportsByCategory(category: string) {
    return this.reports.filter((r) => r.category === category);
  }

  canAccessReport(report: ReportTemplate): boolean {
    if (this.permissionService.isAdmin()) return true;
    if (this.permissionService.isPropertyManager() && report.category !== 'financial') return true;
    if (this.permissionService.isAccountant() && report.category === 'financial') return true;
    return false;
  }

  getTrendIcon(trend: number | undefined): string {
    if (!trend) return '';
    return trend > 0 ? 'trending_up' : 'trending_down';
  }

  getTrendClass(trend: number | undefined): string {
    if (!trend) return '';
    return trend > 0 ? 'trend-positive' : 'trend-negative';
  }
}
