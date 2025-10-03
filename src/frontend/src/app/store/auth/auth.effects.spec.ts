import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { Router } from '@angular/router';
import { provideMockActions } from '@ngrx/effects/testing';
import { Action } from '@ngrx/store';
import { Observable, of } from 'rxjs';
import { AuthEffects } from './auth.effects';
import * as AuthActions from './auth.actions';
import { environment } from '../../../environments/environment';

describe('AuthEffects', () => {
  let effects: AuthEffects;
  let actions$: Observable<Action>;
  let httpMock: HttpTestingController;
  let router: jasmine.SpyObj<Router>;

  const mockLoginResponse = {
    token: 'test-token',
    refresh_token: 'test-refresh-token',
    user: {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      role: 'Admin',
      active: true,
      created_at: '2023-01-01T00:00:00Z',
      updated_at: '2023-01-01T00:00:00Z',
    },
    expires_at: '2023-12-31T23:59:59Z',
  };

  beforeEach(() => {
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        AuthEffects,
        provideMockActions(() => actions$),
        { provide: Router, useValue: routerSpy },
      ],
    });

    effects = TestBed.inject(AuthEffects);
    httpMock = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;

    // Clear localStorage before each test
    localStorage.clear();
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  describe('login$', () => {
    it('should return loginSuccess action on successful demo login', (done) => {
      const credentials = { username: 'demo', password: 'demo123' };
      const action = AuthActions.login({ credentials });

      actions$ = of(action);

      effects.login$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.loginSuccess.type);
        expect((result as any).response.user.username).toBe('demo');
        done();
      });
    });

    it('should return loginFailure action on invalid demo credentials', (done) => {
      const credentials = { username: 'invalid', password: 'invalid' };
      const action = AuthActions.login({ credentials });

      actions$ = of(action);

      effects.login$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.loginFailure.type);
        expect((result as any).error.error).toBe('Invalid demo credentials');
        done();
      });
    });
  });

  describe('loginSuccess$', () => {
    it('should store auth data in localStorage and navigate to dashboard', (done) => {
      const action = AuthActions.loginSuccess({ response: mockLoginResponse });
      actions$ = of(action);

      effects.loginSuccess$.subscribe(() => {
        expect(localStorage.getItem('tenantly_token')).toBe(mockLoginResponse.token);
        expect(localStorage.getItem('tenantly_refresh_token')).toBe(mockLoginResponse.refresh_token);
        expect(localStorage.getItem('tenantly_user')).toBe(JSON.stringify(mockLoginResponse.user));
        expect(localStorage.getItem('tenantly_expires_at')).toBe(mockLoginResponse.expires_at);
        expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
        done();
      });
    });
  });

  describe('logout$', () => {
    it('should return logoutSuccess action in demo mode', (done) => {
      const action = AuthActions.logout();
      actions$ = of(action);

      effects.logout$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.logoutSuccess.type);
        done();
      });
    });
  });

  describe('logoutSuccess$', () => {
    it('should clear auth data from localStorage and navigate to login', (done) => {
      // Set up localStorage with auth data
      localStorage.setItem('tenantly_token', 'test-token');
      localStorage.setItem('tenantly_refresh_token', 'test-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockLoginResponse.user));
      localStorage.setItem('tenantly_expires_at', 'test-expires-at');

      const action = AuthActions.logoutSuccess();
      actions$ = of(action);

      effects.logoutSuccess$.subscribe(() => {
        expect(localStorage.getItem('tenantly_token')).toBeNull();
        expect(localStorage.getItem('tenantly_refresh_token')).toBeNull();
        expect(localStorage.getItem('tenantly_user')).toBeNull();
        expect(localStorage.getItem('tenantly_expires_at')).toBeNull();
        expect(router.navigate).toHaveBeenCalledWith(['/login']);
        done();
      });
    });
  });

  describe('refreshToken$', () => {
    it('should return refreshTokenSuccess action in demo mode', (done) => {
      localStorage.setItem('tenantly_refresh_token', 'test-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockLoginResponse.user));

      const action = AuthActions.refreshToken();
      actions$ = of(action);

      effects.refreshToken$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.refreshTokenSuccess.type);
        expect((result as any).response.user).toEqual(mockLoginResponse.user);
        done();
      });
    });

    it('should return refreshTokenFailure action when no refresh token', (done) => {
      const action = AuthActions.refreshToken();
      actions$ = of(action);

      effects.refreshToken$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.refreshTokenFailure.type);
        expect((result as any).error).toBe('No refresh token available');
        done();
      });
    });
  });

  describe('refreshTokenSuccess$', () => {
    it('should update auth data in localStorage', (done) => {
      const action = AuthActions.refreshTokenSuccess({ response: mockLoginResponse });
      actions$ = of(action);

      effects.refreshTokenSuccess$.subscribe(() => {
        expect(localStorage.getItem('tenantly_token')).toBe(mockLoginResponse.token);
        expect(localStorage.getItem('tenantly_refresh_token')).toBe(mockLoginResponse.refresh_token);
        expect(localStorage.getItem('tenantly_user')).toBe(JSON.stringify(mockLoginResponse.user));
        expect(localStorage.getItem('tenantly_expires_at')).toBe(mockLoginResponse.expires_at);
        done();
      });
    });
  });

  describe('refreshTokenFailure$', () => {
    it('should clear auth data and navigate to login', (done) => {
      // Set up localStorage with auth data
      localStorage.setItem('tenantly_token', 'test-token');
      localStorage.setItem('tenantly_refresh_token', 'test-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockLoginResponse.user));
      localStorage.setItem('tenantly_expires_at', 'test-expires-at');

      const action = AuthActions.refreshTokenFailure({ error: 'Token expired' });
      actions$ = of(action);

      effects.refreshTokenFailure$.subscribe(() => {
        expect(localStorage.getItem('tenantly_token')).toBeNull();
        expect(localStorage.getItem('tenantly_refresh_token')).toBeNull();
        expect(localStorage.getItem('tenantly_user')).toBeNull();
        expect(localStorage.getItem('tenantly_expires_at')).toBeNull();
        expect(router.navigate).toHaveBeenCalledWith(['/login']);
        done();
      });
    });
  });

  describe('changePassword$', () => {
    it('should return changePasswordSuccess action in demo mode', (done) => {
      const request = { current_password: 'old', new_password: 'new' };
      const action = AuthActions.changePassword({ request });
      actions$ = of(action);

      effects.changePassword$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.changePasswordSuccess.type);
        expect((result as any).message).toBe('Password changed successfully');
        done();
      });
    });
  });

  describe('resetPassword$', () => {
    it('should return resetPasswordSuccess action in demo mode', (done) => {
      const request = { email: 'test@example.com' };
      const action = AuthActions.resetPassword({ request });
      actions$ = of(action);

      effects.resetPassword$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.resetPasswordSuccess.type);
        expect((result as any).message).toBe('If the email exists, a password reset link has been sent');
        done();
      });
    });
  });
});