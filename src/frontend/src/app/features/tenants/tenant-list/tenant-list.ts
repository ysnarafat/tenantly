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
import { TranslateModule } from '@ngx-translate/core';
import { TenantService } from '../../../core/services/tenant.service';
import { TenantFormDialogComponent } from '../tenant-form-dialog/tenant-form-dialog';
import { Tenant } from '../../../core/models/tenant.model';
import { cleanEmptyFields } from '../../../shared/utils/object.utils';
import { DataTable } from '../../../shared/components/data-table/data-table';

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
    TranslateModule,
    DataTable,
  ],
  templateUrl: './tenant-list.html',
  styleUrls: ['./tenant-list.scss'],
})
export class TenantList implements OnInit {
  private dialog = inject(MatDialog);
  private tenantService = inject(TenantService);
  private snackBar = inject(MatSnackBar);

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
        console.error('Error fetching tenants:', err);
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
          t.nid_number?.toLowerCase().includes(q)
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
        console.log('Updating tenant with ID:', tenant.id, 'Data:', cleanedResult);
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
    if (confirm(`Delete tenant "${tenant.name}"? This cannot be undone.`)) {
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
  }
}
