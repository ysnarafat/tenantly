import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { Router } from '@angular/router';
import { provideMockActions } from '@ngrx/effects/testing';
import { Action } from '@ngrx/store';
import { Observable, of } from 'rxjs';
import { AuthEffects } from './auth.effects';
import * as AuthActions from './auth.actions';

describe('AuthEffects', () => {
  let effects: AuthEffects;
  let actions$: Observable<Action>;
  let httpMock: HttpTestingController;
  let router: jasmine.SpyObj<Router>;

  const mockUser = {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    role: 'Admin',
    active: true,
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2023-01-01T00:00:00Z',
  };

  const mockOrg1 = {
    id: 10,
    organization_id: 100,
    organization: { id: 100, name: 'Acme Corp', created_at: '', updated_at: '' },
    role: 'Admin',
    created_at: '',
    updated_at: '',
  };

  const mockOrg2 = {
    id: 11,
    organization_id: 101,
    organization: { id: 101, name: 'Beta Ltd', created_at: '', updated_at: '' },
    role: 'PropertyManager',
    created_at: '',
    updated_at: '',
  };

  const mockLoginResponse = {
    token: 'test-token',
    refresh_token: 'test-refresh-token',
    user: mockUser,
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
      spyOn<AuthEffects>(effects, 'isDemoMode' as unknown as keyof AuthEffects).and.returnValue(
        true
      );
      const credentials = { username: 'demo', password: 'demo123' };
      const action = AuthActions.login({ credentials });

      actions$ = of(action);

      effects.login$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.loginSuccess.type);
        expect(
          (result as unknown as { response: { user: { username: string } } }).response.user.username
        ).toBe('demo');
        done();
      });
    });

    it('should return loginFailure action on invalid demo credentials', (done) => {
      spyOn<AuthEffects>(effects, 'isDemoMode' as unknown as keyof AuthEffects).and.returnValue(
        true
      );
      const credentials = { username: 'invalid', password: 'invalid' };
      const action = AuthActions.login({ credentials });

      actions$ = of(action);

      effects.login$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.loginFailure.type);
        expect((result as unknown as { error: { error: string } }).error.error).toBe(
          'Invalid demo credentials'
        );
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

    it('should navigate to /select-organization when user has multiple orgs and no default', (done) => {
      const response = { ...mockLoginResponse, organizations: [mockOrg1, mockOrg2] };
      actions$ = of(AuthActions.loginSuccess({ response }));

      effects.loginSuccess$.subscribe(() => {
        expect(router.navigate).toHaveBeenCalledWith(['/select-organization']);
        done();
      });
    });

    it('should navigate to /dashboard when user has a single org', (done) => {
      const response = { ...mockLoginResponse, organizations: [mockOrg1] };
      actions$ = of(AuthActions.loginSuccess({ response }));

      effects.loginSuccess$.subscribe(() => {
        expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
        done();
      });
    });

    it('should navigate to /dashboard when a default_organization_id is set', (done) => {
      const response = {
        ...mockLoginResponse,
        organizations: [mockOrg1, mockOrg2],
        default_organization_id: 100,
      };
      actions$ = of(AuthActions.loginSuccess({ response }));

      effects.loginSuccess$.subscribe(() => {
        expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
        done();
      });
    });

    it('should persist organizations array to localStorage', (done) => {
      const response = { ...mockLoginResponse, organizations: [mockOrg1, mockOrg2] };
      actions$ = of(AuthActions.loginSuccess({ response }));

      effects.loginSuccess$.subscribe(() => {
        const stored = JSON.parse(localStorage.getItem('tenantly_organizations')!);
        expect(stored).toEqual([mockOrg1, mockOrg2]);
        done();
      });
    });

    it('should store organization_id to localStorage for a single org', (done) => {
      const response = { ...mockLoginResponse, organizations: [mockOrg1] };
      actions$ = of(AuthActions.loginSuccess({ response }));

      effects.loginSuccess$.subscribe(() => {
        expect(localStorage.getItem('tenantly_current_org_id')).toBe('100');
        done();
      });
    });

    it('should store default_organization_id to localStorage when provided', (done) => {
      const response = {
        ...mockLoginResponse,
        organizations: [mockOrg1, mockOrg2],
        default_organization_id: 101,
      };
      actions$ = of(AuthActions.loginSuccess({ response }));

      effects.loginSuccess$.subscribe(() => {
        expect(localStorage.getItem('tenantly_current_org_id')).toBe('101');
        done();
      });
    });
  });

  describe('switchOrganization$', () => {
    it('should call the set-organization API and dispatch switchOrganizationSuccess', (done) => {
      const mockSwitchResponse = {
        token: 'new-token',
        refresh_token: 'new-refresh',
        organization: mockOrg1,
        expires_at: '2099-01-01T00:00:00Z',
      };
      actions$ = of(AuthActions.switchOrganization({ organizationId: 100 }));

      effects.switchOrganization$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.switchOrganizationSuccess.type);
        expect((result as unknown as { response: typeof mockSwitchResponse }).response).toEqual(
          mockSwitchResponse
        );
        done();
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/auth/set-organization`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ organization_id: 100 });
      req.flush(mockSwitchResponse);
    });

    it('should dispatch switchOrganizationFailure on API error', (done) => {
      actions$ = of(AuthActions.switchOrganization({ organizationId: 100 }));

      effects.switchOrganization$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.switchOrganizationFailure.type);
        done();
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/auth/set-organization`);
      req.flush({ error: 'Forbidden' }, { status: 403, statusText: 'Forbidden' });
    });
  });

  describe('switchOrganizationSuccess$', () => {
    it('should update tokens in localStorage and navigate to dashboard', (done) => {
      const response = {
        token: 'new-token',
        refresh_token: 'new-refresh',
        organization: mockOrg1,
        expires_at: '2099-01-01T00:00:00Z',
      };
      actions$ = of(AuthActions.switchOrganizationSuccess({ response }));

      effects.switchOrganizationSuccess$.subscribe(() => {
        expect(localStorage.getItem('tenantly_token')).toBe('new-token');
        expect(localStorage.getItem('tenantly_refresh_token')).toBe('new-refresh');
        expect(localStorage.getItem('tenantly_current_org_id')).toBe('100');
        expect(localStorage.getItem('tenantly_current_org')).toBe(JSON.stringify(mockOrg1));
        expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
        done();
      });
    });
  });

  describe('logout$', () => {
    it('should return logoutSuccess action in demo mode', (done) => {
      spyOn<AuthEffects>(effects, 'isDemoMode' as unknown as keyof AuthEffects).and.returnValue(
        true
      );
      const action = AuthActions.logout();
      actions$ = of(action);

      effects.logout$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.logoutSuccess.type);
        done();
      });
    });
  });

  describe('logoutSuccess$', () => {
    it('should clear auth and org data from localStorage and navigate to login', (done) => {
      localStorage.setItem('tenantly_token', 'test-token');
      localStorage.setItem('tenantly_refresh_token', 'test-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockLoginResponse.user));
      localStorage.setItem('tenantly_expires_at', 'test-expires-at');
      localStorage.setItem('tenantly_organizations', JSON.stringify([mockOrg1]));
      localStorage.setItem('tenantly_current_org_id', '100');
      localStorage.setItem('tenantly_current_org', JSON.stringify(mockOrg1));

      actions$ = of(AuthActions.logoutSuccess());

      effects.logoutSuccess$.subscribe(() => {
        expect(localStorage.getItem('tenantly_token')).toBeNull();
        expect(localStorage.getItem('tenantly_refresh_token')).toBeNull();
        expect(localStorage.getItem('tenantly_user')).toBeNull();
        expect(localStorage.getItem('tenantly_expires_at')).toBeNull();
        expect(localStorage.getItem('tenantly_organizations')).toBeNull();
        expect(localStorage.getItem('tenantly_current_org_id')).toBeNull();
        expect(localStorage.getItem('tenantly_current_org')).toBeNull();
        expect(router.navigate).toHaveBeenCalledWith(['/login']);
        done();
      });
    });
  });

  describe('refreshToken$', () => {
    it('should return refreshTokenSuccess action in demo mode', (done) => {
      spyOn<AuthEffects>(effects, 'isDemoMode' as unknown as keyof AuthEffects).and.returnValue(
        true
      );
      localStorage.setItem('tenantly_refresh_token', 'test-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockLoginResponse.user));

      const action = AuthActions.refreshToken();
      actions$ = of(action);

      effects.refreshToken$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.refreshTokenSuccess.type);
        expect(
          (result as unknown as { response: { user: typeof mockLoginResponse.user } }).response.user
        ).toEqual(mockLoginResponse.user);
        done();
      });
    });

    it('should return refreshTokenFailure action when no refresh token', (done) => {
      const action = AuthActions.refreshToken();
      actions$ = of(action);

      effects.refreshToken$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.refreshTokenFailure.type);
        expect((result as unknown as { error: string }).error).toBe('No refresh token available');
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
        expect(localStorage.getItem('tenantly_refresh_token')).toBe(
          mockLoginResponse.refresh_token
        );
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
      spyOn<AuthEffects>(effects, 'isDemoMode' as unknown as keyof AuthEffects).and.returnValue(
        true
      );
      const request = { current_password: 'old', new_password: 'new' };
      const action = AuthActions.changePassword({ request });
      actions$ = of(action);

      effects.changePassword$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.changePasswordSuccess.type);
        expect((result as unknown as { message: string }).message).toBe(
          'Password changed successfully'
        );
        done();
      });
    });
  });

  describe('resetPassword$', () => {
    it('should return resetPasswordSuccess action in demo mode', (done) => {
      spyOn<AuthEffects>(effects, 'isDemoMode' as unknown as keyof AuthEffects).and.returnValue(
        true
      );
      const request = { email: 'test@example.com' };
      const action = AuthActions.resetPassword({ request });
      actions$ = of(action);

      effects.resetPassword$.subscribe((result) => {
        expect(result.type).toBe(AuthActions.resetPasswordSuccess.type);
        expect((result as unknown as { message: string }).message).toBe(
          'If the email exists, a password reset link has been sent'
        );
        done();
      });
    });
  });
});
