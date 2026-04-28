import { Component, OnInit, OnDestroy, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { Router } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { Store } from '@ngrx/store';
import { AppState } from '../../store';
import * as AuthSelectors from '../../store/auth/auth.selectors';
import * as AuthActions from '../../store/auth/auth.actions';
import { OrganizationService } from '../../core/services/organization.service';
import { UserOrganization } from '../../core/models/organization.model';

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
  ],
  template: `
    <div class="org-selector">
      <!-- Organization Selector Button -->
      <button
        mat-button
        [matMenuTriggerFor]="orgMenu"
        class="org-button"
        [disabled]="isLoading()"
        matTooltip="Switch Organization"
      >
        <mat-icon class="org-icon">domain</mat-icon>
        <span class="org-name">{{ currentOrgName() }}</span>
        <mat-icon class="dropdown-icon">expand_more</mat-icon>
      </button>

      <!-- Organization Menu -->
      <mat-menu #orgMenu="matMenu" class="org-menu">
        <!-- Current Organization Indicator -->
        <div class="menu-header">
          <span class="menu-title">Organizations</span>
        </div>

        @for (org of userOrganizations(); track org.id) {
          <button
            mat-menu-item
            (click)="selectOrganization(org)"
            class="org-menu-item"
            [class.active]="isCurrentOrg(org.organization_id)"
          >
            <mat-icon class="org-icon-menu">
              @if (isCurrentOrg(org.organization_id)) {
                check_circle
              } @else {
                radio_button_unchecked
              }
            </mat-icon>
            <div class="org-menu-content">
              <span class="org-menu-name">{{ org.organization.name }}</span>
              <span class="org-menu-role">{{ org.role }}</span>
            </div>
          </button>
        }

        @if (userOrganizations().length > 0) {
          <mat-divider class="menu-divider"></mat-divider>
        }

        <!-- Manage Organizations -->
        <button mat-menu-item (click)="manageOrganizations()" class="manage-orgs">
          <mat-icon>settings</mat-icon>
          <span>Manage Organizations</span>
        </button>
      </mat-menu>

      @if (isLoading()) {
        <mat-spinner diameter="20" class="loading-spinner"></mat-spinner>
      }
    </div>
  `,
  styles: [
    `
      .org-selector {
        display: flex;
        align-items: center;
        position: relative;
      }

      .org-button {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 4px 12px;
        border-radius: 20px;
        transition: all 0.3s ease;
        background-color: rgba(255, 255, 255, 0.1);

        &:hover:not(:disabled) {
          background-color: rgba(255, 255, 255, 0.2);
        }

        &:disabled {
          opacity: 0.7;
        }
      }

      .org-icon {
        font-size: 18px;
        width: 18px;
        height: 18px;
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
        margin-left: 4px;
        transition: transform 0.3s ease;
      }

      .org-button:disabled .dropdown-icon {
        transform: rotate(180deg);
      }

      .loading-spinner {
        position: absolute;
        right: -35px;
      }

      /* Menu Styles */
      ::ng-deep .org-menu {
        .mat-mdc-menu-content {
          padding: 0 !important;
          max-height: 400px;
          overflow-y: auto;
        }

        .mat-mdc-menu-item {
          height: auto !important;
          line-height: normal !important;
          padding: 8px 12px !important;
        }
      }

      .menu-header {
        padding: 12px 16px;
        font-size: 12px;
        font-weight: 600;
        color: rgba(0, 0, 0, 0.54);
        text-transform: uppercase;
        letter-spacing: 0.5px;
        border-bottom: 1px solid rgba(0, 0, 0, 0.12);
      }

      .menu-title {
        display: block;
      }

      .org-menu-item {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        padding: 8px 12px !important;
        transition: background-color 0.2s ease;

        &:hover {
          background-color: rgba(0, 0, 0, 0.04);
        }

        &.active {
          background-color: rgba(63, 81, 181, 0.08);
          color: #3f51b5;
        }
      }

      .org-icon-menu {
        flex-shrink: 0;
        font-size: 20px;
        width: 20px;
        height: 20px;
        margin-top: 2px;
      }

      .org-menu-content {
        display: flex;
        flex-direction: column;
        gap: 2px;
        flex: 1;
      }

      .org-menu-name {
        display: block;
        font-size: 14px;
        font-weight: 500;
        color: rgba(0, 0, 0, 0.87);
      }

      .org-menu-role {
        display: block;
        font-size: 12px;
        color: rgba(0, 0, 0, 0.54);
      }

      .menu-divider {
        margin: 8px 0 !important;
      }

      .manage-orgs {
        color: #3f51b5;
        font-weight: 500;

        mat-icon {
          margin-right: 8px;
        }
      }
    `,
  ],
})
export class OrganizationSelector implements OnInit, OnDestroy {
  private router = inject(Router);
  private orgService = inject(OrganizationService);
  private store = inject(Store<AppState>);
  private destroy$ = new Subject<void>();

  // State signals
  userOrganizations = signal<UserOrganization[]>([]);
  currentOrganizationId = signal<number | null>(null);
  isLoading = signal(false);
  currentOrgName = computed(() => {
    const orgs = this.userOrganizations();
    const currentId = this.currentOrganizationId();
    if (!currentId || orgs.length === 0) return 'Organization';

    const currentOrg = orgs.find((org) => org.organization_id === currentId);
    return currentOrg?.organization.name || 'Organization';
  });

  ngOnInit() {
    // Initialize organizations from store
    this.store
      .select(AuthSelectors.selectUserOrganizations)
      .pipe(takeUntil(this.destroy$))
      .subscribe((orgs) => {
        this.userOrganizations.set(orgs || []);
      });

    // Initialize current organization
    this.store
      .select(AuthSelectors.selectCurrentOrganizationId)
      .pipe(takeUntil(this.destroy$))
      .subscribe((orgId) => {
        this.currentOrganizationId.set(orgId);
        this.isLoading.set(false);
      });

    // Restore organization context on init
    this.orgService.restoreOrganizationContext();
  }

  selectOrganization(organization: UserOrganization) {
    if (this.isCurrentOrg(organization.organization_id)) {
      return; // Already selected
    }

    this.isLoading.set(true);
    this.store.dispatch(
      AuthActions.switchOrganization({ organizationId: organization.organization_id })
    );
  }

  isCurrentOrg(orgId: number): boolean {
    return this.currentOrganizationId() === orgId;
  }

  manageOrganizations() {
    // Navigate to organization management page
    this.router.navigate(['/settings/organizations']);
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
