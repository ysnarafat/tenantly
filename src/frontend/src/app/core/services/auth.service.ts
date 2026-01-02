import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { tap, map, take } from 'rxjs/operators';
import { Store } from '@ngrx/store';
import { environment } from '../../../environments/environment';
import { AppState } from '../../store';
import * as AuthSelectors from '../../store/auth/auth.selectors';
import * as AuthActions from '../../store/auth/auth.actions';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface LoginResponse {
  token: string;
  refresh_token: string;
  user: User;
  expires_at: string | Date;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface ResetPasswordRequest {
  email: string;
}

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private http = inject(HttpClient);
  private store = inject(Store<AppState>);
  private readonly TOKEN_KEY = 'tenantly_token';
  private readonly REFRESH_TOKEN_KEY = 'tenantly_refresh_token';
  private readonly USER_KEY = 'tenantly_user';
  private readonly EXPIRES_AT_KEY = 'tenantly_expires_at';

  // DEMO MODE: Set to true to enable demo login (disable for production)
  private readonly DEMO_MODE = false; // <-- Set to false to disable demo login

  // NgRx store selectors for reactive access
  public isAuthenticated$ = this.store.select(AuthSelectors.selectIsAuthenticated);
  public user$ = this.store.select(AuthSelectors.selectUser);
  public loading$ = this.store.select(AuthSelectors.selectAuthLoading);
  public error$ = this.store.select(AuthSelectors.selectAuthError);
  public userRole$ = this.store.select(AuthSelectors.selectUserRole);

  login(credentials: LoginRequest): void {
    // Dispatch login action to NgRx store
    this.store.dispatch(AuthActions.login({ credentials }));
  }

  logout(): void {
    // Dispatch logout action to NgRx store
    this.store.dispatch(AuthActions.logout());
  }

  refreshToken(): void {
    // Dispatch refresh token action to NgRx store
    this.store.dispatch(AuthActions.refreshToken());
  }

  changePassword(request: ChangePasswordRequest): void {
    // Dispatch change password action to NgRx store
    this.store.dispatch(AuthActions.changePassword({ request }));
  }

  resetPassword(request: ResetPasswordRequest): void {
    // Dispatch reset password action to NgRx store
    this.store.dispatch(AuthActions.resetPassword({ request }));
  }

  clearError(): void {
    // Dispatch clear error action to NgRx store
    this.store.dispatch(AuthActions.clearError());
  }

  initializeAuth(): void {
    // Dispatch initialize auth action to NgRx store
    this.store.dispatch(AuthActions.initializeAuth());
  }

  getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  getRefreshToken(): string | null {
    return localStorage.getItem(this.REFRESH_TOKEN_KEY);
  }

  getUser(): User | null {
    // Synchronous access to user from NgRx store
    let user: User | null = null;
    this.store
      .select(AuthSelectors.selectUser)
      .pipe(take(1))
      .subscribe((u) => (user = u));
    return user;
  }

  getUserRole(): string {
    // Synchronous access to user role from NgRx store
    let role = '';
    this.store
      .select(AuthSelectors.selectUserRole)
      .pipe(take(1))
      .subscribe((r) => (role = r));
    return role;
  }

  isAuthenticated(): boolean {
    // Synchronous access to authentication status from NgRx store
    let isAuth = false;
    this.store
      .select(AuthSelectors.selectIsAuthenticated)
      .pipe(take(1))
      .subscribe((auth) => (isAuth = auth));
    return isAuth;
  }

  isTokenExpired(): boolean {
    const expiresAt = localStorage.getItem(this.EXPIRES_AT_KEY);
    if (!expiresAt) return true;

    return new Date() >= new Date(expiresAt);
  }

  hasRole(role: string): boolean {
    return this.getUserRole() === role;
  }

  hasAnyRole(roles: string[]): boolean {
    const userRole = this.getUserRole();
    return roles.includes(userRole);
  }

  isAdmin(): boolean {
    let isAdmin = false;
    this.store
      .select(AuthSelectors.selectIsAdmin)
      .pipe(take(1))
      .subscribe((admin) => (isAdmin = admin));
    return isAdmin;
  }

  isPropertyManager(): boolean {
    let isPM = false;
    this.store
      .select(AuthSelectors.selectIsPropertyManager)
      .pipe(take(1))
      .subscribe((pm) => (isPM = pm));
    return isPM;
  }

  isAccountant(): boolean {
    let isAccountant = false;
    this.store
      .select(AuthSelectors.selectIsAccountant)
      .pipe(take(1))
      .subscribe((acc) => (isAccountant = acc));
    return isAccountant;
  }

  // Reactive versions for components
  hasRole$(role: string): Observable<boolean> {
    return this.store.select(AuthSelectors.selectHasRole(role));
  }

  hasAnyRole$(roles: string[]): Observable<boolean> {
    return this.store.select(AuthSelectors.selectHasAnyRole(roles));
  }

  isAdmin$(): Observable<boolean> {
    return this.store.select(AuthSelectors.selectIsAdmin);
  }

  isPropertyManager$(): Observable<boolean> {
    return this.store.select(AuthSelectors.selectIsPropertyManager);
  }

  isAccountant$(): Observable<boolean> {
    return this.store.select(AuthSelectors.selectIsAccountant);
  }

  private hasToken(): boolean {
    return !!localStorage.getItem(this.TOKEN_KEY);
  }

  private storeAuthData(response: LoginResponse): void {
    localStorage.setItem(this.TOKEN_KEY, response.token);
    localStorage.setItem(this.REFRESH_TOKEN_KEY, response.refresh_token);
    localStorage.setItem(this.USER_KEY, JSON.stringify(response.user));
    localStorage.setItem(
      this.EXPIRES_AT_KEY,
      typeof response.expires_at === 'string'
        ? response.expires_at
        : new Date(response.expires_at).toISOString()
    );
    // Note: NgRx effects will handle state updates
  }

  private clearAuthData(): void {
    localStorage.removeItem(this.TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_TOKEN_KEY);
    localStorage.removeItem(this.USER_KEY);
    localStorage.removeItem(this.EXPIRES_AT_KEY);
    // Note: NgRx effects will handle state updates
  }
}
