import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, Router, RouterModule } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { take } from 'rxjs/operators';
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
  ],
  templateUrl: './app.html',
  styleUrls: ['./app.scss'],
})
export class App implements OnInit {
  public authFacade = inject(AuthFacade);
  private router = inject(Router);

  // Observable streams from NgRx store
  isAuthenticated$ = this.authFacade.isAuthenticated$;
  userRole$ = this.authFacade.userRole$;
  isAdmin$ = this.authFacade.isAdmin$;
  isProp= this.authFacade.user$;

  // Temporary fallback for debugging
  get hasTokenInStorage(): boolean {
    return !!localStorage.getItem('tenantly_token');
  }

  ngOnInit() {
    // Initialize auth state from localStorage
    this.authFacade.initializeAuth();
    
    // Debug: Log authentication status
    this.isAuthenticated$.subscribe(isAuth => {
      console.log('Authentication status:', isAuth);
    });
  }

  logout() {
    this.authFacade.logout();
  }

  // Backward compatibility method
  isAdmin(): boolean {
    // This is synchronous for template usage
    // For reactive usage, use isAdmin$ observable
    let isAdmin = false;
    this.isAdmin$.pipe(take(1)).subscribe(admin => isAdmin = admin);
    return isAdmin;
  }
}
