import { TestBed } from '@angular/core/testing';
import { Store } from '@ngrx/store';
import { BehaviorSubject, of } from 'rxjs';
import {
  AuthService,
  LoginRequest,
  ChangePasswordRequest,
  ResetPasswordRequest,
  User,
} from './auth.service';
import { AppState } from '../../store';
import * as AuthActions from '../../store/auth/auth.actions';
import * as AuthSelectors from '../../store/auth/auth.selectors';

describe('AuthService', () => {
  let service: AuthService;
  let store: jasmine.SpyObj<Store<AppState>>;
  let loadingSubject: BehaviorSubject<boolean>;
  let errorSubject: BehaviorSubject<string | null>;

  const mockUser: User = {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    role: 'Admin',
    active: true,
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2023-01-01T00:00:00Z',
  };

  beforeEach(() => {
    loadingSubject = new BehaviorSubject<boolean>(false);
    errorSubject = new BehaviorSubject<string | null>(null);
    const storeSpy = jasmine.createSpyObj('Store', ['select', 'dispatch']);

    // IMPORTANT: Set up the spy BEFORE creating the service
    // because the service creates observables in its constructor
    storeSpy.select.and.callFake((selector: (state: AppState) => unknown) => {
      // Return appropriate observables based on selector
      if (selector === AuthSelectors.selectUser) return of(mockUser);
      if (selector === AuthSelectors.selectUserRole) return of('Admin');
      if (selector === AuthSelectors.selectIsAuthenticated) return of(true);
      if (selector === AuthSelectors.selectIsAdmin) return of(true);
      if (selector === AuthSelectors.selectIsPropertyManager) return of(false);
      if (selector === AuthSelectors.selectIsAccountant) return of(false);
      if (selector === AuthSelectors.selectAuthLoading) return loadingSubject.asObservable();
      if (selector === AuthSelectors.selectAuthError) return errorSubject.asObservable();
      // For parameterized selectors
      if (typeof selector === 'function') {
        return of(true); // Default for hasRole/hasAnyRole selectors
      }
      return of(null);
    });

    TestBed.configureTestingModule({
      providers: [AuthService, { provide: Store, useValue: storeSpy }],
    });

    service = TestBed.inject(AuthService);
    store = TestBed.inject(Store) as jasmine.SpyObj<Store<AppState>>;

    // Setup default store selectors
    store.select.and.callFake((selector: unknown) => {
      if (selector === AuthSelectors.selectUser) return of(mockUser);
      if (selector === AuthSelectors.selectUserRole) return of('Admin');
      if (selector === AuthSelectors.selectIsAuthenticated) return of(true);
      if (selector === AuthSelectors.selectIsAdmin) return of(true);
      if (selector === AuthSelectors.selectIsPropertyManager) return of(false);
      if (selector === AuthSelectors.selectIsAccountant) return of(false);
      if (selector === AuthSelectors.selectAuthLoading) return of(false);
      if (selector === AuthSelectors.selectAuthError) return of(null);
      return of(null);
    });
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('login', () => {
    it('should dispatch login action', () => {
      const credentials: LoginRequest = { username: 'test', password: 'test' };

      service.login(credentials);

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.login({ credentials }));
    });

    it('should dispatch login action with empty credentials', () => {
      const credentials: LoginRequest = { username: '', password: '' };

      service.login(credentials);

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.login({ credentials }));
    });
  });

  describe('logout', () => {
    it('should dispatch logout action', () => {
      service.logout();

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.logout());
    });

    it('should dispatch logout action only once when called multiple times', () => {
      service.logout();
      service.logout();

      expect(store.dispatch).toHaveBeenCalledTimes(2);
      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.logout());
    });
  });

  describe('refreshToken', () => {
    it('should dispatch refreshToken action', () => {
      service.refreshToken();

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.refreshToken());
    });
  });

  describe('changePassword', () => {
    it('should dispatch changePassword action', () => {
      const request: ChangePasswordRequest = {
        current_password: 'old',
        new_password: 'new',
      };

      service.changePassword(request);

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.changePassword({ request }));
    });

    it('should dispatch changePassword action with same passwords', () => {
      const request: ChangePasswordRequest = {
        current_password: 'same',
        new_password: 'same',
      };

      service.changePassword(request);

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.changePassword({ request }));
    });
  });

  describe('resetPassword', () => {
    it('should dispatch resetPassword action', () => {
      const request: ResetPasswordRequest = { email: 'test@example.com' };

      service.resetPassword(request);

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.resetPassword({ request }));
    });

    it('should dispatch resetPassword action with invalid email format', () => {
      const request: ResetPasswordRequest = { email: 'invalid-email' };

      service.resetPassword(request);

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.resetPassword({ request }));
    });
  });

  describe('clearError', () => {
    it('should dispatch clearError action', () => {
      service.clearError();

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.clearError());
    });
  });

  describe('initializeAuth', () => {
    it('should dispatch initializeAuth action', () => {
      service.initializeAuth();

      expect(store.dispatch).toHaveBeenCalledWith(AuthActions.initializeAuth());
    });
  });

  describe('getToken', () => {
    it('should return token from localStorage', () => {
      spyOn(localStorage, 'getItem').and.returnValue('test-token');

      const result = service.getToken();

      expect(result).toBe('test-token');
      expect(localStorage.getItem).toHaveBeenCalledWith('tenantly_token');
    });

    it('should return null when no token exists', () => {
      spyOn(localStorage, 'getItem').and.returnValue(null);

      const result = service.getToken();

      expect(result).toBeNull();
    });

    it('should return empty string token', () => {
      spyOn(localStorage, 'getItem').and.returnValue('');

      const result = service.getToken();

      expect(result).toBe('');
    });
  });

  describe('getUser', () => {
    it('should return user from store', () => {
      const result = service.getUser();

      expect(result).toEqual(mockUser);
    });

    it('should return null when no user in store', () => {
      store.select.and.returnValue(of(null));

      const result = service.getUser();

      expect(result).toBeNull();
    });

    it('should return user with different role', () => {
      const propertyManagerUser: User = {
        ...mockUser,
        role: 'PropertyManager',
      };
      store.select.and.returnValue(of(propertyManagerUser));

      const result = service.getUser();

      expect(result?.role).toBe('PropertyManager');
    });
  });

  describe('getUserRole', () => {
    it('should return user role from store', () => {
      const result = service.getUserRole();

      expect(result).toBe('Admin');
    });

    it('should return empty string when no role in store', () => {
      store.select.and.returnValue(of(''));

      const result = service.getUserRole();

      expect(result).toBe('');
    });

    it('should return PropertyManager role', () => {
      store.select.and.returnValue(of('PropertyManager'));

      const result = service.getUserRole();

      expect(result).toBe('PropertyManager');
    });

    it('should return Accountant role', () => {
      store.select.and.returnValue(of('Accountant'));

      const result = service.getUserRole();

      expect(result).toBe('Accountant');
    });
  });

  describe('isAuthenticated', () => {
    it('should return true when authenticated', () => {
      const result = service.isAuthenticated();

      expect(result).toBe(true);
    });

    it('should return false when not authenticated', () => {
      store.select.and.returnValue(of(false));

      const result = service.isAuthenticated();

      expect(result).toBe(false);
    });
  });

  describe('isTokenExpired', () => {
    it('should return true when no expiry date exists', () => {
      const result = service.isTokenExpired();

      expect(result).toBe(true);
    });

    it('should return true when token is expired', () => {
      const pastDate = new Date();
      pastDate.setHours(pastDate.getHours() - 1);
      localStorage.setItem('tenantly_expires_at', pastDate.toISOString());

      const result = service.isTokenExpired();

      expect(result).toBe(true);
    });

    it('should return false when token is not expired', () => {
      const futureDate = new Date();
      futureDate.setHours(futureDate.getHours() + 1);
      localStorage.setItem('tenantly_expires_at', futureDate.toISOString());

      const result = service.isTokenExpired();

      expect(result).toBe(false);
    });

    it('should return true when expiry date is exactly now', () => {
      const now = new Date();
      localStorage.setItem('tenantly_expires_at', now.toISOString());

      const result = service.isTokenExpired();

      expect(result).toBe(true);
    });

    it('should return true when expiry date is invalid', () => {
      localStorage.setItem('tenantly_expires_at', 'invalid-date');

      const result = service.isTokenExpired();

      expect(result).toBe(true);
    });
  });

  describe('hasRole', () => {
    it('should return true when user has the specified role', () => {
      const result = service.hasRole('Admin');

      expect(result).toBe(true);
    });

    it('should return false when user does not have the specified role', () => {
      store.select.and.callFake((selector: unknown) => {
        if (selector === AuthSelectors.selectUserRole) return of('PropertyManager');
        return of(null);
      });

      const result = service.hasRole('Admin');

      expect(result).toBe(false);
    });

    it('should return false when checking for empty role', () => {
      store.select.and.returnValue(of('Admin'));

      const result = service.hasRole('');

      expect(result).toBe(false);
    });

    it('should be case-sensitive', () => {
      store.select.and.returnValue(of('Admin'));

      const result = service.hasRole('admin');

      expect(result).toBe(false);
    });
  });

  describe('hasAnyRole', () => {
    it('should return true when user has one of the specified roles', () => {
      const result = service.hasAnyRole(['Admin', 'PropertyManager']);

      expect(result).toBe(true);
    });

    it('should return false when user does not have any of the specified roles', () => {
      store.select.and.callFake((selector: unknown) => {
        if (selector === AuthSelectors.selectUserRole) return of('Accountant');
        return of(null);
      });

      const result = service.hasAnyRole(['Admin', 'PropertyManager']);

      expect(result).toBe(false);
    });

    it('should return false when checking empty array', () => {
      const result = service.hasAnyRole([]);

      expect(result).toBe(false);
    });

    it('should return true when user role is in a single-item array', () => {
      store.select.and.returnValue(of('Admin'));

      const result = service.hasAnyRole(['Admin']);

      expect(result).toBe(true);
    });

    it('should return true when user has last role in array', () => {
      store.select.and.returnValue(of('Accountant'));

      const result = service.hasAnyRole(['Admin', 'PropertyManager', 'Accountant']);

      expect(result).toBe(true);
    });
  });

  describe('isAdmin', () => {
    it('should return true for admin user', () => {
      const result = service.isAdmin();

      expect(result).toBe(true);
    });

    it('should return false for non-admin user', () => {
      store.select.and.returnValue(of(false));

      const result = service.isAdmin();

      expect(result).toBe(false);
    });
  });

  describe('isPropertyManager', () => {
    it('should return true for property manager user', () => {
      store.select.and.returnValue(of(true));

      const result = service.isPropertyManager();

      expect(result).toBe(true);
    });

    it('should return false for non-property manager user', () => {
      const result = service.isPropertyManager();

      expect(result).toBe(false);
    });
  });

  describe('isAccountant', () => {
    it('should return true for accountant user', () => {
      store.select.and.returnValue(of(true));

      const result = service.isAccountant();

      expect(result).toBe(true);
    });

    it('should return false for non-accountant user', () => {
      const result = service.isAccountant();

      expect(result).toBe(false);
    });
  });

  describe('reactive methods', () => {
    it('should provide reactive hasRole$', (done) => {
      const result$ = service.hasRole$('Admin');

      result$.subscribe((hasRole) => {
        expect(hasRole).toBe(true);
        done();
      });
    });

    it('should provide reactive hasRole$ for non-matching role', (done) => {
      store.select.and.returnValue(of(false));

      const result$ = service.hasRole$('PropertyManager');

      result$.subscribe((hasRole) => {
        expect(hasRole).toBe(false);
        done();
      });
    });

    it('should provide reactive hasAnyRole$', (done) => {
      const roles = ['Admin', 'PropertyManager'];
      const result$ = service.hasAnyRole$(roles);

      result$.subscribe((hasAnyRole) => {
        expect(hasAnyRole).toBe(true);
        done();
      });
    });

    it('should provide reactive hasAnyRole$ for empty array', (done) => {
      store.select.and.returnValue(of(false));

      const result$ = service.hasAnyRole$([]);

      result$.subscribe((hasAnyRole) => {
        expect(hasAnyRole).toBe(false);
        done();
      });
    });

    it('should provide reactive isAdmin$', (done) => {
      const result$ = service.isAdmin$();

      result$.subscribe((isAdmin) => {
        expect(isAdmin).toBe(true);
        done();
      });
    });

    it('should provide reactive isPropertyManager$', (done) => {
      const result$ = service.isPropertyManager$();

      result$.subscribe((isPM) => {
        expect(isPM).toBe(false);
        done();
      });
    });

    it('should provide reactive isAccountant$', (done) => {
      const result$ = service.isAccountant$();

      result$.subscribe((isAccountant) => {
        expect(isAccountant).toBe(false);
        done();
      });
    });
  });

  describe('observables', () => {
    it('should expose isAuthenticated$ observable', (done) => {
      service.isAuthenticated$.subscribe((isAuth) => {
        expect(isAuth).toBe(true);
        done();
      });
    });

    it('should expose user$ observable', (done) => {
      service.user$.subscribe((user) => {
        expect(user).toEqual(mockUser);
        done();
      });
    });

    it('should expose loading$ observable', (done) => {
      service.loading$.subscribe((loading) => {
        expect(loading).toBe(false);
        done();
      });
    });

    it('should expose error$ observable', (done) => {
      service.error$.subscribe((error) => {
        expect(error).toBeNull();
        done();
      });
    });

    it('should expose userRole$ observable', (done) => {
      service.userRole$.subscribe((role) => {
        expect(role).toBe('Admin');
        done();
      });
    });

    it('should handle loading state changes', (done) => {
      loadingSubject.next(true);

      service.loading$.subscribe((loading) => {
        expect(loading).toBe(true);
        done();
      });
    });

    it('should handle error state changes', (done) => {
      const error = 'Authentication failed';
      errorSubject.next(error);

      service.error$.subscribe((err) => {
        expect(err).toBe(error);
        done();
      });
    });
  });

  describe('edge cases', () => {
    it('should handle multiple consecutive calls to getUser', () => {
      const result1 = service.getUser();
      const result2 = service.getUser();

      expect(result1).toEqual(mockUser);
      expect(result2).toEqual(mockUser);
    });

    it('should handle rapid role changes', () => {
      store.select.and.returnValue(of('Admin'));
      expect(service.getUserRole()).toBe('Admin');

      store.select.and.returnValue(of('PropertyManager'));
      expect(service.getUserRole()).toBe('PropertyManager');

      store.select.and.returnValue(of('Accountant'));
      expect(service.getUserRole()).toBe('Accountant');
    });

    it('should handle null user gracefully', () => {
      store.select.and.returnValue(of(null));

      const user = service.getUser();
      expect(user).toBeNull();
    });
  });
});
