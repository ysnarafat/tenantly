import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { catchError, switchMap, throwError, take, filter } from 'rxjs';
import { AuthService } from '../services/auth.service';
import { AppState } from '../../store';
import * as AuthSelectors from '../../store/auth/auth.selectors';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const store = inject(Store<AppState>);
  const router = inject(Router);

  // Skip auth for login and refresh endpoints
  if (
    req.url.includes('/auth/login') ||
    req.url.includes('/auth/refresh') ||
    req.url.includes('/auth/reset-password') ||
    req.url.includes('/auth/confirm-reset-password')
  ) {
    return next(req);
  }

  // Get token from NgRx store
  let token: string | null = null;
  let isTokenExpired = false;

  store
    .select(AuthSelectors.selectToken)
    .pipe(take(1))
    .subscribe((t) => (token = t));
  store
    .select(AuthSelectors.selectIsTokenExpired)
    .pipe(take(1))
    .subscribe((expired) => (isTokenExpired = expired));

  if (token && !isTokenExpired) {
    const authReq = req.clone({
      headers: req.headers.set('Authorization', `Bearer ${token}`),
    });

    return next(authReq).pipe(
      catchError((error: HttpErrorResponse) => {
        if (error.status === 401) {
          // Token might be expired, try to refresh
          authService.refreshToken();

          // Wait for refresh to complete and retry
          return store.select(AuthSelectors.selectToken).pipe(
            filter((newToken) => !!newToken && newToken !== token),
            take(1),
            switchMap((newToken) => {
              const retryReq = req.clone({
                headers: req.headers.set('Authorization', `Bearer ${newToken}`),
              });
              return next(retryReq);
            }),
            catchError((refreshError) => {
              // Refresh failed, redirect to login
              authService.logout();
              router.navigate(['/login']);
              return throwError(() => refreshError);
            })
          );
        }
        return throwError(() => error);
      })
    );
  } else if (localStorage.getItem('tenantly_refresh_token') && isTokenExpired) {
    // Token is expired but we have a refresh token
    authService.refreshToken();

    return store.select(AuthSelectors.selectToken).pipe(
      filter((newToken) => !!newToken),
      take(1),
      switchMap((newToken) => {
        const authReq = req.clone({
          headers: req.headers.set('Authorization', `Bearer ${newToken}`),
        });
        return next(authReq);
      }),
      catchError((error) => {
        // Refresh failed, redirect to login
        authService.logout();
        router.navigate(['/login']);
        return throwError(() => error);
      })
    );
  }

  return next(req);
};
