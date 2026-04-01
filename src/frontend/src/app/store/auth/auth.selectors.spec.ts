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
});
