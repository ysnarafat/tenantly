import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatTableModule } from '@angular/material/table';
import { MatSortModule, Sort } from '@angular/material/sort';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatChipsModule } from '@angular/material/chips';
import { MatMenuModule } from '@angular/material/menu';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatDialog } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSelectModule } from '@angular/material/select';
import { MatOptionModule } from '@angular/material/core';
import { TranslateModule } from '@ngx-translate/core';
import { LeaseService, LeaseWithDetails } from '../../../core/services/lease.service';
import { AuthService } from '../../../core/services/auth.service';
import { CreateLeaseDialog } from '../create-lease-dialog/create-lease-dialog';
import { EditLeaseDialog } from '../edit-lease-dialog/edit-lease-dialog';
import { RenewLeaseDialog } from '../renew-lease-dialog/renew-lease-dialog';
import { LoadingSpinner } from '../../../shared/components/loading-spinner/loading-spinner';
import { ConfirmDeleteDialogComponent } from '../../../shared/components/confirm-delete-dialog/confirm-delete-dialog';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

@Component({
  selector: 'app-lease-list',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatTableModule,
    MatSortModule,
    MatButtonModule,
    MatIconModule,
    MatCardModule,
    MatChipsModule,
    MatMenuModule,
    MatFormFieldModule,
    MatInputModule,
    MatPaginatorModule,
    MatTooltipModule,
    MatSelectModule,
    MatOptionModule,
    TranslateModule,
    LoadingSpinner,
  ],
  templateUrl: './lease-list.html',
  styleUrls: ['./lease-list.scss'],
})
export class LeaseList implements OnInit {
  private leaseService = inject(LeaseService);
  private authService = inject(AuthService);
  private snackBar = inject(MatSnackBar);
  private dialog = inject(MatDialog);

  leases: LeaseWithDetails[] = [];
  filteredLeases: LeaseWithDetails[] = [];
  pagedLeases: LeaseWithDetails[] = [];
  loading = false;
  expandedLease: LeaseWithDetails | null = null;

  searchQuery = '';
  statusFilter: 'all' | 'active' | 'expiring' | 'expired' | 'inactive' = 'all';
  typeFilter: 'all' | 'Residential' | 'Commercial' = 'all';

  pageSize = 10;
  pageIndex = 0;

  displayedColumns: string[] = [
    'unit_info',
    'tenant_name',
    'lease_type',
    'monthly_rent',
    'duration',
    'status',
    'actions',
  ];
  expandableColumns = [...this.displayedColumns, 'expandedDetail'];

  sortState: Sort = { active: '', direction: '' };

  private readonly sortAccessors: Record<string, (l: LeaseWithDetails) => string | number> = {
    tenant_name: (l) => l.tenant_name?.toLowerCase() ?? '',
    lease_type: (l) => l.lease_type ?? '',
    monthly_rent: (l) => l.monthly_rent,
    duration: (l) => l.duration_months,
    status: (l) =>
      ({ active: 0, expiring: 1, expired: 2, inactive: 3 })[this.getStatusClass(l)] ?? 99,
  };

  ngOnInit() {
    this.loadLeases();
  }

  loadLeases() {
    this.loading = true;
    this.leaseService.getAllLeases(1, 100).subscribe({
      next: (response) => {
        this.leases = response.leases;
        this.applyFilters();
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading leases:', safeErrorMessage(error));
        notifyError(this.snackBar, 'Error loading leases');
        this.loading = false;
      },
    });
  }

  applyFilters() {
    let result = [...this.leases];

    if (this.searchQuery.trim()) {
      const q = this.searchQuery.toLowerCase();
      result = result.filter(
        (l) =>
          l.tenant_name?.toLowerCase().includes(q) ||
          l.unit_number?.toLowerCase().includes(q) ||
          l.property_name?.toLowerCase().includes(q) ||
          l.building_name?.toLowerCase().includes(q) ||
          l.tenant_phone?.toLowerCase().includes(q)
      );
    }

    if (this.statusFilter === 'active')
      result = result.filter((l) => l.active && !l.is_expired && l.days_remaining > 30);
    if (this.statusFilter === 'expiring')
      result = result.filter((l) => l.active && !l.is_expired && l.days_remaining <= 30);
    if (this.statusFilter === 'expired') result = result.filter((l) => l.is_expired);
    if (this.statusFilter === 'inactive') result = result.filter((l) => !l.active);

    if (this.typeFilter !== 'all') result = result.filter((l) => l.lease_type === this.typeFilter);

    this.filteredLeases = result;
    this.applySort();
    this.pageIndex = 0;
    this.expandedLease = null;
    this.updatePagedData();
  }

  onSortChange(sort: Sort) {
    this.sortState = sort;
    this.applySort();
    this.updatePagedData();
  }

  private applySort() {
    const { active, direction } = this.sortState;
    const accessor = this.sortAccessors[active];
    if (!direction || !accessor) return;

    const dir = direction === 'asc' ? 1 : -1;
    this.filteredLeases = [...this.filteredLeases].sort((a, b) => {
      const valueA = accessor(a);
      const valueB = accessor(b);
      if (valueA < valueB) return -dir;
      if (valueA > valueB) return dir;
      return 0;
    });
  }

  updatePagedData() {
    const start = this.pageIndex * this.pageSize;
    this.pagedLeases = this.filteredLeases.slice(start, start + this.pageSize);
  }

  onPageChange(event: PageEvent) {
    this.pageIndex = event.pageIndex;
    this.pageSize = event.pageSize;
    this.expandedLease = null;
    this.updatePagedData();
  }

  setStatusFilter(filter: 'all' | 'active' | 'expiring' | 'expired' | 'inactive') {
    this.statusFilter = filter;
    this.applyFilters();
  }

  setTypeFilter(filter: 'all' | 'Residential' | 'Commercial') {
    this.typeFilter = filter;
    this.applyFilters();
  }

  clearSearch() {
    this.searchQuery = '';
    this.applyFilters();
  }

  resetFilters() {
    this.searchQuery = '';
    this.statusFilter = 'all';
    this.typeFilter = 'all';
    this.applyFilters();
  }

  toggleExpand(lease: LeaseWithDetails, event: Event) {
    event.stopPropagation();
    this.expandedLease = this.expandedLease === lease ? null : lease;
  }

  getStatusClass(lease: LeaseWithDetails): string {
    if (!lease.active) return 'inactive';
    if (lease.is_expired) return 'expired';
    if (lease.days_remaining <= 30) return 'expiring';
    return 'active';
  }

  getStatusText(lease: LeaseWithDetails): string {
    if (!lease.active) {
      // end_reason distinguishes an early move-out or a renewal from a lease
      // that simply reached the end of its term — same badge color/sort
      // order (still "inactive"), more specific label.
      if (lease.end_reason === 'Terminated') return 'Terminated';
      if (lease.end_reason === 'Renewed') return 'Renewed';
      return 'Inactive';
    }
    if (lease.is_expired) return 'Expired';
    if (lease.days_remaining <= 30) return `${lease.days_remaining}d left`;
    return 'Active';
  }

  formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-BD', {
      style: 'currency',
      currency: 'BDT',
      minimumFractionDigits: 0,
    }).format(amount);
  }

  formatDate(dateString: string): string {
    if (!dateString) return '—';
    return new Date(dateString).toLocaleDateString('en-BD');
  }

  canEdit(): boolean {
    return this.authService.isAdmin() || this.authService.isPropertyManager();
  }

  canDelete(): boolean {
    return this.authService.isAdmin();
  }

  editLease(lease: LeaseWithDetails): void {
    const dialogRef = this.dialog.open(EditLeaseDialog, {
      width: '520px',
      data: { lease },
    });
    dialogRef.afterClosed().subscribe((result) => {
      if (result) this.loadLeases();
    });
  }

  renewLease(lease: LeaseWithDetails): void {
    const dialogRef = this.dialog.open(RenewLeaseDialog, {
      width: '520px',
      data: { lease },
    });
    dialogRef.afterClosed().subscribe((result) => {
      if (result) this.loadLeases();
    });
  }

  deleteLease(lease: LeaseWithDetails) {
    const dialogRef = this.dialog.open(ConfirmDeleteDialogComponent, {
      width: '480px',
      data: { entityLabel: 'lease', entityName: lease.tenant_name },
    });

    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.leaseService.deleteLease(lease.id).subscribe({
          next: () => {
            notifySuccess(this.snackBar, 'Lease deleted');
            this.loadLeases();
          },
          error: (error) => {
            console.error('Error deleting lease:', safeErrorMessage(error));
            notifyError(this.snackBar, 'Error deleting lease');
          },
        });
      }
    });
  }

  createLease() {
    const dialogRef = this.dialog.open(CreateLeaseDialog, {
      width: '600px',
      maxHeight: '90vh',
    });
    dialogRef.afterClosed().subscribe((result) => {
      if (result) this.loadLeases();
    });
  }

  get hasActiveFilters(): boolean {
    return !!(this.searchQuery || this.statusFilter !== 'all' || this.typeFilter !== 'all');
  }

  get activeCount() {
    return this.leases.filter((l) => l.active && !l.is_expired && l.days_remaining > 30).length;
  }
  get expiringCount() {
    return this.leases.filter((l) => l.active && !l.is_expired && l.days_remaining <= 30).length;
  }
  get expiredCount() {
    return this.leases.filter((l) => l.is_expired).length;
  }
  get inactiveCount() {
    return this.leases.filter((l) => !l.active).length;
  }

  // Predicate functions for mat-table row expansion
  isMainRow(row: LeaseWithDetails): boolean {
    return this.expandedLease !== row;
  }

  isDetailRow(row: LeaseWithDetails): boolean {
    return this.expandedLease === row;
  }
}
