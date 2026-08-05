import { Component, OnInit, OnDestroy, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { Router } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { Store } from '@ngrx/store';
import { Actions, ofType } from '@ngrx/effects';
import { AppState } from '../../store';
import * as AuthSelectors from '../../store/auth/auth.selectors';
import * as AuthActions from '../../store/auth/auth.actions';
import { OrganizationService } from '../../core/services/organization.service';
import { UserOrganization } from '../../core/models/organization.model';
import { avatarColorFor, avatarInitials } from '../utils/avatar.utils';

const ROLE_LABELS: Record<string, string> = {
  SUPER_ADMIN: 'Super Admin',
  ORG_ADMIN: 'Org Admin',
  Admin: 'Admin',
  PropertyManager: 'Property Manager',
  Accountant: 'Accountant',
};

@Component({
  selector: 'app-organization-selector',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatIconModule,
    MatMenuModule,
    MatTooltipModule,
    MatDividerModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
  ],
  template: `
    <div class="org-selector">
      <button
        mat-button
        [matMenuTriggerFor]="orgMenu"
        class="org-button"
        [disabled]="isLoading()"
        matTooltip="Switch organization"
        aria-label="Switch organization"
      >
        <!--
          Everything lives inside one wrapper span rather than as direct
          children of the button. mat-button internally re-projects any
          top-level <mat-icon> into its own slot, separate from the rest of
          the label — which split the chevron from the avatar/name onto a
          different line. Angular's content projection only inspects direct
          children, so nesting the icon in here keeps the whole row intact
          regardless of how Material's internals handle it.
        -->
        <span class="org-button-content">
          <span
            class="org-avatar"
            [style.background-color]="avatarColorFor(currentOrgName())"
            aria-hidden="true"
          >
            {{ avatarInitials(currentOrgName()) }}
          </span>
          <span class="org-name">{{ currentOrgName() }}</span>
          @if (isLoading()) {
            <mat-spinner class="org-spinner" diameter="16"></mat-spinner>
          } @else {
            <mat-icon class="dropdown-icon">expand_more</mat-icon>
          }
        </span>
      </button>
      <span class="sr-only" role="status" aria-live="polite">
        {{ isLoading() ? 'Switching organization…' : '' }}
      </span>

      <mat-menu #orgMenu="matMenu" class="org-menu">
        <div class="menu-header">
          <span class="menu-title">Organizations</span>
        </div>

        @if (userOrganizations().length === 0) {
          <div class="empty-orgs">No organizations yet</div>
        }

        @for (org of userOrganizations(); track org.id) {
          <button
            mat-menu-item
            (click)="selectOrganization(org)"
            class="org-menu-item"
            [class.active]="isCurrentOrg(org.organization_id)"
            [attr.aria-current]="isCurrentOrg(org.organization_id) ? 'true' : null"
          >
            <!-- Single wrapper span, same reason as the trigger button above:
                 mat-menu-item projects <mat-icon> into its own slot too. -->
            <span class="org-menu-row">
              <span
                class="org-avatar org-avatar-sm"
                [style.background-color]="avatarColorFor(org.organization.name)"
                aria-hidden="true"
              >
                {{ avatarInitials(org.organization.name) }}
              </span>
              <span class="org-menu-content">
                <span class="org-menu-name">{{ org.organization.name }}</span>
                <span class="org-menu-role">{{ formatRole(org.role) }}</span>
              </span>
              @if (isCurrentOrg(org.organization_id)) {
                <mat-icon class="org-check" aria-hidden="true">check_circle</mat-icon>
              }
            </span>
          </button>
        }

        @if (isSuperAdmin()) {
          <mat-divider class="menu-divider"></mat-divider>
          <button mat-menu-item (click)="manageOrganizations()" class="manage-orgs">
            <span class="manage-orgs-row">
              <mat-icon aria-hidden="true">settings</mat-icon>
              <span>Manage Organizations</span>
            </span>
          </button>
        }
      </mat-menu>
    </div>
  `,
  styles: [
    `
      .org-selector {
        display: flex;
        align-items: center;
        position: relative;
      }

      .sr-only {
        position: absolute;
        width: 1px;
        height: 1px;
        overflow: hidden;
        clip: rect(0, 0, 0, 0);
      }

      .org-button {
        // padding !important: Material's own .mat-mdc-button:has(mat-icon)
        // rule overrides padding at higher specificity than a plain class
        // selector; this needs to win so the pill shape stays consistent.
        padding: 4px 12px 4px 4px !important;
        border-radius: 20px;
        transition: background-color 0.2s ease;
        background-color: rgba(255, 255, 255, 0.1);

        &:hover:not(:disabled) {
          background-color: rgba(255, 255, 255, 0.2);
        }

        &:disabled {
          opacity: 0.8;
        }
      }

      .org-button-content {
        display: flex;
        flex-wrap: nowrap;
        align-items: center;
        gap: 8px;
        white-space: nowrap;
      }

      .org-avatar {
        flex-shrink: 0;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        height: 24px;
        border-radius: 50%;
        color: #fff;
        font-size: 11px;
        font-weight: 700;
        letter-spacing: 0.2px;
      }

      .org-avatar-sm {
        width: 28px;
        height: 28px;
        font-size: 12px;
        margin-top: 2px;
      }

      .org-name {
        font-size: 13px;
        font-weight: 500;
        max-width: 150px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .dropdown-icon {
        font-size: 18px;
        width: 18px;
        height: 18px;
        transition: transform 0.2s ease;
      }

      .org-button[aria-expanded='true'] .dropdown-icon {
        transform: rotate(180deg);
      }

      .org-spinner {
        ::ng-deep circle {
          stroke: currentColor;
        }
      }

      ::ng-deep .org-menu {
        .mat-mdc-menu-content {
          padding: 0 !important;
          max-height: 420px;
          overflow-y: auto;
          width: 300px;
        }

        .mat-mdc-menu-item {
          height: auto !important;
          line-height: normal !important;
          padding: 10px 16px !important;
        }
      }

      .menu-header {
        padding: 12px 16px;
        font-size: 12px;
        font-weight: 600;
        color: var(--text-secondary);
        text-transform: uppercase;
        letter-spacing: 0.5px;
        border-bottom: 1px solid var(--border-color);
      }

      .empty-orgs {
        padding: 16px;
        font-size: 13px;
        color: var(--text-hint);
        text-align: center;
      }

      .org-menu-item.active {
        background-color: var(--bg-selected);
      }

      .org-menu-row {
        display: flex;
        flex-wrap: nowrap;
        align-items: flex-start;
        gap: 12px;
        width: 100%;
      }

      .org-menu-content {
        display: flex;
        flex-direction: column;
        gap: 2px;
        flex: 1;
        min-width: 0;
      }

      .org-menu-name {
        display: block;
        font-size: 14px;
        font-weight: 500;
        color: var(--text-primary);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .org-menu-role {
        display: block;
        font-size: 12px;
        color: var(--text-secondary);
      }

      .org-check {
        flex-shrink: 0;
        margin-top: 2px;
        color: var(--color-primary);
      }

      .menu-divider {
        margin: 4px 0 !important;
      }

      .manage-orgs {
        color: var(--color-primary);
      }

      .manage-orgs-row {
        display: flex;
        flex-wrap: nowrap;
        align-items: center;
        gap: 8px;
      }
    `,
  ],
})
export class OrganizationSelector implements OnInit, OnDestroy {
  private router = inject(Router);
  private orgService = inject(OrganizationService);
  private store = inject(Store<AppState>);
  private actions$ = inject(Actions);
  private snackBar = inject(MatSnackBar);
  private destroy$ = new Subject<void>();

  readonly avatarColorFor = avatarColorFor;
  readonly avatarInitials = avatarInitials;

  userOrganizations = signal<UserOrganization[]>([]);
  currentOrganizationId = signal<number | null>(null);
  isSuperAdmin = signal(false);
  isLoading = signal(false);

  currentOrgName = computed(() => {
    const orgs = this.userOrganizations();
    const currentId = this.currentOrganizationId();
    if (!currentId || orgs.length === 0) return 'Organization';

    const currentOrg = orgs.find((org) => org.organization_id === currentId);
    return currentOrg?.organization.name || 'Organization';
  });

  ngOnInit() {
    this.store
      .select(AuthSelectors.selectUserOrganizations)
      .pipe(takeUntil(this.destroy$))
      .subscribe((orgs) => {
        this.userOrganizations.set(orgs || []);
      });

    this.store
      .select(AuthSelectors.selectCurrentOrganizationId)
      .pipe(takeUntil(this.destroy$))
      .subscribe((orgId) => {
        this.currentOrganizationId.set(orgId);
      });

    this.store
      .select(AuthSelectors.selectIsSuperAdmin)
      .pipe(takeUntil(this.destroy$))
      .subscribe((isSA) => this.isSuperAdmin.set(isSA));

    this.actions$
      .pipe(ofType(AuthActions.switchOrganizationSuccess), takeUntil(this.destroy$))
      .subscribe(() => this.isLoading.set(false));

    this.actions$
      .pipe(ofType(AuthActions.switchOrganizationFailure), takeUntil(this.destroy$))
      .subscribe(() => {
        this.isLoading.set(false);
        this.snackBar.open('Could not switch organization. Please try again.', 'Close', {
          duration: 4000,
        });
      });

    this.orgService.restoreOrganizationContext();
  }

  selectOrganization(organization: UserOrganization) {
    if (this.isCurrentOrg(organization.organization_id)) {
      return;
    }

    this.isLoading.set(true);
    this.store.dispatch(
      AuthActions.switchOrganization({ organizationId: organization.organization_id })
    );
  }

  isCurrentOrg(orgId: number): boolean {
    return this.currentOrganizationId() === orgId;
  }

  formatRole(role: string): string {
    return ROLE_LABELS[role] || role;
  }

  manageOrganizations() {
    this.router.navigate(['/admin/organizations']);
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
