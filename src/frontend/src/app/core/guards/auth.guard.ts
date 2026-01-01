import { inject, Injectable } from '@angular/core';
import { CanActivate, Router } from '@angular/router';
import { Observable, of } from 'rxjs';
import { map, take } from 'rxjs/operators';
import { AuthFacade } from '../../store/auth/auth.facade';

@Injectable({
  providedIn: 'root',
})
export class AuthGuard implements CanActivate {
  private authFacade = inject(AuthFacade);
  private router = inject(Router);

  canActivate(): Observable<boolean> {
    return this.authFacade.isAuthenticated$.pipe(
      take(1),
      map((isAuthenticated) => {
        if (isAuthenticated) {
          return true;
        } else {
          // Check localStorage as fallback
          const token = localStorage.getItem('tenantly_token');
          const expiresAt = localStorage.getItem('tenantly_expires_at');

          if (token && expiresAt) {
            const isExpired = new Date() >= new Date(expiresAt);
            if (!isExpired) {
              // Initialize auth state if token exists and is valid
              this.authFacade.initializeAuth();
              return true;
            }
          }

          this.router.navigate(['/login']);
          return false;
        }
      })
    );
  }
}
