import { Component, inject, OnInit, ViewChild, signal, computed, effect } from '@angular/core';

import {
  RouterOutlet,
  RouterModule,
  Router,
  NavigationEnd,
  NavigationStart,
  NavigationCancel,
  NavigationError,
  NavigationSkipped,
} from '@angular/router';
import { filter, switchMap, map } from 'rxjs/operators';
import { timer, of, EMPTY } from 'rxjs';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSidenavModule, MatSidenav, MatSidenavContainer } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatMenuModule } from '@angular/material/menu';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { TranslateModule } from '@ngx-translate/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AuthFacade } from './store/auth/auth.facade';
import { PermissionService } from './core/services/permission.service';
import { Permission } from './core/models/role.model';
import { User } from './core/services/auth.service';
import { LanguageService } from './core/services/language.service';
import { ThemeService } from './core/services/theme.service';
import { OrganizationSelector } from './shared/organization-selector/organization-selector';
import { avatarColorFor, avatarInitials } from './shared/utils/avatar.utils';

const AUTH_ROUTE_PREFIXES = [
  '/home',
  '/login',
  '/select-organization',
  '/401',
  '/unauthorized',
  '/404',
];

const SIDENAV_COLLAPSED_KEY = 'tenantly-sidenav-collapsed';

function getStoredSidenavCollapsed(): boolean {
  return localStorage.getItem(SIDENAV_COLLAPSED_KEY) === 'true';
}

// Only surface the bar once a navigation has been pending this long — most
// route changes resolve near-instantly (chunk already cached), and flashing
// a loader for those reads as jank rather than feedback.
const NAVIGATION_LOADER_DELAY_MS = 150;

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
    MatProgressBarModule,
    TranslateModule,
    OrganizationSelector,
  ],
  templateUrl: './app.html',
  styleUrls: ['./app.scss'],
})
export class App implements OnInit {
  @ViewChild('sidenav') sidenav!: MatSidenav;
  @ViewChild(MatSidenavContainer) sidenavContainer!: MatSidenavContainer;

  public authFacade = inject(AuthFacade);
  public permissions = inject(PermissionService);
  public languageService = inject(LanguageService);
  public themeService = inject(ThemeService);
  private breakpointObserver = inject(BreakpointObserver);
  private router = inject(Router);

  isAuthenticated = signal(false);
  currentUrl = signal(this.router.url);
  userRole = signal('');
  user = signal<User | null>(null);
  isMobile = signal(false);
  sidenavCollapsed = signal(getStoredSidenavCollapsed());
  navigating = signal(false);

  // Computed permission signals
  isSuperAdmin = computed(() => this.permissions.isSuperAdmin());
  isOrgAdmin = computed(() => this.permissions.isOrgAdmin());
  // Workspace nav is gated on view access, so read-only roles (e.g. Accountant)
  // still see the sections they're allowed to open.
  canViewProperties = computed(() => this.permissions.hasPermission(Permission.VIEW_PROPERTIES));
  canViewTenants = computed(() => this.permissions.hasPermission(Permission.VIEW_TENANTS));
  canViewLeases = computed(() => this.permissions.hasPermission(Permission.VIEW_LEASES));
  canViewPayments = computed(() => this.permissions.hasPermission(Permission.VIEW_PAYMENTS));
  canViewReports = computed(() => this.permissions.hasPermission(Permission.VIEW_REPORTS));
  canViewDocuments = computed(() => this.permissions.hasPermission(Permission.VIEW_DOCUMENTS));

  // Admin nav is gated on management capability, matching the route guards.
  canManageUsers = computed(() => this.permissions.hasPermission(Permission.MANAGE_USERS));
  canManageOrganizations = computed(() => this.permissions.canManageOrganizations());
  canInviteUsers = computed(() => this.permissions.canInviteUsers());

  // Computed derived state
  sidenavMode = computed(() => (this.isMobile() ? ('over' as const) : ('side' as const)));
  sidenavOpened = computed(() => !this.isMobile());
  fixedTopGap = computed(() => (this.isMobile() ? 64 : 0));

  // User avatar
  userInitials = computed(() => avatarInitials(this.user()?.username || ''));

  userAvatarColor = computed(() => avatarColorFor(this.user()?.username || ''));

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
      this.canViewProperties() ||
      this.canViewTenants() ||
      this.canViewLeases() ||
      this.canViewPayments() ||
      this.canViewReports() ||
      this.canViewDocuments()
  );

  isOnAuthRoute = computed(() =>
    AUTH_ROUTE_PREFIXES.some((prefix) => this.currentUrl().startsWith(prefix))
  );

  showShell = computed(() => this.isAuthenticated() && !this.isOnAuthRoute());

  constructor() {
    this.authFacade.isAuthenticated$
      .pipe(takeUntilDestroyed())
      .subscribe((isAuth) => this.isAuthenticated.set(isAuth));

    this.router.events
      .pipe(
        filter((event): event is NavigationEnd => event instanceof NavigationEnd),
        takeUntilDestroyed()
      )
      .subscribe((event) => this.currentUrl.set(event.urlAfterRedirects));

    // switchMap cancels the pending delay timer as soon as a NavigationEnd/
    // Cancel/Error/Skipped event arrives, so a fast navigation never flashes
    // the loader — only one that outlives NAVIGATION_LOADER_DELAY_MS does.
    this.router.events
      .pipe(
        switchMap((event) => {
          if (event instanceof NavigationStart) {
            return timer(NAVIGATION_LOADER_DELAY_MS).pipe(map(() => true));
          }
          if (
            event instanceof NavigationEnd ||
            event instanceof NavigationCancel ||
            event instanceof NavigationError ||
            event instanceof NavigationSkipped
          ) {
            return of(false);
          }
          return EMPTY;
        }),
        takeUntilDestroyed()
      )
      .subscribe((navigating) => this.navigating.set(navigating));

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

    effect(() => {
      const lang = this.languageService.currentLang();
      document.documentElement.setAttribute('lang', lang);
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
      this.sidenavCollapsed.update((v) => {
        const next = !v;
        localStorage.setItem(SIDENAV_COLLAPSED_KEY, String(next));
        return next;
      });
      // MatSidenavContainer only recalculates the content margin on drawer
      // open/close or viewport resize — a pure CSS width change (our
      // .collapsed class) isn't one of its triggers, so the reserved space
      // for the sidenav never updates on its own. Nudge it manually, once
      // now and once after the width transition ($nav-transition in
      // app.scss) finishes so the final width is captured.
      this.sidenavContainer?.updateContentMargins();
      setTimeout(() => this.sidenavContainer?.updateContentMargins(), 250);
    }
  }
}
