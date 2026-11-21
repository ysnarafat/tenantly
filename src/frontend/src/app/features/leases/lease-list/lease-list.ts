import { Component, OnInit, inject, signal, computed } from '@angular/core';
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
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatPaginatorModule, PageEvent, MatPaginator } from '@angular/material/paginator';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSelectModule } from '@angular/material/select';
import { MatOptionModule } from '@angular/material/core';
import { TranslateModule } from '@ngx-translate/core';
import { BreakpointObserver, Breakpoints, BreakpointState } from '@angular/cdk/layout';
import {
  LeaseService,
  LeaseWithDetails,
  LeaseListResponse,
} from '../../../core/services/lease.service';
import { EditLeaseDialogComponent } from '../edit-lease-dialog/edit-lease-dialog';
import { CreateLeaseDialogComponent } from '../create-lease-dialog/create-lease-dialog';
import { RenewLeaseDialogComponent } from '../renew-lease-dialog/renew-lease-dialog';
import { ConfirmDeleteDialogComponent } from '../../../shared/components/confirm-delete-dialog/confirm-delete-dialog';
import { LoadingSpinner } from '../../../shared/components/loading-spinner/loading-spinner';
import { Observable, of } from 'rxjs';
import { map, catchError, finalize } from 'rxjs/operators';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

interface LeaseWithStatus extends LeaseWithDetails {
  status: 'active' | 'expiring' | 'expired' | 'inactive';
}

@Component({
  selector: 'app-lease-list',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TranslateModule,
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
    MatDialogModule,
    LoadingSpinner,
  ],
  templateUrl: './lease-list.html',
  styleUrl: './lease-list.component.scss',
})
export class LeaseListComponent implements OnInit {
  readonly today = new Date();

  // State signals
  searchQuery = signal('');
  statusFilter = signal('all');
  typeFilter = signal('all');
  expandedLease = signal<LeaseWithStatus | null>(null);
  leases = signal<LeaseWithStatus[]>([]);
  loading = signal(false);
  pageIndex = signal(0);
  pageSize = signal(20);
  private _isMobile = signal(false);

  // Computed values
  readonly displayColumns = [
    'unit_info',
    'tenant_name',
    'lease_type',
    'monthly_rent',
    'duration',
    'status',
    'actions',
  ];

  filteredLeases = computed(() => {
    const searchQuery = this.searchQuery().toLowerCase();
    const statusFilter = this.statusFilter();
    const typeFilter = this.typeFilter();

    return this.leases().filter((lease) => {
      // Status filter
      if (statusFilter !== 'all' && lease.status !== statusFilter) {
        return false;
      }

      // Type filter
      if (typeFilter !== 'all' && lease.lease_type?.toLowerCase() !== typeFilter.toLowerCase()) {
        return false;
      }

      // Search query filter
      if (searchQuery) {
        const searchableText =
          `${lease.id} ${lease.property_name} ${lease.unit_number} ${lease.tenant_name} ${lease.building_name}`.toLowerCase();
        if (!searchableText.includes(searchQuery)) {
          return false;
        }
      }

      return true;
    });
  });

  pagedLeases = computed(() => {
    const start = this.pageIndex() * this.pageSize();
    return this.filteredLeases().slice(start, start + this.pageSize());
  });

  activeCount = computed(() => this.filteredLeases().filter((l) => l.status === 'active').length);
  expiringCount = computed(
    () => this.filteredLeases().filter((l) => l.status === 'expiring').length
  );
  expiredCount = computed(() => this.filteredLeases().filter((l) => l.status === 'expired').length);
  inactiveCount = computed(
    () => this.filteredLeases().filter((l) => l.status === 'inactive').length
  );

  hasActiveFilters = computed(
    () =>
      this.statusFilter() !== 'all' || this.typeFilter() !== 'all' || this.searchQuery().length > 0
  );

  readonly snackBar = inject(MatSnackBar);
  readonly dialog = inject(MatDialog);
  readonly breakpointObserver = inject(BreakpointObserver);
  readonly leaseService = inject(LeaseService);

  ngOnInit(): void {
    this.breakpointObserver
      .observe([Breakpoints.Handset, Breakpoints.Tablet])
      .subscribe((result: BreakpointState) => {
        this._isMobile.set(result.matches);
      });

    this.loadLeases();
  }

  private computeStatus(lease: LeaseWithDetails): 'active' | 'expiring' | 'expired' | 'inactive' {
    if (!lease.active) return 'inactive';
    if (lease.is_expired) return 'expired';
    if (lease.days_remaining <= 30) return 'expiring';
    return 'active';
  }

  loadLeases(): void {
    this.loading.set(true);
    this.leaseService
      .getAllLeases(1, 1000)
      .pipe(
        map((response: LeaseListResponse) =>
          (response.leases || []).map((lease) => ({
            ...lease,
            status: this.computeStatus(lease),
          }))
        ),
        catchError((error) => {
          console.error('Error loading leases:', error);
          notifyError(this.snackBar, 'Failed to load leases. Please try again.');
          return of([]);
        }),
        finalize(() => this.loading.set(false))
      )
      .subscribe({
        next: (leases) => {
          this.leases.set(leases || []);
        },
      });
  }

  isMobileView(): boolean {
    return this._isMobile();
  }

  viewLease(id: number, mode: 'view' | 'edit' = 'edit'): void {
    const leaseId = id;
    if (leaseId > 0) {
      this.leaseService.getLeaseById(leaseId).subscribe({
        next: (lease) => {
          const leaseWithStatus: LeaseWithStatus = {
            ...lease,
            status: this.computeStatus(lease),
          };
          this.expandedLease.set(leaseWithStatus);

          const dialogRef = this.dialog.open(EditLeaseDialogComponent, {
            data: leaseWithStatus,
          });

          dialogRef.afterClosed().subscribe((result) => {
            if (result) {
              const { id, ...updateRequest } = result;
              this.leaseService.updateLease(id, updateRequest).subscribe({
                next: () => {
                  notifySuccess(this.snackBar, 'Lease updated successfully');
                  this.loadLeases();
                },
                error: (error) => {
                  console.error('Error updating lease:', error);
                  notifyError(this.snackBar, 'Failed to update lease. Please try again.');
                },
              });
            }
          });
        },
        error: () => {
          console.error('Error fetching lease details');
          notifyError(this.snackBar, 'Could not load lease details');
        },
      });
    } else {
      console.error('Invalid lease ID');
      notifyError(this.snackBar, 'Invalid lease ID');
    }
  }

  deleteLease(lease: LeaseWithStatus): void {
    const dialogRef = this.dialog.open(ConfirmDeleteDialogComponent, {
      data: {
        message: `Are you sure you want to delete lease ${lease.id} (Unit ${lease.unit_number})?`,
      },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result === 'delete') {
        this.leaseService.deleteLease(lease.id).subscribe({
          next: () => {
            notifySuccess(this.snackBar, 'Lease deleted successfully');
            this.loadLeases();
          },
          error: (error) => {
            console.error('Error deleting lease:', error);
            notifyError(this.snackBar, 'Failed to delete lease. Please try again.');
          },
        });
      }
    });
  }

  toggleExpand(lease: LeaseWithStatus, $event?: Event): void {
    if ($event) {
      $event.stopPropagation();
    }
    this.expandedLease.set(this.expandedLease() === lease ? null : lease);
  }

  readonly isMainRow = (row: LeaseWithStatus): boolean => this.expandedLease() !== row;

  readonly isDetailRow = (row: LeaseWithStatus): boolean => this.expandedLease() === row;

  formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
  }

  formatDate(date: string | Date): string {
    const d = new Date(date);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  getStatusClass(status: string): string {
    const classes: Record<string, string> = {
      active: 'green',
      expiring: 'orange',
      expired: 'red',
      inactive: 'grey',
    };
    return classes[status] || 'grey';
  }

  getStatusText(status: string): string {
    return status?.toUpperCase() || 'UNKNOWN';
  }

  canEdit(lease?: LeaseWithStatus): boolean {
    return true;
  }

  canDelete(lease?: LeaseWithStatus): boolean {
    return true;
  }

  createLease(): void {
    const dialogRef = this.dialog.open(CreateLeaseDialogComponent, {
      data: { existingLease: null },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.leaseService.createLease(result).subscribe({
          next: () => {
            notifySuccess(this.snackBar, 'Lease created successfully');
            this.loadLeases();
          },
          error: (error) => {
            console.error('Error creating lease:', error);
            notifyError(this.snackBar, 'Failed to create lease. Please try again.');
          },
        });
      }
    });
  }

  editLease(lease: LeaseWithStatus): void {
    const dialogRef = this.dialog.open(EditLeaseDialogComponent, {
      data: lease,
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        const { id, ...updateRequest } = result;
        this.leaseService.updateLease(id, updateRequest).subscribe({
          next: () => {
            notifySuccess(this.snackBar, 'Lease updated successfully');
            this.loadLeases();
          },
          error: (error) => {
            console.error('Error updating lease:', error);
            notifyError(this.snackBar, 'Failed to update lease. Please try again.');
          },
        });
      }
    });
  }

  renewLease(lease: LeaseWithStatus): void {
    const dialogRef = this.dialog.open(RenewLeaseDialogComponent, {
      data: lease,
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.leaseService.renewLease(lease.id, result).subscribe({
          next: () => {
            notifySuccess(this.snackBar, 'Lease renewed successfully');
            this.loadLeases();
          },
          error: (error) => {
            console.error('Error renewing lease:', error);
            notifyError(this.snackBar, 'Failed to renew lease. Please try again.');
          },
        });
      }
    });
  }

  onSortChange($event: Sort): void {
    // Sorting is handled by the table data source
  }

  onPageChange($event: PageEvent): void {
    this.pageIndex.set($event.pageIndex);
    this.pageSize.set($event.pageSize);
  }

  setStatusFilter(value: string): void {
    this.statusFilter.set(value);
    this.pageIndex.set(0);
  }

  setTypeFilter(value: string): void {
    this.typeFilter.set(value);
    this.pageIndex.set(0);
  }

  clearSearch(): void {
    this.searchQuery.set('');
    this.pageIndex.set(0);
  }

  resetFilters(): void {
    this.statusFilter.set('all');
    this.typeFilter.set('all');
    this.searchQuery.set('');
    this.pageIndex.set(0);
  }

  // Template calls this on search/model change
  applyFilters(): void {
    this.pageIndex.set(0);
  }
}

export { LeaseListComponent as LeaseList };
