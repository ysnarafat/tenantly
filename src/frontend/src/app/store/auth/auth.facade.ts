import { Injectable, inject } from '@angular/core';
import { Store } from '@ngrx/store';
import { Observable } from 'rxjs';
import { AppState } from '../index';
import { LoginRequest, User, ChangePasswordRequest, ResetPasswordRequest } from '../../core/services/auth.service';
import * as AuthActions from './auth.actions';
import * as AuthSelectors from './auth.selectors';

@Injectable({
  providedIn: 'root',
})
export class AuthFacade {
  private store = inject(Store<AppState>);

  // Selectors
  user$ = this.store.select(AuthSelectors.selectUser);
  token$ = this.store.select(AuthSelectors.selectToken);
  isAuthenticated$ = this.store.select(AuthSelectors.selectIsAuthenticated);
  loading$ = this.store.select(AuthSelectors.selectAuthLoading);
  error$ = this.store.select(AuthSelectors.selectAuthError);
  userRole$ = this.store.select(AuthSelectors.selectUserRole);
  isAdmin$ = this.store.select(AuthSelectors.selectIsAdmin);
  isPropertyManager$ = this.store.select(AuthSelectors.selectIsPropertyManager);
  isAccountant$ = this.store.select(AuthSelectors.selectIsAccountant);
  canViewLeases$ = this.store.select(AuthSelectors.selectCanViewLeases);
  canViewAttachments$ = this.store.select(AuthSelectors.selectCanViewAttachments);
  canManageUsers$ = this.store.select(AuthSelectors.selectCanManageUsers);
  isTokenExpired$ = this.store.select(AuthSelectors.selectIsTokenExpired);

  // Actions
  login(credentials: LoginRequest): void {
    this.store.dispatch(AuthActions.login({ credentials }));
  }

  logout(): void {
    this.store.dispatch(AuthActions.logout());
  }

  refreshToken(): void {
    this.store.dispatch(AuthActions.refreshToken());
  }

  changePassword(request: ChangePasswordRequest): void {
    this.store.dispatch(AuthActions.changePassword({ request }));
  }

  resetPassword(request: ResetPasswordRequest): void {
    this.store.dispatch(AuthActions.resetPassword({ request }));
  }

  initializeAuth(): void {
    this.store.dispatch(AuthActions.initializeAuth());
  }

  clearError(): void {
    this.store.dispatch(AuthActions.clearError());
  }

  setLoading(loading: boolean): void {
    this.store.dispatch(AuthActions.setLoading({ loading }));
  }

  // Synchronous getters for backward compatibility
  getCurrentUser(): Observable<User | null> {
    return this.user$;
  }

  getCurrentUserRole(): Observable<string> {
    return this.userRole$;
  }

  getIsAuthenticated(): Observable<boolean> {
    return this.isAuthenticated$;
  }

  // Helper methods for role checking
  hasRole(role: string): Observable<boolean> {
    return this.store.select(AuthSelectors.selectHasRole(role));
  }

  hasAnyRole(roles: string[]): Observable<boolean> {
    return this.store.select(AuthSelectors.selectHasAnyRole(roles));
  }
}