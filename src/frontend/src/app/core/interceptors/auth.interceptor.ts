import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, switchMap, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const router = inject(Router);
  const token = authService.getToken();

  // Skip auth for login and refresh endpoints
  if (req.url.includes('/auth/login') || req.url.includes('/auth/refresh') || req.url.includes('/auth/reset-password')) {
    return next(req);
  }

  if (token && !authService.isTokenExpired()) {
    const authReq = req.clone({
      headers: req.headers.set('Authorization', `Bearer ${token}`),
    });
    
    return next(authReq).pipe(
      catchError((error: HttpErrorResponse) => {
        if (error.status === 401) {
          // Token might be expired, try to refresh
          return authService.refreshToken().pipe(
            switchMap(() => {
              // Retry the original request with new token
              const newToken = authService.getToken();
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
  } else if (authService.getRefreshToken() && authService.isTokenExpired()) {
    // Token is expired but we have a refresh token
    return authService.refreshToken().pipe(
      switchMap(() => {
        const newToken = authService.getToken();
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
