import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import {
  Observable,
  catchError,
  switchMap,
  throwError,
  take,
  filter,
  finalize,
  shareReplay,
} from 'rxjs';
import { AuthService } from '../services/auth.service';
import { AppState } from '../../store';
import * as AuthSelectors from '../../store/auth/auth.selectors';
import * as AuthActions from '../../store/auth/auth.actions';

// Shared across all interceptor invocations (module-level, not per-call) so
// that N concurrent requests hitting a 401 (or an expired token) at once
// trigger exactly one refresh dispatch instead of one each — avoids a stampede
// of parallel /auth/refresh calls racing to rotate the same refresh token.
let refreshInFlight$: Observable<string | null> | null = null;

function triggerRefresh(
  authService: AuthService,
  store: Store<AppState>,
  previousToken: string | null
): Observable<string | null> {
  if (!refreshInFlight$) {
    authService.refreshToken();
    refreshInFlight$ = store.select(AuthSelectors.selectToken).pipe(
      filter((newToken) => !!newToken && newToken !== previousToken),
      take(1),
      finalize(() => {
        refreshInFlight$ = null;
      }),
      shareReplay(1)
    );
  }
  return refreshInFlight$;
}

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const store = inject(Store<AppState>);
  const router = inject(Router);

  // Skip auth for login/refresh endpoints and public static assets (e.g. i18n
  // JSON, fetched on every page including fully anonymous ones). Routing a
  // translation-file request through the expired-token refresh dance below
  // would fire a silent /auth/refresh attempt — and on failure, force-navigate
  // to /login — on pages that have nothing to do with authentication, such as
  // the 401/404 error pages themselves.
  if (
    req.url.includes('/auth/login') ||
    req.url.includes('/auth/refresh') ||
    req.url.includes('/auth/reset-password') ||
    req.url.includes('/auth/confirm-reset-password') ||
    req.url.includes('assets/i18n/')
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
        if (
          error.status === 403 &&
          ['TOKEN_AUDIENCE_MISMATCH', 'OPERATIONS_PORTAL_REQUIRED'].includes(error.error?.code)
        ) {
          store.dispatch(AuthActions.logoutSuccess());
          return throwError(() => error);
        }
        if (error.status === 401) {
          // Token might be expired, try to refresh (deduped — see triggerRefresh)
          return triggerRefresh(authService, store, token).pipe(
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
  } else if (isTokenExpired && token !== null) {
    // Only attempt a refresh if we actually have a token to refresh.
    // When there's no token at all (e.g. first visit, or expired refresh
    // cookie), there's nothing to refresh — just proceed without auth.
    // Token is expired — attempt a refresh. The httpOnly refresh cookie (if
    // any) is sent automatically by the browser on the /auth/refresh call;
    // there's nothing to check in localStorage anymore.
    return triggerRefresh(authService, store, token).pipe(
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
