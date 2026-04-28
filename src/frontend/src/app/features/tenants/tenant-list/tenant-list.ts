import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { TenantService } from '../../../core/services/tenant.service';
import { TenantFormDialogComponent } from '../tenant-form-dialog/tenant-form-dialog';
import { Tenant, PaginationInfo } from '../../../core/models/tenant.model';

@Component({
  selector: 'app-tenant-list',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatDialogModule,
    MatTableModule,
    MatPaginatorModule,
  ],
  templateUrl: './tenant-list.html',
  styleUrls: ['./tenant-list.scss'],
})
export class TenantList implements OnInit {
  private dialog = inject(MatDialog);
  private tenantService = inject(TenantService);

  tenants: Tenant[] = [];
  displayedColumns: string[] = ['name', 'type', 'email', 'phone', 'status', 'created'];
  pagination: PaginationInfo = {
    current_page: 1,
    page_size: 10,
    total_items: 0,
    total_pages: 0,
    has_next: false,
    has_prev: false,
  };

  ngOnInit() {
    this.loadTenants();
  }

  loadTenants(page = 1, pageSize = 10) {
    this.tenantService.getAllTenants(page, pageSize).subscribe({
      next: (response) => {
        this.tenants = response.tenants;
        this.pagination = response.pagination;
      },
      error: (err) => {
        console.error('Error fetching tenants:', err);
      },
    });
  }

  onPageChange(event: PageEvent) {
    this.loadTenants(event.pageIndex + 1, event.pageSize);
  }

  addTenant() {
    const dialogRef = this.dialog.open(TenantFormDialogComponent, {
      width: '600px',
      data: { mode: 'create' },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.tenantService.createTenant(result).subscribe({
          next: (tenant) => {
            console.log('Tenant created successfully:', tenant);
            this.loadTenants(); // Refresh list
          },
          error: (err) => {
            console.error('Error creating tenant:', err);
          },
        });
      }
    });
  }
}
