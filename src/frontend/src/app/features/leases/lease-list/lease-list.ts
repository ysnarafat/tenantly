import { Component, OnInit, inject } from '@angular/core';

import { RouterModule } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatChipsModule } from '@angular/material/chips';
import { MatMenuModule } from '@angular/material/menu';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatDialog } from '@angular/material/dialog';
import { LeaseService } from '../../../core/services/lease.service';
import { LeaseWithDetails } from '../../../core/models/lease.model';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-lease-list',
  standalone: true,
  imports: [
    RouterModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatCardModule,
    MatChipsModule,
    MatMenuModule,
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
  displayedColumns: string[] = [
    'unit_number',
    'building_name',
    'tenant_name',
    'monthly_rent',
    'start_date',
    'end_date',
    'status',
    'actions',
  ];
  loading = false;

  ngOnInit() {
    this.loadLeases();
  }

  loadLeases() {
    this.loading = true;
    this.leaseService.getActiveLeases().subscribe({
      next: (leases) => {
        this.leases = leases;
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading leases:', error);
        this.snackBar.open('Error loading leases', 'Close', { duration: 3000 });
        this.loading = false;
      },
    });
  }

  getStatusColor(lease: LeaseWithDetails): string {
    if (!lease.active) return 'warn';
    if (lease.is_expired) return 'warn';
    if (lease.days_remaining <= 30) return 'accent';
    return 'primary';
  }

  getStatusText(lease: LeaseWithDetails): string {
    if (!lease.active) return 'Inactive';
    if (lease.is_expired) return 'Expired';
    if (lease.days_remaining <= 30) return `Expires in ${lease.days_remaining} days`;
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
    return new Date(dateString).toLocaleDateString('en-BD');
  }

  canEdit(): boolean {
    return this.authService.isAdmin() || this.authService.isPropertyManager();
  }

  canDelete(): boolean {
    return this.authService.isAdmin();
  }

  editLease(lease: LeaseWithDetails): void {
    // TODO: Implement edit lease dialog
    void lease; // Suppress unused variable warning
    this.snackBar.open('Edit lease functionality will be implemented', 'Close', {
      duration: 3000,
    });
  }

  deleteLease(lease: LeaseWithDetails) {
    this.snackBar.open(`Delete lease for ${lease.tenant_name} — not yet implemented`, 'Close', {
      duration: 3000,
    });
  }

  viewDetails(lease: LeaseWithDetails): void {
    void lease; // Suppress unused variable warning
    // TODO: Navigate to lease details page
    this.snackBar.open('Lease details page will be implemented', 'Close', {
      duration: 3000,
    });
  }

  createLease() {
    // TODO: Implement create lease dialog
    this.snackBar.open('Create lease functionality will be implemented', 'Close', {
      duration: 3000,
    });
  }
}
