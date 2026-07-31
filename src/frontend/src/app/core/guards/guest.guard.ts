import { inject, Injectable } from '@angular/core';
import { CanActivate, Router } from '@angular/router';
import { Observable } from 'rxjs';
import { map, take } from 'rxjs/operators';
import { AuthFacade } from '../../store/auth/auth.facade';

@Injectable({
  providedIn: 'root',
})
export class GuestGuard implements CanActivate {
  private authFacade = inject(AuthFacade);
  private router = inject(Router);

  canActivate(): Observable<boolean> {
    return this.authFacade.isAuthenticated$.pipe(
      take(1),
      map((isAuthenticated) => {
        if (!isAuthenticated) {
          const token = localStorage.getItem('tenantly_token');
          const expiresAt = localStorage.getItem('tenantly_expires_at');
          if (token && expiresAt && new Date() < new Date(expiresAt)) {
            this.router.navigate(['/dashboard']);
            return false;
          }
          return true;
        }
        this.router.navigate(['/dashboard']);
        return false;
      })
    );
  }
}
