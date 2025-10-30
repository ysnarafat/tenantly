import * as AuthSelectors from './auth.selectors';
import { AuthState } from './auth.reducer';
import { AppState } from '../index';

describe('Auth Selectors', () => {
  const mockUser = {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    role: 'Admin',
    active: true,
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2023-01-01T00:00:00Z',
  };

  const mockAuthState: AuthState = {
    user: mockUser,
    token: 'test-token',
    refreshToken: 'test-refresh-token',
    expiresAt: '2099-12-31T23:59:59Z',
    isAuthenticated: true,
    loading: false,
    error: null,
  };

  const mockAppState: AppState = {
    auth: mockAuthState,
  };

  describe('selectAuthState', () => {
    it('should select the auth state', () => {
      const result = AuthSelectors.selectAuthState(mockAppState);
      expect(result).toEqual(mockAuthState);
    });
  });

  describe('selectUser', () => {
    it('should select the user', () => {
      const result = AuthSelectors.selectUser.projector(mockAuthState);
      expect(result).toEqual(mockUser);
    });

    it('should return null when user is null', () => {
      const state = { ...mockAuthState, user: null };
      const result = AuthSelectors.selectUser.projector(state);
      expect(result).toBeNull();
    });
  });

  describe('selectToken', () => {
    it('should select the token', () => {
      const result = AuthSelectors.selectToken.projector(mockAuthState);
      expect(result).toBe('test-token');
    });
  });

  describe('selectRefreshToken', () => {
    it('should select the refresh token', () => {
      const result = AuthSelectors.selectRefreshToken.projector(mockAuthState);
      expect(result).toBe('test-refresh-token');
    });
  });

  describe('selectIsAuthenticated', () => {
    it('should select the authentication status', () => {
      const result = AuthSelectors.selectIsAuthenticated.projector(mockAuthState);
      expect(result).toBe(true);
    });

    it('should return false when not authenticated', () => {
      const state = { ...mockAuthState, isAuthenticated: false };
      const result = AuthSelectors.selectIsAuthenticated.projector(state);
      expect(result).toBe(false);
    });
  });

  describe('selectAuthLoading', () => {
    it('should select the loading state', () => {
      const result = AuthSelectors.selectAuthLoading.projector(mockAuthState);
      expect(result).toBe(false);
    });

    it('should return true when loading', () => {
      const state = { ...mockAuthState, loading: true };
      const result = AuthSelectors.selectAuthLoading.projector(state);
      expect(result).toBe(true);
    });
  });

  describe('selectAuthError', () => {
    it('should select the error', () => {
      const result = AuthSelectors.selectAuthError.projector(mockAuthState);
      expect(result).toBeNull();
    });

    it('should return error when present', () => {
      const error = { message: 'Test error' };
      const state = { ...mockAuthState, error };
      const result = AuthSelectors.selectAuthError.projector(state);
      expect(result).toEqual(error);
    });
  });

  describe('selectUserRole', () => {
    it('should select the user role', () => {
      const result = AuthSelectors.selectUserRole.projector(mockUser);
      expect(result).toBe('Admin');
    });

    it('should return empty string when user is null', () => {
      const result = AuthSelectors.selectUserRole.projector(null);
      expect(result).toBe('');
    });
  });

  describe('selectIsAdmin', () => {
    it('should return true for Admin role', () => {
      const result = AuthSelectors.selectIsAdmin.projector('Admin');
      expect(result).toBe(true);
    });

    it('should return false for non-Admin role', () => {
      const result = AuthSelectors.selectIsAdmin.projector('PropertyManager');
      expect(result).toBe(false);
    });
  });

  describe('selectIsPropertyManager', () => {
    it('should return true for PropertyManager role', () => {
      const result = AuthSelectors.selectIsPropertyManager.projector('PropertyManager');
      expect(result).toBe(true);
    });

    it('should return false for non-PropertyManager role', () => {
      const result = AuthSelectors.selectIsPropertyManager.projector('Admin');
      expect(result).toBe(false);
    });
  });

  describe('selectIsAccountant', () => {
    it('should return true for Accountant role', () => {
      const result = AuthSelectors.selectIsAccountant.projector('Accountant');
      expect(result).toBe(true);
    });

    it('should return false for non-Accountant role', () => {
      const result = AuthSelectors.selectIsAccountant.projector('Admin');
      expect(result).toBe(false);
    });
  });

  describe('selectHasRole', () => {
    it('should return true when user has the specified role', () => {
      const selector = AuthSelectors.selectHasRole('Admin');
      const result = selector.projector('Admin');
      expect(result).toBe(true);
    });

    it('should return false when user does not have the specified role', () => {
      const selector = AuthSelectors.selectHasRole('Admin');
      const result = selector.projector('PropertyManager');
      expect(result).toBe(false);
    });
  });

  describe('selectHasAnyRole', () => {
    it('should return true when user has one of the specified roles', () => {
      const selector = AuthSelectors.selectHasAnyRole(['Admin', 'PropertyManager']);
      const result = selector.projector('Admin');
      expect(result).toBe(true);
    });

    it('should return false when user does not have any of the specified roles', () => {
      const selector = AuthSelectors.selectHasAnyRole(['Admin', 'PropertyManager']);
      const result = selector.projector('Accountant');
      expect(result).toBe(false);
    });
  });

  describe('selectIsTokenExpired', () => {
    it('should return false for future expiration date', () => {
      const state = { ...mockAuthState, expiresAt: '2099-12-31T23:59:59Z' };
      const result = AuthSelectors.selectIsTokenExpired.projector(state);
      expect(result).toBe(false);
    });

    it('should return true for past expiration date', () => {
      const state = { ...mockAuthState, expiresAt: '2020-01-01T00:00:00Z' };
      const result = AuthSelectors.selectIsTokenExpired.projector(state);
      expect(result).toBe(true);
    });

    it('should return true when expiresAt is null', () => {
      const state = { ...mockAuthState, expiresAt: null };
      const result = AuthSelectors.selectIsTokenExpired.projector(state);
      expect(result).toBe(true);
    });
  });

  describe('selectCanViewLeases', () => {
    it('should return true for Admin role', () => {
      const result = AuthSelectors.selectCanViewLeases.projector('Admin');
      expect(result).toBe(true);
    });

    it('should return true for PropertyManager role', () => {
      const result = AuthSelectors.selectCanViewLeases.projector('PropertyManager');
      expect(result).toBe(true);
    });

    it('should return false for Accountant role', () => {
      const result = AuthSelectors.selectCanViewLeases.projector('Accountant');
      expect(result).toBe(false);
    });
  });

  describe('selectCanViewAttachments', () => {
    it('should return true for Admin role', () => {
      const result = AuthSelectors.selectCanViewAttachments.projector('Admin');
      expect(result).toBe(true);
    });

    it('should return true for PropertyManager role', () => {
      const result = AuthSelectors.selectCanViewAttachments.projector('PropertyManager');
      expect(result).toBe(true);
    });

    it('should return false for Accountant role', () => {
      const result = AuthSelectors.selectCanViewAttachments.projector('Accountant');
      expect(result).toBe(false);
    });
  });

  describe('selectCanManageUsers', () => {
    it('should return true for Admin role', () => {
      const result = AuthSelectors.selectCanManageUsers.projector('Admin');
      expect(result).toBe(true);
    });

    it('should return false for PropertyManager role', () => {
      const result = AuthSelectors.selectCanManageUsers.projector('PropertyManager');
      expect(result).toBe(false);
    });

    it('should return false for Accountant role', () => {
      const result = AuthSelectors.selectCanManageUsers.projector('Accountant');
      expect(result).toBe(false);
    });
  });
});