import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { TranslateModule } from '@ngx-translate/core';
import { Store } from '@ngrx/store';
import { Observable, forkJoin, of } from 'rxjs';
import { map, switchMap } from 'rxjs/operators';
import { LeaseService } from '../../core/services/lease.service';
import { LeaseWithDetails } from '../../core/models/lease.model';
import { PaymentTransaction } from '../../core/models/payment.model';
import { PaymentService } from '../../core/services/payment.service';
import { Property, PropertyWithStats } from '../../core/models/property.model';
import { PropertyService } from '../../core/services/property.service';
import { PermissionService } from '../../core/services/permission.service';
import { Permission } from '../../core/models/role.model';
import { AttachmentService, Attachment } from '../../core/services/attachment.service';
import { AuthService } from '../../core/services/auth.service';
import { AppState } from '../../store';
import * as AuthSelectors from '../../store/auth/auth.selectors';
import { safeErrorMessage } from '../../shared/utils/error.utils';

interface DashboardStats {
  totalDue: number;
  totalCollected: number;
  pendingPayments: number;
  pendingCount: number;
  overduePayments: number;
  overdueUnitCount: number;
  activeLeases: number;
  expiringLeases: number;
  propertyCount: number;
  totalAttachments: number;
  collectionRate: number;
  occupancyRate: number;
  occupiedUnits: number;
  totalUnits: number;
}

interface MonthlyCollectionPoint {
  key: string;
  year: number;
  label: string;
  totalDue: number;
  totalPaid: number;
  collectionRate: number;
}

interface RecentCollection extends PaymentTransaction {
  tenantName: string;
  propertyName: string;
  unitNumber: string;
  amountDue: number;
  amountPaid: number;
}

interface PropertyLeaseSummary {
  propertyId: number;
  propertyName: string;
  address: string;
  buildingCount: number;
  leaseCount: number;
  monthlyRent: number;
  rentShare: number;
  occupiedUnits: number;
  totalUnits: number;
  vacantUnits: number;
  occupancyRate: number;
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, MatIconModule, MatButtonModule, TranslateModule],
  templateUrl: './dashboard.html',
  styleUrls: ['./dashboard.scss'],
})
export class Dashboard implements OnInit {
  private leaseService = inject(LeaseService);
  private attachmentService = inject(AttachmentService);
  private paymentService = inject(PaymentService);
  private propertyService = inject(PropertyService);
  private permissionService = inject(PermissionService);
  private authService = inject(AuthService);
  private store = inject(Store<AppState>);

  // NgRx selectors for reactive data
  user$ = this.store.select(AuthSelectors.selectUser);
  userRole$ = this.store.select(AuthSelectors.selectUserRole);
  isAdmin$ = this.store.select(AuthSelectors.selectIsAdmin);
  isPropertyManager$ = this.store.select(AuthSelectors.selectIsPropertyManager);
  canViewLeases$ = this.store.select(AuthSelectors.selectCanViewLeases);
  canViewAttachments$ = this.store.select(AuthSelectors.selectCanViewAttachments);
  canViewPayments = this.permissionService.canAccess(Permission.VIEW_PAYMENTS);
  canRecordPayments = this.permissionService.canAccess(Permission.RECORD_PAYMENTS);

  stats: DashboardStats = {
    totalDue: 0,
    totalCollected: 0,
    pendingPayments: 0,
    pendingCount: 0,
    overduePayments: 0,
    overdueUnitCount: 0,
    activeLeases: 0,
    expiringLeases: 0,
    propertyCount: 0,
    totalAttachments: 0,
    collectionRate: 0,
    occupancyRate: 0,
    occupiedUnits: 0,
    totalUnits: 0,
  };

  recentLeases: LeaseWithDetails[] = [];
  expiringLeases: LeaseWithDetails[] = [];
  recentAttachments: Attachment[] = [];
  propertySummaries: PropertyLeaseSummary[] = [];
  private activeLeaseData: LeaseWithDetails[] = [];
  private propertyData: PropertyWithStats[] = [];
  collectionTrend: MonthlyCollectionPoint[] = [];
  recentCollections: RecentCollection[] = [];
  trendRange: 'sixMonths' | 'yearToDate' | 'year' = 'sixMonths';
  readonly currentYear = new Date().getFullYear();
  loading = true;

  ngOnInit() {
    this.loadDashboardData();
  }

  loadDashboardData() {
    this.loading = true;
    this.loadCollectionTrend();
    this.loadPropertyOccupancy();
    this.loadRecentCollections();
    this.loadCurrentMonthPayments();

    // Load leases data
    this.leaseService.getActiveLeases().subscribe({
      next: (leases) => {
        this.processLeaseData(leases);
        this.recentLeases = leases
          .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
          .slice(0, 5);
        this.expiringLeases = leases
          .filter((l) => l.days_remaining <= 30 && l.days_remaining > 0)
          .slice(0, 5);
        this.stats.expiringLeases = leases.filter(
          (l) => l.days_remaining <= 30 && l.days_remaining > 0
        ).length;
      },
      error: (error) => console.error('Error loading leases:', safeErrorMessage(error)),
    });

    // Load attachments data
    this.attachmentService.getAllAttachments().subscribe({
      next: (attachments) => {
        this.processAttachmentData(attachments);
        this.recentAttachments = attachments
          .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
          .slice(0, 5);
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading attachments:', safeErrorMessage(error));
        this.loading = false;
      },
    });
  }

  processLeaseData(leases: LeaseWithDetails[]) {
    this.activeLeaseData = leases;
    const activeLeases = leases.filter((lease) => lease.active);
    this.stats.activeLeases = activeLeases.length;
    this.rebuildPropertySummaries();
  }

  loadCurrentMonthPayments(): void {
    const now = new Date();
    this.paymentService
      .getAllPayments({ year: now.getFullYear(), month: now.getMonth() + 1 })
      .subscribe({
        next: (payments) => {
          this.stats.totalDue = payments.reduce((sum, payment) => sum + payment.amount_due, 0);
          this.stats.totalCollected = payments.reduce(
            (sum, payment) => sum + payment.amount_paid,
            0
          );
          this.stats.pendingPayments = payments
            .filter((payment) => payment.status === 'Due' || payment.status === 'Partial')
            .reduce(
              (sum, payment) => sum + Math.max(payment.amount_due - payment.amount_paid, 0),
              0
            );
          this.stats.pendingCount = payments.filter(
            (payment) => payment.status === 'Due' || payment.status === 'Partial'
          ).length;
          this.stats.overdueUnitCount = new Set(
            payments
              .filter((payment) => payment.status === 'Overdue')
              .map((payment) => `${payment.property_id}-${payment.unit_id}`)
          ).size;
          this.stats.overduePayments = payments
            .filter((payment) => payment.status === 'Overdue')
            .reduce(
              (sum, payment) => sum + Math.max(payment.amount_due - payment.amount_paid, 0),
              0
            );
          this.stats.collectionRate = this.stats.totalDue
            ? (this.stats.totalCollected / this.stats.totalDue) * 100
            : 0;
        },
        error: (error) =>
          console.error('Error loading current month payments:', safeErrorMessage(error)),
      });
  }

  loadCollectionTrend(): void {
    const now = new Date();
    const firstMonth = this.trendRange === 'sixMonths' ? now.getMonth() - 5 : 0;
    const monthCount = this.trendRange === 'year' ? 12 : now.getMonth() - firstMonth + 1;
    const periods = Array.from({ length: monthCount }, (_, index) => {
      const date = new Date(now.getFullYear(), firstMonth + index, 1);
      return { year: date.getFullYear(), month: date.getMonth() + 1, date };
    });

    forkJoin(
      periods.map(({ year, month }) => this.paymentService.getAllPayments({ year, month }))
    ).subscribe({
      next: (results) => {
        this.collectionTrend = periods.map(({ year, month, date }, index) => {
          const payments = results[index];
          const totalDue = payments.reduce((sum, payment) => sum + payment.amount_due, 0);
          const totalPaid = payments.reduce((sum, payment) => sum + payment.amount_paid, 0);
          return {
            key: `${year}-${month}`,
            year,
            label: new Intl.DateTimeFormat(undefined, { month: 'short' }).format(date),
            totalDue,
            totalPaid,
            collectionRate: totalDue ? (totalPaid / totalDue) * 100 : 0,
          };
        });
      },
      error: (error) => console.error('Error loading collection trend:', safeErrorMessage(error)),
    });
  }

  setTrendRange(range: 'sixMonths' | 'yearToDate' | 'year'): void {
    if (this.trendRange === range) return;
    this.trendRange = range;
    this.loadCollectionTrend();
  }

  get trendTotalCollected(): number {
    return this.collectionTrend.reduce((sum, point) => sum + point.totalPaid, 0);
  }

  get averageCollectionRate(): number {
    const totalDue = this.collectionTrend.reduce((sum, point) => sum + point.totalDue, 0);
    return totalDue ? (this.trendTotalCollected / totalDue) * 100 : 0;
  }

  get peakCollection(): MonthlyCollectionPoint | undefined {
    return this.collectionTrend.reduce<MonthlyCollectionPoint | undefined>(
      (peak, point) => (point.totalPaid > (peak?.totalPaid ?? -1) ? point : peak),
      undefined
    );
  }

  get lowestCollection(): MonthlyCollectionPoint | undefined {
    const monthsWithCollections = this.collectionTrend.filter((point) => point.totalPaid > 0);
    return monthsWithCollections.reduce<MonthlyCollectionPoint | undefined>(
      (lowest, point) => (point.totalPaid < (lowest?.totalPaid ?? Infinity) ? point : lowest),
      undefined
    );
  }

  get outstandingReceivables(): number {
    return this.collectionTrend.reduce(
      (sum, point) => sum + Math.max(point.totalDue - point.totalPaid, 0),
      0
    );
  }

  get monthOverMonthChange(): number | undefined {
    const now = new Date();
    const latestKey = `${now.getFullYear()}-${now.getMonth() + 1}`;
    const previousDate = new Date(now.getFullYear(), now.getMonth() - 1, 1);
    const previousKey = `${previousDate.getFullYear()}-${previousDate.getMonth() + 1}`;
    const latest = this.collectionTrend.find((point) => point.key === latestKey);
    const previous = this.collectionTrend.find((point) => point.key === previousKey);
    if (!latest || !previous || previous.totalPaid === 0) return undefined;
    return ((latest.totalPaid - previous.totalPaid) / previous.totalPaid) * 100;
  }

  get chartMaximum(): number {
    return Math.max(
      ...this.collectionTrend.flatMap((point) => [point.totalDue, point.totalPaid]),
      0
    );
  }

  chartX(index: number): number {
    return 74 + index * (660 / Math.max(this.collectionTrend.length, 1));
  }

  chartY(amount: number): number {
    return 205 - (this.chartMaximum ? (amount / this.chartMaximum) * 150 : 0);
  }

  chartHeight(amount: number): number {
    return this.chartMaximum ? (amount / this.chartMaximum) * 150 : 0;
  }

  formatChartAmount(amount: number): string {
    return new Intl.NumberFormat('en-BD', {
      notation: 'compact',
      maximumFractionDigits: 1,
    }).format(amount);
  }

  loadPropertyOccupancy(): void {
    this.getActiveProperties().subscribe({
      next: (properties) => {
        if (!properties.length) {
          this.propertyData = [];
          this.propertySummaries = [];
          this.stats.propertyCount = 0;
          this.stats.totalUnits = 0;
          this.stats.occupiedUnits = 0;
          this.stats.occupancyRate = 0;
          return;
        }

        forkJoin(
          properties.map((property) => this.propertyService.getPropertyWithStats(property.id))
        ).subscribe({
          next: (stats) => this.processPropertyOccupancy(stats),
          error: (error) =>
            console.error('Error loading property occupancy:', safeErrorMessage(error)),
        });
      },
      error: (error) => console.error('Error loading properties:', safeErrorMessage(error)),
    });
  }

  private getActiveProperties(page = 1, accumulated: Property[] = []): Observable<Property[]> {
    return this.propertyService.getProperties({ active: true, page, page_size: 100 }).pipe(
      switchMap((response) => {
        const properties = [...accumulated, ...response.properties];
        if (page >= response.pagination.total_pages || response.properties.length === 0) {
          return of(properties);
        }
        return this.getActiveProperties(page + 1, properties);
      })
    );
  }

  processPropertyOccupancy(properties: PropertyWithStats[]): void {
    this.propertyData = properties;
    const occupiedUnits = properties.reduce((sum, property) => sum + property.occupied_units, 0);
    const totalUnits = properties.reduce((sum, property) => sum + property.unit_count, 0);
    this.stats.occupiedUnits = occupiedUnits;
    this.stats.totalUnits = totalUnits;
    this.stats.occupancyRate = totalUnits ? (occupiedUnits / totalUnits) * 100 : 0;

    this.stats.propertyCount = properties.length;
    this.rebuildPropertySummaries();
  }

  private rebuildPropertySummaries(): void {
    const activeLeases = this.activeLeaseData.filter((lease) => lease.active);
    const totalMonthlyRent = activeLeases.reduce((total, lease) => total + lease.monthly_rent, 0);

    this.propertySummaries = this.propertyData
      .map((property) => {
        const leases = activeLeases.filter((lease) => lease.property_id === property.id);
        const monthlyRent = leases.reduce((total, lease) => total + lease.monthly_rent, 0);
        return {
          propertyId: property.id,
          propertyName: property.property_name,
          address: property.address,
          buildingCount: property.building_count,
          leaseCount: leases.length,
          monthlyRent,
          rentShare: totalMonthlyRent ? (monthlyRent / totalMonthlyRent) * 100 : 0,
          occupiedUnits: property.occupied_units,
          totalUnits: property.unit_count,
          vacantUnits: Math.max(property.unit_count - property.occupied_units, 0),
          occupancyRate: property.unit_count
            ? (property.occupied_units / property.unit_count) * 100
            : 0,
        };
      })
      .sort((a, b) => b.monthlyRent - a.monthlyRent)
      .slice(0, 3);
  }

  loadRecentCollections(): void {
    this.paymentService.getPayments({ page: 1, page_size: 100 }).subscribe({
      next: ({ payments }) => {
        const recentPayments = payments
          .filter((payment) => payment.amount_paid > 0)
          .sort(
            (a, b) =>
              new Date(b.payment_date ?? b.updated_at).getTime() -
              new Date(a.payment_date ?? a.updated_at).getTime()
          )
          .slice(0, 5);

        if (!recentPayments.length) {
          this.recentCollections = [];
          return;
        }

        forkJoin(
          recentPayments.map((payment) =>
            this.paymentService.getPaymentTransactions(payment.id).pipe(
              map((transactions) =>
                transactions.map((transaction) => ({
                  ...transaction,
                  tenantName: payment.tenant_name,
                  propertyName: payment.property_name,
                  unitNumber: payment.unit_number,
                  amountDue: payment.amount_due,
                  amountPaid: payment.amount_paid,
                }))
              )
            )
          )
        ).subscribe({
          next: (transactionGroups) => {
            this.recentCollections = transactionGroups
              .flat()
              .sort(
                (a, b) => new Date(b.payment_date).getTime() - new Date(a.payment_date).getTime()
              )
              .slice(0, 5);
          },
          error: (error) =>
            console.error('Error loading recent transactions:', safeErrorMessage(error)),
        });
      },
      error: (error) => console.error('Error loading recent payments:', safeErrorMessage(error)),
    });
  }

  processAttachmentData(attachments: Attachment[]) {
    this.stats.totalAttachments = attachments.length;
  }

  downloadAttachment(attachment: Attachment): void {
    this.attachmentService.downloadAttachment(attachment.id).subscribe({
      next: (blob) => {
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = attachment.file_name;
        link.click();
        window.URL.revokeObjectURL(url);
      },
      error: (error) => console.error('Error downloading attachment:', safeErrorMessage(error)),
    });
  }

  formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-BD', {
      style: 'currency',
      currency: 'BDT',
      minimumFractionDigits: 0,
    }).format(amount);
  }

  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-BD');
  }

  // Reactive methods using NgRx selectors
  getUserRole$(): Observable<string> {
    return this.userRole$;
  }

  // Backward compatibility methods (synchronous)
  getUserRole(): string {
    return this.authService.getUserRole();
  }

  canViewLeases(): boolean {
    return this.authService.isAdmin() || this.authService.isPropertyManager();
  }

  canViewAttachments(): boolean {
    return this.authService.isAdmin() || this.authService.isPropertyManager();
  }

  getFileIcon(contentType: string): string {
    if (contentType.startsWith('image/')) return 'image';
    if (contentType.includes('pdf')) return 'picture_as_pdf';
    if (contentType.includes('word') || contentType.includes('document')) return 'description';
    if (contentType.includes('excel') || contentType.includes('spreadsheet')) return 'table_chart';
    return 'insert_drive_file';
  }

  getTypeColor(type: string): string {
    switch (type) {
      case 'CONTRACT':
        return 'primary';
      case 'RECEIPT':
        return 'accent';
      case 'INVOICE':
        return 'warn';
      case 'PHOTO':
        return 'primary';
      case 'DOCUMENT':
        return 'accent';
      default:
        return 'basic';
    }
  }
}
