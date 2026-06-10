import { Component, OnInit, inject } from '@angular/core';
import { CommonModule, JsonPipe } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
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
import { ReportService, DashboardMetrics, CollectionSummaryReport, PaymentAnalysisReport, FinancialLedgerReport, TenantSummaryReport, PropertyAnalyticsReport } from '../../core/services/report.service';
import { PaymentService } from '../../core/services/payment.service';
import { BuildingService } from '../../core/services/building.service';
import { Building } from '../../core/models';

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
export class ReportAnalysis implements OnInit {
  private fb = inject(FormBuilder);
  private permissionService = inject(PermissionService);
  private snackBar = inject(MatSnackBar);
  private dialog = inject(MatDialog);
  private reportService = inject(ReportService);
  private paymentService = inject(PaymentService);
  private buildingService = inject(BuildingService);

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
  }

  loadBuildings() {
    this.buildingService.getBuildings({ active: true }).subscribe({
      next: (res) => { this.buildings = res.buildings ?? []; },
      error: () => {},
    });
  }

  initForm() {
    this.reportForm = this.fb.group({
      reportType: ['ledger', Validators.required],
      propertyFilter: [''],
      buildingFilter: [''],
      startDate: [''],
      endDate: [''],
      exportFormat: ['pdf'],
    });
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
    this.reportService.getFinancialLedger(1, 50, filters).subscribe({
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
    this.reportService.getTenantSummary().subscribe({
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
    this.reportService.getPropertyAnalytics(start, end).subscribe({
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
    const start = startDate ? new Date(startDate).toISOString().split('T')[0] : new Date(new Date().setDate(1)).toISOString().split('T')[0];
    const end = endDate ? new Date(endDate).toISOString().split('T')[0] : new Date().toISOString().split('T')[0];

    this.generatingReport = true;
    this.paymentService.getBuildingReport(buildingId, start, end).subscribe({
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
    this.reportService.getCollectionSummary(start, end).subscribe({
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
    this.reportService.getPaymentAnalysis(start, end).subscribe({
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
    if (!this.reportForm.valid) {
      this.snackBar.open('Please fill required fields', 'Close', { duration: 3000 });
      return;
    }

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
    this.snackBar.open(`Exporting as ${format.toUpperCase()}...`, 'Close', { duration: 2000 });
    // Call export service
  }

  getReportsByCategory(category: string) {
    return this.reports.filter((r) => r.category === category);
  }

  canAccessReport(report: ReportTemplate): boolean {
    // Role-based access control
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
