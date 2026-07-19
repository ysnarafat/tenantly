import * as AuthSelectors from './auth.selectors';
import {
  selectUserRole,
  selectUserOrganizationId,
  selectIsSuperAdmin,
  selectIsOrgAdmin,
  selectCanManageUsers,
  selectCanManageOrganizations,
  selectCanInviteUsers,
} from './auth.selectors';
import { User } from '../../core/services/auth.service';
import { AuthState } from './auth.reducer';
import { AppState } from '../index';

describe('Auth Selectors', () => {
  const mockUser: User = {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    role: 'Admin',
    organization_id: 5,
    first_name: 'Test',
    last_name: 'User',
    active: true,
    created_at: '2026-04-01',
    updated_at: '2026-04-01',
  };

  const mockOrg1 = {
    id: 10,
    organization_id: 100,
    organization: {
      id: 100,
      name: 'Acme Corp',
      slug: 'acme-corp',
      subscriptionTier: 'basic' as const,
      maxUsers: 50,
      active: true,
      created_at: '',
      updated_at: '',
    },
    role: 'Admin',
    created_at: '',
    updated_at: '',
  };

  const mockOrg2 = {
    id: 11,
    organization_id: 101,
    organization: {
      id: 101,
      name: 'Beta Ltd',
      slug: 'beta-ltd',
      subscriptionTier: 'professional' as const,
      maxUsers: 100,
      active: true,
      created_at: '',
      updated_at: '',
    },
    role: 'PropertyManager',
    created_at: '',
    updated_at: '',
  };

  const mockAuthState: AuthState = {
    user: mockUser,
    token: 'test-token',
    expiresAt: '2099-12-31T23:59:59Z',
    isAuthenticated: true,
    loading: false,
    error: null,
    userOrganizations: [mockOrg1, mockOrg2],
    currentOrganizationId: 100,
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
    it('should extract user role from state', () => {
      const result = selectUserRole.projector(mockUser);
      expect(result).toBe('Admin');
    });

    it('should return empty string when user is null', () => {
      const result = selectUserRole.projector(null);
      expect(result).toBe('');
    });
  });

  describe('selectUserOrganizationId', () => {
    it('should extract organization_id from user', () => {
      const result = selectUserOrganizationId.projector(mockUser);
      expect(result).toBe(5);
    });

    it('should return null when user is null', () => {
      const result = selectUserOrganizationId.projector(null);
      expect(result).toBeNull();
    });

    it('should return null when organization_id is not set', () => {
      const userWithoutOrg = { ...mockUser, organization_id: undefined };
      const result = selectUserOrganizationId.projector(userWithoutOrg);
      expect(result).toBeNull();
    });
  });

  describe('selectIsSuperAdmin', () => {
    it('should return true when role is SUPER_ADMIN', () => {
      const result = selectIsSuperAdmin.projector('SUPER_ADMIN');
      expect(result).toBe(true);
    });

    it('should return false when role is not SUPER_ADMIN', () => {
      const result = selectIsSuperAdmin.projector('Admin');
      expect(result).toBe(false);
    });
  });

  describe('selectIsOrgAdmin', () => {
    it('should return true when role is ORG_ADMIN', () => {
      const result = selectIsOrgAdmin.projector('ORG_ADMIN');
      expect(result).toBe(true);
    });

    it('should return false when role is not ORG_ADMIN', () => {
      const result = selectIsOrgAdmin.projector('Admin');
      expect(result).toBe(false);
    });
  });

  describe('selectCanManageUsers', () => {
    it('should return true for admin roles', () => {
      expect(selectCanManageUsers.projector('SUPER_ADMIN')).toBe(true);
      expect(selectCanManageUsers.projector('ORG_ADMIN')).toBe(true);
      expect(selectCanManageUsers.projector('Admin')).toBe(true);
    });

    it('should return false for non-admin roles', () => {
      expect(selectCanManageUsers.projector('PropertyManager')).toBe(false);
      expect(selectCanManageUsers.projector('Accountant')).toBe(false);
    });
  });

  describe('selectCanManageOrganizations', () => {
    it('should return true only for SUPER_ADMIN', () => {
      expect(selectCanManageOrganizations.projector('SUPER_ADMIN')).toBe(true);
    });

    it('should return false for all other roles', () => {
      expect(selectCanManageOrganizations.projector('ORG_ADMIN')).toBe(false);
      expect(selectCanManageOrganizations.projector('Admin')).toBe(false);
      expect(selectCanManageOrganizations.projector('PropertyManager')).toBe(false);
      expect(selectCanManageOrganizations.projector('Accountant')).toBe(false);
    });
  });

  describe('selectCanInviteUsers', () => {
    it('should return true for admin roles', () => {
      expect(selectCanInviteUsers.projector('SUPER_ADMIN')).toBe(true);
      expect(selectCanInviteUsers.projector('ORG_ADMIN')).toBe(true);
      expect(selectCanInviteUsers.projector('Admin')).toBe(true);
    });

    it('should return false for non-admin roles', () => {
      expect(selectCanInviteUsers.projector('PropertyManager')).toBe(false);
      expect(selectCanInviteUsers.projector('Accountant')).toBe(false);
    });
  });

  describe('Selector composition', () => {
    it('should provide consistent results across related selectors', () => {
      const role = 'ORG_ADMIN';

      const isOrgAdminResult = selectIsOrgAdmin.projector(role);
      const canManageUsersResult = selectCanManageUsers.projector(role);
      const canManageOrgsResult = selectCanManageOrganizations.projector(role);

      expect(isOrgAdminResult).toBe(true);
      expect(canManageUsersResult).toBe(true);
      expect(canManageOrgsResult).toBe(false);
    });
  });

  describe('Permission matrix validation', () => {
    it('should follow the permission hierarchy', () => {
      const roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin'];

      // All admin roles should be able to manage users
      roles.forEach((role) => {
        expect(selectCanManageUsers.projector(role)).toBe(true);
      });

      // Only SUPER_ADMIN should manage organizations
      expect(selectCanManageOrganizations.projector('SUPER_ADMIN')).toBe(true);
      roles.slice(1).forEach((role) => {
        expect(selectCanManageOrganizations.projector(role)).toBe(false);
      });
    });

    it('should ensure lower roles do not have admin capabilities', () => {
      const lowerRoles = ['PropertyManager', 'Accountant'];

      lowerRoles.forEach((role) => {
        expect(selectCanManageUsers.projector(role)).toBe(false);
        expect(selectCanManageOrganizations.projector(role)).toBe(false);
        expect(selectCanInviteUsers.projector(role)).toBe(false);
      });
    });
  });

  describe('selectUserOrganizations', () => {
    it('should select the userOrganizations array', () => {
      const result = AuthSelectors.selectUserOrganizations.projector(mockAuthState);
      expect(result).toEqual([mockOrg1, mockOrg2]);
    });

    it('should return empty array when no organizations', () => {
      const state = { ...mockAuthState, userOrganizations: [] };
      const result = AuthSelectors.selectUserOrganizations.projector(state);
      expect(result).toEqual([]);
    });
  });

  describe('selectCurrentOrganizationId', () => {
    it('should select the currentOrganizationId', () => {
      const result = AuthSelectors.selectCurrentOrganizationId.projector(mockAuthState);
      expect(result).toBe(100);
    });

    it('should return null when no organization is selected', () => {
      const state = { ...mockAuthState, currentOrganizationId: null };
      const result = AuthSelectors.selectCurrentOrganizationId.projector(state);
      expect(result).toBeNull();
    });
  });
});
