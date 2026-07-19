import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Router } from '@angular/router';
import { of } from 'rxjs';
import { map, exhaustMap, catchError, tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import * as AuthActions from './auth.actions';
import { SetOrganizationResponse } from '../../core/models/organization.model';

function toSerializableError(error: HttpErrorResponse): { status: number; message: string } {
  return {
    status: error.status,
    message: error.error?.error ?? error.message ?? 'Request failed',
  };
}

@Injectable()
export class AuthEffects {
  private actions$ = inject(Actions);
  private http = inject(HttpClient);
  private router = inject(Router);

  login$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.login),
      exhaustMap(({ credentials }) =>
        // withCredentials: the backend sets the refresh token as an httpOnly cookie on this response.
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        this.http.post<any>(`${environment.apiUrl}/auth/login`, credentials, { withCredentials: true }).pipe(
          map((response) => AuthActions.loginSuccess({ response })),
          catchError((error: HttpErrorResponse) =>
            of(AuthActions.loginFailure({ error: toSerializableError(error) }))
          )
        )
      )
    )
  );

  loginSuccess$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.loginSuccess),
        tap(({ response }) => {
          // Store auth data in localStorage. The refresh token is NOT stored
          // here — the backend sets it as an httpOnly cookie the browser
          // manages on its own; it's never present in this JSON response.
          localStorage.setItem('tenantly_token', response.token);
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
              localStorage.setItem(
                'tenantly_current_org_id',
                response.default_organization_id!.toString()
              );
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
          .post<SetOrganizationResponse>(
            `${environment.apiUrl}/auth/set-organization`,
            { organization_id: organizationId },
            { withCredentials: true }
          )
          .pipe(
            map((response) => AuthActions.switchOrganizationSuccess({ response })),
            catchError((error: HttpErrorResponse) =>
              of(AuthActions.switchOrganizationFailure({ error: toSerializableError(error) }))
            )
          )
      )
    )
  );

  switchOrganizationSuccess$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.switchOrganizationSuccess),
        tap(({ response }) => {
          // Update tokens in localStorage (refresh token: see loginSuccess$ note above)
          localStorage.setItem('tenantly_token', response.token);
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
      exhaustMap(() =>
        // withCredentials: sends the httpOnly refresh cookie so the backend can clear it.
        this.http.post(`${environment.apiUrl}/auth/logout`, {}, { withCredentials: true }).pipe(
          map(() => AuthActions.logoutSuccess()),
          catchError((error: HttpErrorResponse) =>
            of(AuthActions.logoutFailure({ error: toSerializableError(error) }))
          )
        )
      )
    )
  );

  logoutSuccess$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.logoutSuccess),
        tap(() => {
          // Clear auth data from localStorage (the backend clears the
          // httpOnly refresh cookie itself as part of the logout response)
          localStorage.removeItem('tenantly_token');
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
      exhaustMap(() =>
        // No body needed — the httpOnly refresh cookie is sent automatically
        // via withCredentials; the browser holds it, not this code.
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        this.http.post<any>(`${environment.apiUrl}/auth/refresh`, {}, { withCredentials: true }).pipe(
          map((response) => AuthActions.refreshTokenSuccess({ response })),
          catchError((error: HttpErrorResponse) =>
            of(AuthActions.refreshTokenFailure({ error: toSerializableError(error) }))
          )
        )
      )
    )
  );

  refreshTokenSuccess$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.refreshTokenSuccess),
        tap(({ response }) => {
          // Update auth data in localStorage (refresh token: see loginSuccess$ note above)
          localStorage.setItem('tenantly_token', response.token);
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
      exhaustMap(({ request }) =>
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        this.http.post<any>(`${environment.apiUrl}/auth/change-password`, request).pipe(
          map((response) => AuthActions.changePasswordSuccess({ message: response.message })),
          catchError((error: HttpErrorResponse) =>
            of(AuthActions.changePasswordFailure({ error: toSerializableError(error) }))
          )
        )
      )
    )
  );

  resetPassword$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.resetPassword),
      exhaustMap(({ request }) =>
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        this.http.post<any>(`${environment.apiUrl}/auth/reset-password`, request).pipe(
          map((response) => AuthActions.resetPasswordSuccess({ message: response.message })),
          catchError((error: HttpErrorResponse) =>
            of(AuthActions.resetPasswordFailure({ error: toSerializableError(error) }))
          )
        )
      )
    )
  );
}
