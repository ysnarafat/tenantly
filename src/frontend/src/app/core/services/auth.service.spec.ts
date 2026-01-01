import { TestBed } from '@angular/core/testing';
import { Store } from '@ngrx/store';
import { of } from 'rxjs';
import {
  AuthService,
  LoginRequest,
  ChangePasswordRequest,
  ResetPasswordRequest,
} from './auth.service';
import { AppState } from '../../store';
import * as AuthActions from '../../store/auth/auth.actions';
import * as AuthSelectors from '../../store/auth/auth.selectors';

describe('AuthService', () => {
  let service: AuthService;
  let store: jasmine.SpyObj<Store<AppState>>;

  const mockUser = {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    role: 'Admin',
    active: true,
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2023-01-01T00:00:00Z',
  };

  beforeEach(() => {
    const storeSpy = jasmine.createSpyObj('Store', ['select', 'dispatch']);

    TestBed.configureTestingModule({
      providers: [AuthService, { provide: Store, useValue: storeSpy }],
    });

    service = TestBed.inject(AuthService);
    store = TestBed.inject(Store) as jasmine.SpyObj<Store<AppState>>;

    // Setup default store selectors
    store.select.and.callFake((selector: any) => {
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
  });

  describe('logout', () => {
    it('should dispatch logout action', () => {
      service.logout();

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
      const request: ChangePasswordRequest = { current_password: 'old', new_password: 'new' };

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

  describe('getUser', () => {
    it('should return user from store', () => {
      const result = service.getUser();

      expect(result).toEqual(mockUser);
      expect(store.select).toHaveBeenCalledWith(AuthSelectors.selectUser);
    });
  });

  describe('getUserRole', () => {
    it('should return user role from store', () => {
      const result = service.getUserRole();

      expect(result).toBe('Admin');
      expect(store.select).toHaveBeenCalledWith(AuthSelectors.selectUserRole);
    });
  });

  describe('isAuthenticated', () => {
    it('should return authentication status from store', () => {
      const result = service.isAuthenticated();

      expect(result).toBe(true);
      expect(store.select).toHaveBeenCalledWith(AuthSelectors.selectIsAuthenticated);
    });
  });

  describe('hasRole', () => {
    it('should return true when user has the specified role', () => {
      const result = service.hasRole('Admin');

      expect(result).toBe(true);
    });

    it('should return false when user does not have the specified role', () => {
      store.select.and.callFake((selector: any) => {
        if (selector === AuthSelectors.selectUserRole) return of('PropertyManager');
        return of(null);
      });

      const result = service.hasRole('Admin');

      expect(result).toBe(false);
    });
  });

  describe('hasAnyRole', () => {
    it('should return true when user has one of the specified roles', () => {
      const result = service.hasAnyRole(['Admin', 'PropertyManager']);

      expect(result).toBe(true);
    });

    it('should return false when user does not have any of the specified roles', () => {
      store.select.and.callFake((selector: any) => {
        if (selector === AuthSelectors.selectUserRole) return of('Accountant');
        return of(null);
      });

      const result = service.hasAnyRole(['Admin', 'PropertyManager']);

      expect(result).toBe(false);
    });
  });

  describe('isAdmin', () => {
    it('should return true for admin user', () => {
      const result = service.isAdmin();

      expect(result).toBe(true);
      expect(store.select).toHaveBeenCalledWith(AuthSelectors.selectIsAdmin);
    });
  });

  describe('isPropertyManager', () => {
    it('should return false for admin user', () => {
      const result = service.isPropertyManager();

      expect(result).toBe(false);
      expect(store.select).toHaveBeenCalledWith(AuthSelectors.selectIsPropertyManager);
    });
  });

  describe('isAccountant', () => {
    it('should return false for admin user', () => {
      const result = service.isAccountant();

      expect(result).toBe(false);
      expect(store.select).toHaveBeenCalledWith(AuthSelectors.selectIsAccountant);
    });
  });

  describe('reactive methods', () => {
    it('should provide reactive hasRole$', () => {
      const result$ = service.hasRole$('Admin');

      expect(result$).toBeDefined();
    });

    it('should provide reactive hasAnyRole$', () => {
      const roles = ['Admin', 'PropertyManager'];
      const result$ = service.hasAnyRole$(roles);

      expect(result$).toBeDefined();
    });

    it('should provide reactive isAdmin$', () => {
      const result$ = service.isAdmin$();

      expect(result$).toBeDefined();
    });

    it('should provide reactive isPropertyManager$', () => {
      const result$ = service.isPropertyManager$();

      expect(result$).toBeDefined();
    });

    it('should provide reactive isAccountant$', () => {
      const result$ = service.isAccountant$();

      expect(result$).toBeDefined();
    });
  });

  describe('observables', () => {
    it('should expose isAuthenticated$ observable', () => {
      expect(service.isAuthenticated$).toBeDefined();
    });

    it('should expose user$ observable', () => {
      expect(service.user$).toBeDefined();
    });

    it('should expose loading$ observable', () => {
      expect(service.loading$).toBeDefined();
    });

    it('should expose error$ observable', () => {
      expect(service.error$).toBeDefined();
    });

    it('should expose userRole$ observable', () => {
      expect(service.userRole$).toBeDefined();
    });
  });
});
