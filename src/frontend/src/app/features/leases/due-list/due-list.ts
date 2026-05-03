import { Component, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { LeaseService, LeaseDue, DueSummary } from '../../../core/services/lease.service';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-due-list',
  standalone: true,
  imports: [
    CommonModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatCardModule,
    MatChipsModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './due-list.html',
  styleUrls: ['./due-list.scss'],
})
export class DueList {
  private leaseService = inject(LeaseService);
  private authService = inject(AuthService);
  private snackBar = inject(MatSnackBar);

  leases = signal<LeaseDue[]>([]);
  summary = signal<DueSummary | null>(null);
  loading = signal(true);
  displayedColumns = [
    'tenant_name',
    'property_name',
    'building_name',
    'unit_number',
    'monthly_rent',
    'days_overdue',
  ];

  // Computed property for empty state
  isEmpty = computed(() => this.leases().length === 0 && !this.loading());

  ngOnInit() {
    this.loadDueList();
  }

  loadDueList() {
    this.loading.set(true);

    // Load both leases and summary in parallel
    Promise.all([
      this.leaseService.getLeasesDue().toPromise(),
      this.leaseService.getDueSummary().toPromise(),
    ])
      .then(([leasesData, summaryData]) => {
        this.leases.set(leasesData || []);
        this.summary.set(summaryData || null);
        this.loading.set(false);
      })
      .catch((error) => {
        console.error('Error loading due list:', error);
        this.snackBar.open('ভাড়া তালিকা লোড করতে সমস্যা হয়েছে', 'Close', { duration: 3000 });
        this.loading.set(false);
      });
  }

  formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-BD', {
      style: 'currency',
      currency: 'BDT',
      minimumFractionDigits: 0,
    }).format(amount);
  }

  getOverdueColor(days: number): string {
    if (days <= 5) return 'overdue-low';
    if (days <= 15) return 'overdue-medium';
    return 'overdue-high';
  }

  getOverdueIcon(days: number): string {
    if (days <= 5) return 'warning';
    if (days <= 15) return 'report_problem';
    return 'error';
  }

  canViewDueList(): boolean {
    return (
      this.authService.isAdmin() ||
      this.authService.isPropertyManager() ||
      this.authService.isAccountant()
    );
  }

  refresh() {
    this.loadDueList();
  }
}
