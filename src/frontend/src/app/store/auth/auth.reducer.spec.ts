import { authReducer, initialState, AuthState } from './auth.reducer';
import { Action } from '@ngrx/store';
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
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  describe('unknown action', () => {
    it('should return the previous state', () => {
      const action = {} as Action;
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

    it('should populate userOrganizations from login response', () => {
      const response = { ...mockLoginResponse, organizations: [mockOrg1, mockOrg2] };
      const state = authReducer(initialState, AuthActions.loginSuccess({ response }));

      expect(state.userOrganizations).toEqual([mockOrg1, mockOrg2]);
    });

    it('should set currentOrganizationId to organization_id when single org returned', () => {
      const response = { ...mockLoginResponse, organizations: [mockOrg1] };
      const state = authReducer(initialState, AuthActions.loginSuccess({ response }));

      expect(state.currentOrganizationId).toBe(100);
    });

    it('should set currentOrganizationId from default_organization_id when provided', () => {
      const response = {
        ...mockLoginResponse,
        organizations: [mockOrg1, mockOrg2],
        default_organization_id: 101,
      };
      const state = authReducer(initialState, AuthActions.loginSuccess({ response }));

      expect(state.currentOrganizationId).toBe(101);
    });

    it('should leave currentOrganizationId null when multiple orgs and no default', () => {
      const response = { ...mockLoginResponse, organizations: [mockOrg1, mockOrg2] };
      const state = authReducer(initialState, AuthActions.loginSuccess({ response }));

      expect(state.currentOrganizationId).toBeNull();
    });

    it('should set userOrganizations to empty array when no organizations in response', () => {
      const state = authReducer(
        initialState,
        AuthActions.loginSuccess({ response: mockLoginResponse })
      );

      expect(state.userOrganizations).toEqual([]);
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
        userOrganizations: [mockOrg1],
        currentOrganizationId: 100,
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
      expect(state.userOrganizations).toEqual([]);
      expect(state.currentOrganizationId).toBeNull();
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
      localStorage.setItem('tenantly_token', 'stored-token');
      localStorage.setItem('tenantly_refresh_token', 'stored-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockUser));
      localStorage.setItem('tenantly_expires_at', '2099-12-31T23:59:59Z');

      const action = AuthActions.initializeAuth();
      const state = authReducer(initialState, action);

      expect(state.user).toEqual(mockUser);
      expect(state.token).toBe('stored-token');
      expect(state.refreshToken).toBe('stored-refresh-token');
      expect(state.expiresAt).toBe('2099-12-31T23:59:59Z');
      expect(state.isAuthenticated).toBe(true);
    });

    it('should restore userOrganizations and currentOrganizationId from localStorage on initializeAuth', () => {
      localStorage.setItem('tenantly_token', 'stored-token');
      localStorage.setItem('tenantly_refresh_token', 'stored-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockUser));
      localStorage.setItem('tenantly_expires_at', '2099-12-31T23:59:59Z');
      localStorage.setItem('tenantly_organizations', JSON.stringify([mockOrg1, mockOrg2]));
      localStorage.setItem('tenantly_current_org_id', '100');

      const action = AuthActions.initializeAuth();
      const state = authReducer(initialState, action);

      expect(state.userOrganizations).toEqual([mockOrg1, mockOrg2]);
      expect(state.currentOrganizationId).toBe(100);
    });

    it('should not initialize auth from localStorage if token is expired', () => {
      localStorage.setItem('tenantly_token', 'expired-token');
      localStorage.setItem('tenantly_refresh_token', 'expired-refresh-token');
      localStorage.setItem('tenantly_user', JSON.stringify(mockUser));
      localStorage.setItem('tenantly_expires_at', '2020-01-01T00:00:00Z');

      const action = AuthActions.initializeAuth();
      const state = authReducer(initialState, action);

      expect(state).toEqual(initialState);
    });
  });

  describe('organization actions', () => {
    it('should set organizations and defaultOrganizationId on setUserOrganizations', () => {
      const action = AuthActions.setUserOrganizations({
        organizations: [mockOrg1, mockOrg2],
        defaultOrganizationId: 100,
      });
      const state = authReducer(initialState, action);

      expect(state.userOrganizations).toEqual([mockOrg1, mockOrg2]);
      expect(state.currentOrganizationId).toBe(100);
    });

    it('should preserve existing currentOrganizationId when no default provided', () => {
      const baseState: AuthState = { ...initialState, currentOrganizationId: 100 };
      const action = AuthActions.setUserOrganizations({ organizations: [mockOrg1] });
      const state = authReducer(baseState, action);

      expect(state.currentOrganizationId).toBe(100);
    });

    it('should update currentOrganizationId on setCurrentOrganization', () => {
      const action = AuthActions.setCurrentOrganization({ organizationId: 101 });
      const state = authReducer(initialState, action);

      expect(state.currentOrganizationId).toBe(101);
    });

    it('should set loading to true on switchOrganization', () => {
      const action = AuthActions.switchOrganization({ organizationId: 100 });
      const state = authReducer(initialState, action);

      expect(state.loading).toBe(true);
      expect(state.error).toBe(null);
    });

    it('should update token, refreshToken, expiresAt and currentOrganizationId on switchOrganizationSuccess', () => {
      const loadingState: AuthState = { ...initialState, loading: true };
      const response = {
        token: 'new-token',
        refresh_token: 'new-refresh-token',
        organization: mockOrg1,
        expires_at: '2099-06-01T00:00:00Z',
      };
      const action = AuthActions.switchOrganizationSuccess({ response });
      const state = authReducer(loadingState, action);

      expect(state.token).toBe('new-token');
      expect(state.refreshToken).toBe('new-refresh-token');
      expect(state.expiresAt).toBe('2099-06-01T00:00:00Z');
      expect(state.currentOrganizationId).toBe(100); // mockOrg1.organization_id
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
    });

    it('should convert Date expires_at to ISO string on switchOrganizationSuccess', () => {
      const date = new Date('2099-06-01T00:00:00Z');
      const response = {
        token: 'new-token',
        refresh_token: 'new-refresh',
        organization: mockOrg1,
        expires_at: date,
      };
      const action = AuthActions.switchOrganizationSuccess({ response });
      const state = authReducer(initialState, action);

      expect(state.expiresAt).toBe(date.toISOString());
    });

    it('should set error and clear loading on switchOrganizationFailure', () => {
      const loadingState: AuthState = { ...initialState, loading: true };
      const error = { message: 'Switch failed' };
      const action = AuthActions.switchOrganizationFailure({ error });
      const state = authReducer(loadingState, action);

      expect(state.loading).toBe(false);
      expect(state.error).toBe(error);
    });
  });
});
