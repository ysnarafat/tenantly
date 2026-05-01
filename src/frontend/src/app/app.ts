import { Component, inject, OnInit, ViewChild, signal, computed } from '@angular/core';

import { RouterOutlet, RouterModule } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSidenavModule, MatSidenav } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatMenuModule } from '@angular/material/menu';
import { MatDividerModule } from '@angular/material/divider';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AuthFacade } from './store/auth/auth.facade';
import { PermissionService } from './core/services/permission.service';
import { Permission } from './core/models/role.model';
import { User } from './core/services/auth.service';
import { OrganizationSelector } from './shared/organization-selector/organization-selector';

const AVATAR_COLORS = [
  '#1565c0',
  '#2e7d32',
  '#c62828',
  '#6a1b9a',
  '#0277bd',
  '#e65100',
  '#37474f',
  '#00695c',
];

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    RouterOutlet,
    RouterModule,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatSidenavModule,
    MatListModule,
    MatTooltipModule,
    MatMenuModule,
    MatDividerModule,
    OrganizationSelector,
  ],
  templateUrl: './app.html',
  styleUrls: ['./app.scss'],
})
export class App implements OnInit {
  @ViewChild('sidenav') sidenav!: MatSidenav;

  public authFacade = inject(AuthFacade);
  public permissions = inject(PermissionService);
  private breakpointObserver = inject(BreakpointObserver);

  isAuthenticated = signal(false);
  userRole = signal('');
  user = signal<User | null>(null);
  isMobile = signal(false);
  sidenavCollapsed = signal(false);

  // Computed permission signals
  isSuperAdmin = computed(() => this.permissions.isSuperAdmin());
  isOrgAdmin = computed(() => this.permissions.isOrgAdmin());
  canManageProperties = computed(() =>
    this.permissions.hasPermission(Permission.MANAGE_PROPERTIES)
  );
  canManageTenants = computed(() => this.permissions.hasPermission(Permission.MANAGE_TENANTS));
  canManageDocuments = computed(() => this.permissions.hasPermission(Permission.MANAGE_DOCUMENTS));
  canManageUsers = computed(() => this.permissions.hasPermission(Permission.MANAGE_USERS));
  canManageOrganizations = computed(() => this.permissions.canManageOrganizations());
  canInviteUsers = computed(() => this.permissions.canInviteUsers());
  canViewPayments = computed(() => this.permissions.hasPermission(Permission.VIEW_PAYMENTS));
  canViewReports = computed(() => this.permissions.hasPermission(Permission.VIEW_REPORTS));

  // Computed derived state
  sidenavMode = computed(() => (this.isMobile() ? ('over' as const) : ('side' as const)));
  sidenavOpened = computed(() => !this.isMobile());
  fixedTopGap = computed(() => (this.isMobile() ? 64 : 0));

  sidenavWidth = computed(() => {
    if (this.isMobile()) return '280px';
    return this.sidenavCollapsed() ? '64px' : '260px';
  });

  // User avatar
  userInitials = computed(() => {
    const name = this.user()?.username || '';
    return name.slice(0, 2).toUpperCase() || '?';
  });

  userAvatarColor = computed(() => {
    const name = this.user()?.username || '';
    const idx = (name.charCodeAt(0) || 0) % AVATAR_COLORS.length;
    return AVATAR_COLORS[idx];
  });

  hasAdminNav = computed(
    () =>
      this.canManageOrganizations() ||
      this.canInviteUsers() ||
      this.isOrgAdmin() ||
      this.isSuperAdmin() ||
      this.canManageUsers()
  );

  hasWorkspaceNav = computed(
    () =>
      this.canManageProperties() ||
      this.canManageTenants() ||
      this.canViewPayments() ||
      this.canViewReports() ||
      this.canManageDocuments()
  );

  constructor() {
    this.authFacade.isAuthenticated$
      .pipe(takeUntilDestroyed())
      .subscribe((isAuth) => this.isAuthenticated.set(isAuth));

    this.authFacade.userRole$
      .pipe(takeUntilDestroyed())
      .subscribe((role) => this.userRole.set(role));

    this.authFacade.user$.pipe(takeUntilDestroyed()).subscribe((user) => this.user.set(user));

    this.breakpointObserver
      .observe([Breakpoints.Handset])
      .pipe(takeUntilDestroyed())
      .subscribe((result) => {
        this.isMobile.set(result.matches);
        if (result.matches) this.sidenavCollapsed.set(false);
      });
  }

  ngOnInit() {
    this.authFacade.initializeAuth();
  }

  logout() {
    this.authFacade.logout();
  }

  onNavigate() {
    if (this.isMobile() && this.sidenav) {
      this.sidenav.close();
    }
  }

  toggleSidenav() {
    if (this.isMobile()) {
      this.sidenav?.toggle();
    } else {
      this.sidenavCollapsed.update((v) => !v);
    }
  }
}
