import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatGridListModule } from '@angular/material/grid-list';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { TranslateModule } from '@ngx-translate/core';
import { Store } from '@ngrx/store';
import { Observable } from 'rxjs';
import { LeaseService } from '../../core/services/lease.service';
import { LeaseWithDetails } from '../../core/models/lease.model';
import { AttachmentService, Attachment } from '../../core/services/attachment.service';
import { AuthService } from '../../core/services/auth.service';
import { AppState } from '../../store';
import * as AuthSelectors from '../../store/auth/auth.selectors';

interface DashboardStats {
  totalRentDue: number;
  collectedThisMonth: number;
  pendingPayments: number;
  overdueShops: number;
  activeLeases: number;
  expiringLeases: number;
  totalAttachments: number;
  recentAttachments: number;
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    MatCardModule,
    MatGridListModule,
    MatIconModule,
    MatButtonModule,
    MatChipsModule,
    TranslateModule,
  ],
  templateUrl: './dashboard.html',
  styleUrls: ['./dashboard.scss'],
})
export class Dashboard implements OnInit {
  private leaseService = inject(LeaseService);
  private attachmentService = inject(AttachmentService);
  private authService = inject(AuthService);
  private store = inject(Store<AppState>);

  // NgRx selectors for reactive data
  user$ = this.store.select(AuthSelectors.selectUser);
  userRole$ = this.store.select(AuthSelectors.selectUserRole);
  isAdmin$ = this.store.select(AuthSelectors.selectIsAdmin);
  isPropertyManager$ = this.store.select(AuthSelectors.selectIsPropertyManager);
  canViewLeases$ = this.store.select(AuthSelectors.selectCanViewLeases);
  canViewAttachments$ = this.store.select(AuthSelectors.selectCanViewAttachments);

  stats: DashboardStats = {
    totalRentDue: 0,
    collectedThisMonth: 0,
    pendingPayments: 0,
    overdueShops: 0,
    activeLeases: 0,
    expiringLeases: 0,
    totalAttachments: 0,
    recentAttachments: 0,
  };

  recentLeases: LeaseWithDetails[] = [];
  expiringLeases: LeaseWithDetails[] = [];
  recentAttachments: Attachment[] = [];
  loading = true;

  ngOnInit() {
    this.loadDashboardData();
  }

  loadDashboardData() {
    this.loading = true;

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
      error: (error) => console.error('Error loading leases:', error),
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
        console.error('Error loading attachments:', error);
        this.loading = false;
      },
    });
  }

  processLeaseData(leases: LeaseWithDetails[]) {
    this.stats.activeLeases = leases.filter((l) => l.active).length;
    this.stats.overdueShops = leases.filter((l) => l.is_expired).length;

    // Calculate total rent due (sum of all active leases' monthly rent)
    this.stats.totalRentDue = leases
      .filter((l) => l.active)
      .reduce((sum, lease) => sum + lease.monthly_rent, 0);

    // Demo data for collected and pending payments
    this.stats.collectedThisMonth = this.stats.totalRentDue * 0.7; // 70% collected
    this.stats.pendingPayments = this.stats.totalRentDue * 0.3; // 30% pending
  }

  processAttachmentData(attachments: Attachment[]) {
    this.stats.totalAttachments = attachments.length;

    // Count recent attachments (last 7 days)
    const sevenDaysAgo = new Date();
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);

    this.stats.recentAttachments = attachments.filter(
      (a) => new Date(a.created_at) >= sevenDaysAgo
    ).length;
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
