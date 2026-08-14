import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatMenuModule } from '@angular/material/menu';
import { MatSelectModule } from '@angular/material/select';
import { MatOptionModule } from '@angular/material/core';
import { TranslateModule } from '@ngx-translate/core';
import { switchMap } from 'rxjs/operators';
import { TenantService } from '../../../core/services/tenant.service';
import { MfaService } from '../../../core/services/mfa.service';
import { TenantFormDialogComponent } from '../tenant-form-dialog/tenant-form-dialog';
import { Tenant } from '../../../core/models/tenant.model';
import { cleanEmptyFields } from '../../../shared/utils/object.utils';
import { DataTable } from '../../../shared/components/data-table/data-table';
import { PermissionService } from '../../../core/services/permission.service';
import {
  maskFromLastFour,
  maskPhone,
  RESTRICTED_LABEL,
} from '../../../shared/utils/pii-mask.utils';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { ConfirmDeleteDialogComponent } from '../../../shared/components/confirm-delete-dialog/confirm-delete-dialog';

@Component({
  selector: 'app-tenant-list',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatDialogModule,
    MatTableModule,
    MatPaginatorModule,
    MatInputModule,
    MatFormFieldModule,
    MatTooltipModule,
    MatProgressSpinnerModule,
    MatMenuModule,
    MatSelectModule,
    MatOptionModule,
    TranslateModule,
    DataTable,
  ],
  templateUrl: './tenant-list.html',
  styleUrls: ['./tenant-list.scss'],
})
export class TenantList implements OnInit {
  private dialog = inject(MatDialog);
  private tenantService = inject(TenantService);
  private mfaService = inject(MfaService);
  private snackBar = inject(MatSnackBar);
  private permissionService = inject(PermissionService);

  private revealState = new Map<number, { nid: boolean; phone: boolean }>();
  get canRevealPii(): boolean {
    return (
      this.permissionService.isSuperAdmin() ||
      this.permissionService.isOrgAdmin() ||
      this.permissionService.isAdmin() ||
      this.permissionService.isPropertyManager()
    );
  }

  get isAccountantRole(): boolean {
    return this.permissionService.isAccountant();
  }

  // Cache of full NIDs fetched from the role-gated reveal endpoint, keyed by
  // tenant id. Populated lazily on first reveal.
  private revealedNid = new Map<number, string>();

  toggleNidReveal(id: number): void {
    const current = this.revealState.get(id) ?? { nid: false, phone: false };
    const willReveal = !current.nid;
    this.revealState.set(id, { ...current, nid: willReveal });

    // Fetch the full NID on demand the first time it is revealed, gated by an
    // MFA step-up challenge.
    if (willReveal && !this.revealedNid.has(id)) {
      this.mfaService
        .ensureStepUp()
        .pipe(switchMap((token) => this.tenantService.getTenantNid(id, token)))
        .subscribe({
          next: (res) => this.revealedNid.set(id, res.nid_number),
          error: (err) => {
            this.mfaService.clearStepUp();
            console.error('Error revealing NID:', safeErrorMessage(err));
            this.snackBar.open('Failed to reveal NID', 'Close', { duration: 3000 });
            const state = this.revealState.get(id) ?? { nid: false, phone: false };
            this.revealState.set(id, { ...state, nid: false });
          },
        });
    }
  }

  togglePhoneReveal(id: number): void {
    const current = this.revealState.get(id) ?? { nid: false, phone: false };
    this.revealState.set(id, { ...current, phone: !current.phone });
  }

  isNidRevealed(id: number): boolean {
    return this.revealState.get(id)?.nid ?? false;
  }

  isPhoneRevealed(id: number): boolean {
    return this.revealState.get(id)?.phone ?? false;
  }

  getDisplayNid(tenant: Tenant): string {
    if (!tenant.nid_last_four) return '—';
    if (this.isAccountantRole) return RESTRICTED_LABEL;
    if (this.isNidRevealed(tenant.id)) {
      return this.revealedNid.get(tenant.id) ?? maskFromLastFour(tenant.nid_last_four);
    }
    return maskFromLastFour(tenant.nid_last_four);
  }

  getDisplayPhone(tenant: Tenant): string {
    if (!tenant.phone_number) return '—';
    if (this.isAccountantRole) return RESTRICTED_LABEL;
    return this.isPhoneRevealed(tenant.id) ? tenant.phone_number : maskPhone(tenant.phone_number);
  }

  tenants: Tenant[] = [];
  filteredTenants: Tenant[] = [];
  pagedTenants: Tenant[] = [];
  loading = false;

  searchQuery = '';
  activeFilter: 'all' | 'active' | 'inactive' = 'all';
  typeFilter: 'all' | 'Individual' | 'Business' = 'all';

  pageSize = 10;
  pageIndex = 0;

  displayedColumns: string[] = ['name', 'type', 'contact', 'nid', 'status', 'created', 'actions'];

  ngOnInit() {
    this.loadTenants();
  }

  loadTenants() {
    this.loading = true;
    this.tenantService.getAllTenants(1, 100).subscribe({
      next: (response) => {
        this.tenants = response.tenants || [];
        this.applyFilters();
        this.loading = false;
      },
      error: (err) => {
        console.error('Error fetching tenants:', safeErrorMessage(err));
        this.snackBar.open('Failed to load tenants', 'Close', { duration: 3000 });
        this.loading = false;
      },
    });
  }

  applyFilters() {
    let result = [...this.tenants];

    if (this.searchQuery.trim()) {
      const q = this.searchQuery.toLowerCase();
      result = result.filter(
        (t) =>
          t.name.toLowerCase().includes(q) ||
          t.phone_number?.toLowerCase().includes(q) ||
          t.email?.toLowerCase().includes(q) ||
          t.nid_last_four?.toLowerCase().includes(q)
      );
    }

    if (this.activeFilter === 'active') result = result.filter((t) => t.active);
    if (this.activeFilter === 'inactive') result = result.filter((t) => !t.active);
    if (this.typeFilter !== 'all') result = result.filter((t) => t.tenant_type === this.typeFilter);

    this.filteredTenants = result;
    this.pageIndex = 0;
    this.updatePagedData();
  }

  updatePagedData() {
    const start = this.pageIndex * this.pageSize;
    this.pagedTenants = this.filteredTenants.slice(start, start + this.pageSize);
  }

  onPageChange(event: PageEvent) {
    this.pageIndex = event.pageIndex;
    this.pageSize = event.pageSize;
    this.updatePagedData();
  }

  setActiveFilter(filter: 'all' | 'active' | 'inactive') {
    this.activeFilter = filter;
    this.applyFilters();
  }

  setTypeFilter(filter: 'all' | 'Individual' | 'Business') {
    this.typeFilter = filter;
    this.applyFilters();
  }

  onSearchChange() {
    this.applyFilters();
  }

  clearSearch() {
    this.searchQuery = '';
    this.applyFilters();
  }

  get hasActiveFilters(): boolean {
    return !!(this.searchQuery || this.activeFilter !== 'all' || this.typeFilter !== 'all');
  }

  get activeCount() {
    return this.tenants.filter((t) => t.active).length;
  }
  get inactiveCount() {
    return this.tenants.filter((t) => !t.active).length;
  }

  addTenant() {
    const dialogRef = this.dialog.open(TenantFormDialogComponent, {
      width: '600px',
      data: { mode: 'create' },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.tenantService.createTenant(result).subscribe({
          next: () => {
            this.snackBar.open('Tenant created successfully', 'Close', { duration: 3000 });
            this.loadTenants();
          },
          error: (err) => {
            this.snackBar.open(err.error?.message || 'Failed to create tenant', 'Close', {
              duration: 3000,
            });
          },
        });
      }
    });
  }

  viewTenant(tenant: Tenant) {
    this.dialog.open(TenantFormDialogComponent, {
      width: '600px',
      data: { mode: 'view', tenant },
    });
  }

  editTenant(tenant: Tenant) {
    const dialogRef = this.dialog.open(TenantFormDialogComponent, {
      width: '600px',
      data: { mode: 'edit', tenant },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        const cleanedResult = cleanEmptyFields(result);
        this.tenantService.updateTenant(tenant.id, cleanedResult).subscribe({
          next: () => {
            this.snackBar.open('Tenant updated successfully', 'Close', { duration: 3000 });
            this.loadTenants();
          },
          error: (err) => {
            this.snackBar.open(err.error?.message || 'Failed to update tenant', 'Close', {
              duration: 3000,
            });
          },
        });
      }
    });
  }

  deleteTenant(tenant: Tenant) {
    const dialogRef = this.dialog.open(ConfirmDeleteDialogComponent, {
      width: '480px',
      data: { entityLabel: 'tenant', entityName: tenant.name },
    });

    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.tenantService.deleteTenant(tenant.id).subscribe({
          next: () => {
            this.snackBar.open('Tenant deleted', 'Close', { duration: 3000 });
            this.loadTenants();
          },
          error: (err) => {
            this.snackBar.open(err.error?.message || 'Failed to delete tenant', 'Close', {
              duration: 3000,
            });
          },
        });
      }
    });
  }
}
