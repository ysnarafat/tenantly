import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject } from 'rxjs';
import { tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface User {
  id: string;
  username: string;
  role: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private http = inject(HttpClient);
  private readonly TOKEN_KEY = 'tenantly_token';
  private readonly USER_KEY = 'tenantly_user';

  // DEMO MODE: Set to true to enable demo login (disable for production)
  private readonly DEMO_MODE = true; // <-- Set to false to disable demo login

  private isAuthenticatedSubject = new BehaviorSubject<boolean>(this.hasToken());
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();

  login(credentials: LoginRequest): Observable<LoginResponse> {
    if (this.DEMO_MODE) {
      // DEMO: Accept demo credentials (username: demo, password: demo123)
      if (credentials.username === 'demo' && credentials.password === 'demo123') {
        const demoResponse: LoginResponse = {
          token: 'demo-token',
          user: {
            id: '1',
            username: 'demo',
            role: 'Admin',
          },
        };
        localStorage.setItem(this.TOKEN_KEY, demoResponse.token);
        localStorage.setItem(this.USER_KEY, JSON.stringify(demoResponse.user));
        this.isAuthenticatedSubject.next(true);
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
    return this.http.post<LoginResponse>(`${environment.apiUrl}/auth/login`, credentials).pipe(
      tap((response) => {
        localStorage.setItem(this.TOKEN_KEY, response.token);
        localStorage.setItem(this.USER_KEY, JSON.stringify(response.user));
        this.isAuthenticatedSubject.next(true);
      })
    );
  }

  logout(): void {
    localStorage.removeItem(this.TOKEN_KEY);
    localStorage.removeItem(this.USER_KEY);
    this.isAuthenticatedSubject.next(false);
  }

  getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  getUser(): User | null {
    const user = localStorage.getItem(this.USER_KEY);
    return user ? JSON.parse(user) : null;
  }

  getUserRole(): string {
    const user = this.getUser();
    return user?.role || '';
  }

  isAuthenticated(): boolean {
    return this.hasToken();
  }

  private hasToken(): boolean {
    return !!localStorage.getItem(this.TOKEN_KEY);
  }
}
