import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatCardModule } from '@angular/material/card';
import { MatSelectModule } from '@angular/material/select';
import { FormsModule } from '@angular/forms';
import { Store } from '@ngrx/store';
import { toSignal } from '@angular/core/rxjs-interop';
import { UserInvitationService } from '../../../core/services/user-invitation.service';
import { OrganizationService } from '../../../core/services/organization.service';
import { UserInvitation, Organization } from '../../../core/models';
import { AppState } from '../../../store';
import * as AuthSelectors from '../../../store/auth/auth.selectors';
import { actWithUndo } from '../../../shared/utils/undo-toast.utils';
import { notifyError } from '../../../shared/utils/notify.utils';
import { DataTable } from '../../../shared/components/data-table/data-table';

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
    MatSnackBarModule,
    MatTooltipModule,
    MatCardModule,
    MatSelectModule,
    DataTable,
  ],
  template: `
    <div class="invitations-container">
      <div class="header">
        <div class="header-titles">
          <h1>Pending Invitations</h1>
          @if (!isSuperAdmin() && callerOrgName()) {
            <p class="header-subtitle">
              for <strong>{{ callerOrgName() }}</strong>
            </p>
          }
        </div>
        <button mat-raised-button color="primary" (click)="inviteUser()">
          <mat-icon>person_add</mat-icon>
          Invite User
        </button>
      </div>

      @if (isSuperAdmin()) {
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
      }

      <app-data-table
        [dataSource]="invitations()"
        [displayedColumns]="displayedColumns"
        [loading]="loading()"
        [showPaginator]="false"
        emptyIcon="mail_outline"
        emptyMessage="No pending invitations"
      >
        <ng-container matColumnDef="email">
          <th mat-header-cell *matHeaderCellDef>Email</th>
          <td mat-cell *matCellDef="let inv" data-label="Email">{{ inv.email }}</td>
        </ng-container>

        <ng-container matColumnDef="role">
          <th mat-header-cell *matHeaderCellDef>Role</th>
          <td mat-cell *matCellDef="let inv" data-label="Role">
            <mat-chip>{{ inv.role }}</mat-chip>
          </td>
        </ng-container>

        <ng-container matColumnDef="expires">
          <th mat-header-cell *matHeaderCellDef>Expires</th>
          <td mat-cell *matCellDef="let inv" data-label="Expires">
            {{ inv.expires_at | date: 'mediumDate' }}
          </td>
        </ng-container>

        <ng-container matColumnDef="created">
          <th mat-header-cell *matHeaderCellDef>Sent</th>
          <td mat-cell *matCellDef="let inv" data-label="Sent">
            {{ inv.created_at | date: 'mediumDate' }}
          </td>
        </ng-container>

        <ng-container matColumnDef="actions">
          <th mat-header-cell *matHeaderCellDef>Actions</th>
          <td mat-cell *matCellDef="let inv" data-label="">
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

        <button dtEmptyAction mat-button color="primary" (click)="inviteUser()">
          Send First Invitation
        </button>
      </app-data-table>
    </div>
  `,
  styles: [
    `
      .invitations-container {
        padding: 24px;
        max-width: 1400px;
        margin: 0 auto;
      }
      @media (max-width: 768px) {
        .invitations-container {
          padding: 16px;
        }
      }
      @media (max-width: 480px) {
        .invitations-container {
          padding: 12px;
        }
      }
      .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 24px;
        flex-wrap: wrap;
        gap: 8px;
      }
      .table-scroll {
        overflow-x: auto;
        -webkit-overflow-scrolling: touch;
      }
      h1 {
        margin: 0;
        font-size: 28px;
      }
      .header-subtitle {
        margin: 4px 0 0;
        font-size: 14px;
        color: var(--text-secondary, #666);
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
      @media (max-width: 480px) {
        .invitations-container {
          padding: 16px;
        }
      }
    `,
  ],
})
export class PendingInvitationsComponent implements OnInit {
  private invitationService = inject(UserInvitationService);
  private organizationService = inject(OrganizationService);
  private snackBar = inject(MatSnackBar);
  private router = inject(Router);
  private store = inject(Store<AppState>);

  loading = signal(false);
  invitations = signal<UserInvitation[]>([]);
  organizations = signal<Organization[]>([]);
  callerOrgName = signal<string>('');
  selectedOrgId: number | null = null;

  private callerUser = toSignal(this.store.select(AuthSelectors.selectUser));
  private callerOrganizations = toSignal(this.store.select(AuthSelectors.selectUserOrganizations), {
    initialValue: [],
  });
  isSuperAdmin = computed(() => this.callerUser()?.role === 'SUPER_ADMIN');

  displayedColumns = ['email', 'role', 'expires', 'created', 'actions'];

  ngOnInit(): void {
    // GET /organizations (list-all) is SUPER_ADMIN-only on the backend — a
    // caller scoped to a single org (ORG_ADMIN/Admin) has no reason to see
    // every organization anyway, so they skip straight to their own org
    // instead of 403ing on a dropdown they'd never need.
    this.store.select(AuthSelectors.selectUser).subscribe((user) => {
      if (!user) return;

      if (user.role === 'SUPER_ADMIN') {
        this.organizationService.getOrganizations().subscribe({
          next: (res) => {
            this.organizations.set(res.organizations);
            if (res.organizations.length > 0) {
              this.selectedOrgId = res.organizations[0].id;
              this.loadInvitations(this.selectedOrgId);
            }
          },
          error: () => notifyError(this.snackBar, 'Failed to load organizations'),
        });
      } else if (user.organization_id) {
        this.selectedOrgId = user.organization_id;
        this.loadInvitations(user.organization_id);
        // GET /organizations/:id is SUPER_ADMIN-only too — the caller's own
        // org name is already sitting in their login-derived org list, so
        // look it up there instead of a call that would just 403.
        const org = this.callerOrganizations().find(
          (o) => o.organization_id === user.organization_id
        );
        this.callerOrgName.set(org?.organization.name || '');
      }
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
        notifyError(this.snackBar, 'Failed to load invitations');
        this.loading.set(false);
      },
    });
  }

  revokeInvitation(inv: UserInvitation): void {
    const orgId = this.selectedOrgId;
    if (!orgId) return;

    const previousInvitations = this.invitations();
    this.invitations.set(previousInvitations.filter((i) => i.id !== inv.id));

    actWithUndo(
      this.snackBar,
      `Invitation for ${inv.email} revoked`,
      () => {
        this.invitationService.revokeInvitation(inv.id, orgId).subscribe({
          error: (err) => {
            notifyError(this.snackBar, err.error?.error || 'Failed to revoke invitation');
            this.loadInvitations(orgId);
          },
        });
      },
      { onUndo: () => this.invitations.set(previousInvitations) }
    );
  }

  inviteUser(): void {
    this.router.navigate(['/admin/invitations/new']);
  }
}
