import { Component, inject, OnInit, ViewChild, signal, computed, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, RouterModule } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSidenavModule, MatSidenav } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatTooltipModule } from '@angular/material/tooltip';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AuthFacade } from './store/auth/auth.facade';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    RouterModule,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatSidenavModule,
    MatListModule,
    MatTooltipModule,
  ],
  templateUrl: './app.html',
  styleUrls: ['./app.scss'],
})
export class App implements OnInit {
  @ViewChild('sidenav') sidenav!: MatSidenav;

  public authFacade = inject(AuthFacade);
  private breakpointObserver = inject(BreakpointObserver);

  // Signals for reactive state
  isAuthenticated = signal(false);
  userRole = signal('');
  isAdmin = signal(false);
  isPropertyManager = signal(false);
  isMobile = signal(false);

  // Computed signals for derived state
  sidenavMode = computed(() => this.isMobile() ? 'over' as const : 'side' as const);
  sidenavOpened = computed(() => !this.isMobile());
  fixedTopGap = computed(() => this.isMobile() ? 64 : 0);

  constructor() {
    // Subscribe to observables and update signals
    this.authFacade.isAuthenticated$
      .pipe(takeUntilDestroyed())
      .subscribe(isAuth => this.isAuthenticated.set(isAuth));

    this.authFacade.userRole$
      .pipe(takeUntilDestroyed())
      .subscribe(role => this.userRole.set(role));

    this.authFacade.isAdmin$
      .pipe(takeUntilDestroyed())
      .subscribe(isAdmin => this.isAdmin.set(isAdmin));

    this.authFacade.isPropertyManager$
      .pipe(takeUntilDestroyed())
      .subscribe(isPM => this.isPropertyManager.set(isPM));

    this.breakpointObserver.observe([Breakpoints.Handset])
      .pipe(takeUntilDestroyed())
      .subscribe(result => this.isMobile.set(result.matches));
  }

  ngOnInit() {
    // Initialize auth state from localStorage
    this.authFacade.initializeAuth();
  }

  logout() {
    this.authFacade.logout();
  }

  // Close sidenav on mobile after navigation
  onNavigate() {
    if (this.isMobile() && this.sidenav) {
      this.sidenav.close();
    }
  }

  // Toggle sidenav
  toggleSidenav() {
    if (this.sidenav) {
      this.sidenav.toggle();
    }
  }
}
