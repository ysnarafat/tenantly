import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Router } from '@angular/router';
import { of } from 'rxjs';
import { map, exhaustMap, catchError, tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import * as AuthActions from './auth.actions';
import { LoginResponse } from '../../core/services/auth.service';
import { SetOrganizationResponse } from '../../core/models/organization.model';

@Injectable()
export class AuthEffects {
  private actions$ = inject(Actions);
  private http = inject(HttpClient);
  private router = inject(Router);

  login$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.login),
      exhaustMap(({ credentials }) => {
        // Handle demo mode
        if (
          this.isDemoMode() &&
          credentials.username === 'demo' &&
          credentials.password === 'demo123'
        ) {
          const demoResponse = this.createDemoResponse();
          return of(AuthActions.loginSuccess({ response: demoResponse }));
        } else if (this.isDemoMode()) {
          return of(AuthActions.loginFailure({ error: { error: 'Invalid demo credentials' } }));
        }

        // Production API call
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return this.http.post<any>(`${environment.apiUrl}/auth/login`, credentials).pipe(
          map((response) => AuthActions.loginSuccess({ response })),
          catchError((error) => of(AuthActions.loginFailure({ error })))
        );
      })
    )
  );

  loginSuccess$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.loginSuccess),
        tap(({ response }) => {
          // Store auth data in localStorage
          localStorage.setItem('tenantly_token', response.token);
          localStorage.setItem('tenantly_refresh_token', response.refresh_token);
          localStorage.setItem('tenantly_user', JSON.stringify(response.user));
          localStorage.setItem(
            'tenantly_expires_at',
            typeof response.expires_at === 'string'
              ? response.expires_at
              : new Date(response.expires_at).toISOString()
          );

          // Persist organizations array for post-refresh restoration
          const orgs = response.organizations || [];
          if (orgs.length > 0) {
            localStorage.setItem('tenantly_organizations', JSON.stringify(orgs));
          }

          // If user belongs to multiple organizations without a default, show org picker
          const hasDefault = !!response.default_organization_id;
          if (orgs.length > 1 && !hasDefault) {
            this.router.navigate(['/select-organization']);
          } else {
            // Store current org context using organization_id (actual org ID)
            if (orgs.length === 1) {
              localStorage.setItem('tenantly_current_org_id', orgs[0].organization_id.toString());
            } else if (hasDefault) {
              localStorage.setItem('tenantly_current_org_id', response.default_organization_id!.toString());
            }
            this.router.navigate(['/dashboard']);
          }
        })
      ),
    { dispatch: false }
  );

  switchOrganization$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.switchOrganization),
      exhaustMap(({ organizationId }) =>
        this.http
          .post<SetOrganizationResponse>(`${environment.apiUrl}/auth/set-organization`, {
            organization_id: organizationId,
          })
          .pipe(
            map((response) => AuthActions.switchOrganizationSuccess({ response })),
            catchError((error) => of(AuthActions.switchOrganizationFailure({ error })))
          )
      )
    )
  );

  switchOrganizationSuccess$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.switchOrganizationSuccess),
        tap(({ response }) => {
          // Update tokens in localStorage
          localStorage.setItem('tenantly_token', response.token);
          localStorage.setItem('tenantly_refresh_token', response.refresh_token);
          localStorage.setItem(
            'tenantly_expires_at',
            typeof response.expires_at === 'string'
              ? response.expires_at
              : new Date(response.expires_at).toISOString()
          );
          localStorage.setItem(
            'tenantly_current_org_id',
            response.organization.organization_id.toString()
          );
          localStorage.setItem('tenantly_current_org', JSON.stringify(response.organization));

          // Navigate to dashboard after org switch
          this.router.navigate(['/dashboard']);
        })
      ),
    { dispatch: false }
  );

  logout$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.logout),
      exhaustMap(() => {
        // Handle demo mode
        if (this.isDemoMode()) {
          return of(AuthActions.logoutSuccess());
        }

        // Production API call
        return this.http.post(`${environment.apiUrl}/auth/logout`, {}).pipe(
          map(() => AuthActions.logoutSuccess()),
          catchError((error) => of(AuthActions.logoutFailure({ error })))
        );
      })
    )
  );

  logoutSuccess$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.logoutSuccess),
        tap(() => {
          // Clear auth data from localStorage
          localStorage.removeItem('tenantly_token');
          localStorage.removeItem('tenantly_refresh_token');
          localStorage.removeItem('tenantly_user');
          localStorage.removeItem('tenantly_expires_at');
          localStorage.removeItem('tenantly_organizations');
          localStorage.removeItem('tenantly_current_org_id');
          localStorage.removeItem('tenantly_current_org');

          // Navigate to login
          this.router.navigate(['/login']);
        })
      ),
    { dispatch: false }
  );

  refreshToken$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.refreshToken),
      exhaustMap(() => {
        const refreshToken = localStorage.getItem('tenantly_refresh_token');
        if (!refreshToken) {
          return of(AuthActions.refreshTokenFailure({ error: 'No refresh token available' }));
        }

        // Handle demo mode
        if (this.isDemoMode()) {
          const demoResponse = this.createDemoResponse('refreshed');
          return of(AuthActions.refreshTokenSuccess({ response: demoResponse }));
        }

        // Production API call
        const request = { refresh_token: refreshToken };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return this.http.post<any>(`${environment.apiUrl}/auth/refresh`, request).pipe(
          map((response) => AuthActions.refreshTokenSuccess({ response })),
          catchError((error) => of(AuthActions.refreshTokenFailure({ error })))
        );
      })
    )
  );

  refreshTokenSuccess$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.refreshTokenSuccess),
        tap(({ response }) => {
          // Update auth data in localStorage
          localStorage.setItem('tenantly_token', response.token);
          localStorage.setItem('tenantly_refresh_token', response.refresh_token);
          localStorage.setItem('tenantly_user', JSON.stringify(response.user));
          localStorage.setItem(
            'tenantly_expires_at',
            typeof response.expires_at === 'string'
              ? response.expires_at
              : new Date(response.expires_at).toISOString()
          );
        })
      ),
    { dispatch: false }
  );

  refreshTokenFailure$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.refreshTokenFailure),
        tap(() => {
          // Clear auth data and redirect to login
          localStorage.removeItem('tenantly_token');
          localStorage.removeItem('tenantly_refresh_token');
          localStorage.removeItem('tenantly_user');
          localStorage.removeItem('tenantly_expires_at');

          this.router.navigate(['/login']);
        })
      ),
    { dispatch: false }
  );

  changePassword$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.changePassword),
      exhaustMap(({ request }) => {
        // Handle demo mode
        if (this.isDemoMode()) {
          return of(
            AuthActions.changePasswordSuccess({ message: 'Password changed successfully' })
          );
        }

        // Production API call
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return this.http.post<any>(`${environment.apiUrl}/auth/change-password`, request).pipe(
          map((response) => AuthActions.changePasswordSuccess({ message: response.message })),
          catchError((error) => of(AuthActions.changePasswordFailure({ error })))
        );
      })
    )
  );

  resetPassword$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.resetPassword),
      exhaustMap(({ request }) => {
        // Handle demo mode
        if (this.isDemoMode()) {
          return of(
            AuthActions.resetPasswordSuccess({
              message: 'If the email exists, a password reset link has been sent',
            })
          );
        }

        // Production API call
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return this.http.post<any>(`${environment.apiUrl}/auth/reset-password`, request).pipe(
          map((response) => AuthActions.resetPasswordSuccess({ message: response.message })),
          catchError((error) => of(AuthActions.resetPasswordFailure({ error })))
        );
      })
    )
  );

  // Helper methods
  private isDemoMode(): boolean {
    return false; // Set to false for production
  }

  private createDemoResponse(suffix = ''): LoginResponse {
    const user = localStorage.getItem('tenantly_user');
    const existingUser = user ? JSON.parse(user) : null;

    return {
      token: `demo-token${suffix ? '-' + suffix : ''}`,
      refresh_token: `demo-refresh-token${suffix ? '-' + suffix : ''}`,
      user: existingUser || {
        id: 1,
        username: 'demo',
        email: 'demo@tenantly.com',
        role: 'Admin',
        active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      expires_at: new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString(), // 8 hours
    };
  }
}
