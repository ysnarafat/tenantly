import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatCardModule } from '@angular/material/card';
import { MatSelectModule } from '@angular/material/select';
import { FormsModule } from '@angular/forms';
import { UserInvitationService } from '../../../core/services/user-invitation.service';
import { OrganizationService } from '../../../core/services/organization.service';
import { UserInvitation, Organization } from '../../../core/models';

@Component({
  selector: 'app-pending-invitations',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    MatTooltipModule,
    MatCardModule,
    MatSelectModule,
  ],
  template: `
    <div class="invitations-container">
      <div class="header">
        <h2>Pending Invitations</h2>
        <button mat-raised-button color="primary" (click)="inviteUser()">
          <mat-icon>person_add</mat-icon>
          Invite User
        </button>
      </div>

      <div class="filters">
        <mat-select
          placeholder="Filter by organization"
          [(ngModel)]="selectedOrgId"
          (ngModelChange)="onOrgChange($event)"
        >
          <mat-option [value]="null">All Organizations</mat-option>
          @for (org of organizations(); track org.id) {
            <mat-option [value]="org.id">{{ org.name }}</mat-option>
          }
        </mat-select>
      </div>

      @if (loading()) {
        <div class="spinner-wrap">
          <mat-spinner diameter="40"></mat-spinner>
        </div>
      } @else if (invitations().length === 0) {
        <div class="empty-state">
          <mat-icon>mail_outline</mat-icon>
          <p>No pending invitations</p>
          <button mat-button color="primary" (click)="inviteUser()">Send First Invitation</button>
        </div>
      } @else {
        <table mat-table [dataSource]="invitations()">
          <ng-container matColumnDef="email">
            <th mat-header-cell *matHeaderCellDef>Email</th>
            <td mat-cell *matCellDef="let inv">{{ inv.email }}</td>
          </ng-container>

          <ng-container matColumnDef="role">
            <th mat-header-cell *matHeaderCellDef>Role</th>
            <td mat-cell *matCellDef="let inv">
              <mat-chip>{{ inv.role }}</mat-chip>
            </td>
          </ng-container>

          <ng-container matColumnDef="expires">
            <th mat-header-cell *matHeaderCellDef>Expires</th>
            <td mat-cell *matCellDef="let inv">{{ inv.expires_at | date: 'mediumDate' }}</td>
          </ng-container>

          <ng-container matColumnDef="created">
            <th mat-header-cell *matHeaderCellDef>Sent</th>
            <td mat-cell *matCellDef="let inv">{{ inv.created_at | date: 'mediumDate' }}</td>
          </ng-container>

          <ng-container matColumnDef="actions">
            <th mat-header-cell *matHeaderCellDef>Actions</th>
            <td mat-cell *matCellDef="let inv">
              <button
                mat-icon-button
                color="warn"
                matTooltip="Revoke invitation"
                (click)="revokeInvitation(inv)"
              >
                <mat-icon>cancel</mat-icon>
              </button>
            </td>
          </ng-container>

          <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
          <tr mat-row *matRowDef="let row; columns: displayedColumns"></tr>
        </table>
      }
    </div>
  `,
  styles: [
    `
      .invitations-container {
        padding: 24px;
        max-width: 1000px;
        margin: 0 auto;
      }
      .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
      }
      h2 {
        margin: 0;
      }
      .filters {
        margin-bottom: 16px;
      }
      mat-select {
        min-width: 220px;
      }
      .spinner-wrap {
        display: flex;
        justify-content: center;
        padding: 40px;
      }
      .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 60px 0;
        color: #666;
        gap: 8px;
      }
      .empty-state mat-icon {
        font-size: 48px;
        width: 48px;
        height: 48px;
      }
      table {
        width: 100%;
      }
    `,
  ],
})
export class PendingInvitationsComponent implements OnInit {
  private invitationService = inject(UserInvitationService);
  private organizationService = inject(OrganizationService);
  private snackBar = inject(MatSnackBar);
  private router = inject(Router);

  loading = signal(false);
  invitations = signal<UserInvitation[]>([]);
  organizations = signal<Organization[]>([]);
  selectedOrgId: number | null = null;

  displayedColumns = ['email', 'role', 'expires', 'created', 'actions'];

  ngOnInit(): void {
    this.organizationService.getOrganizations().subscribe({
      next: (res) => {
        this.organizations.set(res.organizations);
        if (res.organizations.length > 0) {
          this.selectedOrgId = res.organizations[0].id;
          this.loadInvitations(this.selectedOrgId);
        }
      },
      error: () => this.snackBar.open('Failed to load organizations', 'Close', { duration: 3000 }),
    });
  }

  onOrgChange(orgId: number | null): void {
    if (orgId) {
      this.loadInvitations(orgId);
    } else {
      this.invitations.set([]);
    }
  }

  loadInvitations(orgId: number): void {
    this.loading.set(true);
    this.invitationService.getPendingInvitations(orgId).subscribe({
      next: (res) => {
        this.invitations.set(res.invitations);
        this.loading.set(false);
      },
      error: () => {
        this.snackBar.open('Failed to load invitations', 'Close', { duration: 3000 });
        this.loading.set(false);
      },
    });
  }

  revokeInvitation(inv: UserInvitation): void {
    if (!confirm(`Revoke invitation for ${inv.email}?`)) return;
    if (!this.selectedOrgId) return;

    this.invitationService.revokeInvitation(inv.id, this.selectedOrgId).subscribe({
      next: () => {
        this.snackBar.open('Invitation revoked', 'Close', { duration: 3000 });
        if (this.selectedOrgId) this.loadInvitations(this.selectedOrgId);
      },
      error: (err) => {
        this.snackBar.open(err.error?.error || 'Failed to revoke invitation', 'Close', {
          duration: 5000,
        });
      },
    });
  }

  inviteUser(): void {
    this.router.navigate(['/admin/invitations/new']);
  }
}
