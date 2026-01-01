import { authReducer, initialState, AuthState } from './auth.reducer';
import * as AuthActions from './auth.actions';

describe('AuthReducer', () => {
  const mockUser = {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    role: 'Admin',
    active: true,
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2023-01-01T00:00:00Z',
  };

  const mockLoginResponse = {
    token: 'test-token',
    refresh_token: 'test-refresh-token',
    user: mockUser,
    expires_at: '2023-12-31T23:59:59Z',
  };

  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  describe('unknown action', () => {
    it('should return the previous state', () => {
      const action = {} as any;
      const result = authReducer(initialState, action);
      expect(result).toBe(initialState);
    });
  });

  describe('login actions', () => {
    it('should set loading to true on login', () => {
      const credentials = { username: 'test', password: 'test' };
      const action = AuthActions.login({ credentials });
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(true);
      expect(state.error).toBe(null);
    });

    it('should set user data and authentication state on loginSuccess', () => {
      const action = AuthActions.loginSuccess({ response: mockLoginResponse });
      const state = authReducer(initialState, action);

      expect(state.user).toEqual(mockUser);
      expect(state.token).toBe('test-token');
      expect(state.refreshToken).toBe('test-refresh-token');
      expect(state.expiresAt).toBe('2023-12-31T23:59:59Z');
      expect(state.isAuthenticated).toBe(true);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    it('should clear user data and set error on loginFailure', () => {
      const error = { message: 'Invalid credentials' };
      const action = AuthActions.loginFailure({ error });
      const state = authReducer(initialState, action);

      expect(state.user).toBe(null);
      expect(state.token).toBe(null);
      expect(state.refreshToken).toBe(null);
      expect(state.expiresAt).toBe(null);
      expect(state.isAuthenticated).toBe(false);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(error);
    });
  });

  describe('logout actions', () => {
    it('should set loading to true on logout', () => {
      const authenticatedState: AuthState = {
        ...initialState,
        user: mockUser,
        token: 'test-token',
        isAuthenticated: true,
      };
      const action = AuthActions.logout();
      const state = authReducer(authenticatedState, action);

      expect(state.loading).toBe(true);
      expect(state.error).toBe(null);
    });

    it('should reset to initial state on logoutSuccess', () => {
      const authenticatedState: AuthState = {
        ...initialState,
        user: mockUser,
        token: 'test-token',
        isAuthenticated: true,
      };
      const action = AuthActions.logoutSuccess();
      const state = authReducer(authenticatedState, action);

      expect(state.user).toBe(null);
      expect(state.token).toBe(null);
      expect(state.refreshToken).toBe(null);
      expect(state.expiresAt).toBe(null);
      expect(state.isAuthenticated).toBe(false);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    it('should set error on logoutFailure', () => {
      const error = { message: 'Logout failed' };
      const action = AuthActions.logoutFailure({ error });
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(false);
      expect(state.error).toBe(error);
    });
  });

  describe('refresh token actions', () => {
    it('should set loading to true on refreshToken', () => {
      const action = AuthActions.refreshToken();
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(true);
      expect(state.error).toBe(null);
    });

    it('should update tokens and user on refreshTokenSuccess', () => {
      const action = AuthActions.refreshTokenSuccess({ response: mockLoginResponse });
      const state = authReducer(initialState, action);

      expect(state.user).toEqual(mockUser);
      expect(state.token).toBe('test-token');
      expect(state.refreshToken).toBe('test-refresh-token');
      expect(state.expiresAt).toBe('2023-12-31T23:59:59Z');
      expect(state.isAuthenticated).toBe(true);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    it('should clear tokens and set unauthenticated on refreshTokenFailure', () => {
      const authenticatedState: AuthState = {
        ...initialState,
        user: mockUser,
        token: 'test-token',
        isAuthenticated: true,
      };
      const error = { message: 'Token refresh failed' };
      const action = AuthActions.refreshTokenFailure({ error });
      const state = authReducer(authenticatedState, action);

      expect(state.token).toBe(null);
      expect(state.refreshToken).toBe(null);
      expect(state.expiresAt).toBe(null);
      expect(state.isAuthenticated).toBe(false);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(error);
    });
  });

  describe('password management actions', () => {
    it('should set loading to true on changePassword', () => {
      const request = { current_password: 'old', new_password: 'new' };
      const action = AuthActions.changePassword({ request });
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(true);
      expect(state.error).toBe(null);
    });

    it('should clear loading on changePasswordSuccess', () => {
      const loadingState: AuthState = { ...initialState, loading: true };
      const action = AuthActions.changePasswordSuccess({ message: 'Success' });
      const state = authReducer(loadingState, action);

      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    it('should set error on changePasswordFailure', () => {
      const error = { message: 'Password change failed' };
      const action = AuthActions.changePasswordFailure({ error });
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(false);
      expect(state.error).toBe(error);
    });

    it('should set loading to true on resetPassword', () => {
      const request = { email: 'test@example.com' };
      const action = AuthActions.resetPassword({ request });
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(true);
      expect(state.error).toBe(null);
    });

    it('should clear loading on resetPasswordSuccess', () => {
      const loadingState: AuthState = { ...initialState, loading: true };
      const action = AuthActions.resetPasswordSuccess({ message: 'Success' });
      const state = authReducer(loadingState, action);

      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    it('should set error on resetPasswordFailure', () => {
      const error = { message: 'Reset failed' };
      const action = AuthActions.resetPasswordFailure({ error });
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(false);
      expect(state.error).toBe(error);
    });
  });

  describe('utility actions', () => {
    it('should clear error on clearError', () => {
      const errorState: AuthState = { ...initialState, error: { message: 'Some error' } };
      const action = AuthActions.clearError();
      const state = authReducer(errorState, action);

      expect(state.error).toBe(null);
    });

    it('should set loading state on setLoading', () => {
      const action = AuthActions.setLoading({ loading: true });
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(true);
    });

    it('should initialize auth from localStorage on initializeAuth', () => {
      // Set up localStorage
      localStorage.setItem('tenantly_token', 'stored-token');
      localStorage.setItem('tenantly_refresh_token', 'stored-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockUser));
      localStorage.setItem('tenantly_expires_at', '2099-12-31T23:59:59Z'); // Future date

      const action = AuthActions.initializeAuth();
      const state = authReducer(initialState, action);

      expect(state.user).toEqual(mockUser);
      expect(state.token).toBe('stored-token');
      expect(state.refreshToken).toBe('stored-refresh-token');
      expect(state.expiresAt).toBe('2099-12-31T23:59:59Z');
      expect(state.isAuthenticated).toBe(true);
    });

    it('should not initialize auth from localStorage if token is expired', () => {
      // Set up localStorage with expired token
      localStorage.setItem('tenantly_token', 'expired-token');
      localStorage.setItem('tenantly_refresh_token', 'expired-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockUser));
      localStorage.setItem('tenantly_expires_at', '2020-01-01T00:00:00Z'); // Past date

      const action = AuthActions.initializeAuth();
      const state = authReducer(initialState, action);

      // Should remain in initial state since token is expired
      expect(state).toEqual(initialState);
    });
  });
});
