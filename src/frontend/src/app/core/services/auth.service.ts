import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { tap, map, take } from 'rxjs/operators';
import { Store } from '@ngrx/store';
import { environment } from '../../../environments/environment';
import { AppState } from '../../store';
import * as AuthSelectors from '../../store/auth/auth.selectors';

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
  expires_at: string;
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
  private readonly DEMO_MODE = true; // <-- Set to false to disable demo login

  // Backward compatibility - now uses NgRx store
  public isAuthenticated$ = this.store.select(AuthSelectors.selectIsAuthenticated);

  login(credentials: LoginRequest): Observable<LoginResponse> {
    if (this.DEMO_MODE) {
      // DEMO: Accept demo credentials (username: demo, password: demo123)
      if (credentials.username === 'demo' && credentials.password === 'demo123') {
        const demoResponse: LoginResponse = {
          token: 'demo-token',
          refresh_token: 'demo-refresh-token',
          user: {
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
        this.storeAuthData(demoResponse);
        // Return observable that emits the demo response
        return new Observable<LoginResponse>((observer) => {
          observer.next(demoResponse);
          observer.complete();
        });
      } else {
        // Simulate failed login
        return new Observable<LoginResponse>((observer) => {
          observer.error({ error: 'Invalid demo credentials' });
        });
      }
    }
    // PRODUCTION: Use real API
    return this.http.post<LoginResponse>(`${environment.apiUrl}/api/v1/auth/login`, credentials).pipe(
      tap((response) => {
        this.storeAuthData(response);
      })
    );
  }

  logout(): Observable<any> {
    if (this.DEMO_MODE) {
      this.clearAuthData();
      return new Observable((observer) => {
        observer.next({ message: 'Successfully logged out' });
        observer.complete();
      });
    }

    return this.http.post(`${environment.apiUrl}/api/v1/auth/logout`, {}).pipe(
      tap(() => {
        this.clearAuthData();
      })
    );
  }

  refreshToken(): Observable<LoginResponse> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    if (this.DEMO_MODE) {
      const demoResponse: LoginResponse = {
        token: 'demo-token-refreshed',
        refresh_token: 'demo-refresh-token-refreshed',
        user: this.getUser()!,
        expires_at: new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString(),
      };
      this.storeAuthData(demoResponse);
      return new Observable((observer) => {
        observer.next(demoResponse);
        observer.complete();
      });
    }

    const request: RefreshTokenRequest = { refresh_token: refreshToken };
    return this.http.post<LoginResponse>(`${environment.apiUrl}/api/v1/auth/refresh`, request).pipe(
      tap((response) => {
        this.storeAuthData(response);
      })
    );
  }

  changePassword(request: ChangePasswordRequest): Observable<any> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        observer.next({ message: 'Password changed successfully' });
        observer.complete();
      });
    }

    return this.http.post(`${environment.apiUrl}/api/v1/auth/change-password`, request);
  }

  resetPassword(request: ResetPasswordRequest): Observable<any> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        observer.next({ message: 'If the email exists, a password reset link has been sent' });
        observer.complete();
      });
    }

    return this.http.post(`${environment.apiUrl}/api/v1/auth/reset-password`, request);
  }

  getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  getRefreshToken(): string | null {
    return localStorage.getItem(this.REFRESH_TOKEN_KEY);
  }

  getUser(): User | null {
    // For backward compatibility, still read from localStorage
    // In the future, this should use NgRx store
    const user = localStorage.getItem(this.USER_KEY);
    return user ? JSON.parse(user) : null;
  }

  getUserRole(): string {
    // For backward compatibility, still read from localStorage
    // In the future, this should use NgRx store
    const user = this.getUser();
    return user?.role || '';
  }

  isAuthenticated(): boolean {
    return this.hasToken() && !this.isTokenExpired();
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
    return this.hasRole('Admin');
  }

  isPropertyManager(): boolean {
    return this.hasRole('PropertyManager');
  }

  isAccountant(): boolean {
    return this.hasRole('Accountant');
  }

  private hasToken(): boolean {
    return !!localStorage.getItem(this.TOKEN_KEY);
  }

  private storeAuthData(response: LoginResponse): void {
    localStorage.setItem(this.TOKEN_KEY, response.token);
    localStorage.setItem(this.REFRESH_TOKEN_KEY, response.refresh_token);
    localStorage.setItem(this.USER_KEY, JSON.stringify(response.user));
    localStorage.setItem(this.EXPIRES_AT_KEY, response.expires_at);
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
