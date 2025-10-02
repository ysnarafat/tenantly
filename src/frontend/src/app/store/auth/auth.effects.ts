import { Injectable, inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Router } from '@angular/router';
import { of } from 'rxjs';
import { map, exhaustMap, catchError, tap } from 'rxjs/operators';
import { AuthService } from '../../core/services/auth.service';
import * as AuthActions from './auth.actions';

@Injectable()
export class AuthEffects {
  private actions$ = inject(Actions);
  private authService = inject(AuthService);
  private router = inject(Router);

  login$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.login),
      exhaustMap(({ credentials }) =>
        this.authService.login(credentials).pipe(
          map((response) => AuthActions.loginSuccess({ response })),
          catchError((error) => of(AuthActions.loginFailure({ error })))
        )
      )
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
          localStorage.setItem('tenantly_expires_at', response.expires_at);
          
          // Navigate to dashboard
          this.router.navigate(['/dashboard']);
        })
      ),
    { dispatch: false }
  );

  logout$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.logout),
      exhaustMap(() =>
        this.authService.logout().pipe(
          map(() => AuthActions.logoutSuccess()),
          catchError((error) => of(AuthActions.logoutFailure({ error })))
        )
      )
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
        this.authService.refreshToken().pipe(
          map((response) => AuthActions.refreshTokenSuccess({ response })),
          catchError((error) => of(AuthActions.refreshTokenFailure({ error })))
        )
      )
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
          localStorage.setItem('tenantly_expires_at', response.expires_at);
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
      exhaustMap(({ request }) =>
        this.authService.changePassword(request).pipe(
          map((response) => AuthActions.changePasswordSuccess({ message: response.message })),
          catchError((error) => of(AuthActions.changePasswordFailure({ error })))
        )
      )
    )
  );

  resetPassword$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.resetPassword),
      exhaustMap(({ request }) =>
        this.authService.resetPassword(request).pipe(
          map((response) => AuthActions.resetPasswordSuccess({ message: response.message })),
          catchError((error) => of(AuthActions.resetPasswordFailure({ error })))
        )
      )
    )
  );
}