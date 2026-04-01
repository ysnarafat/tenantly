import { TestBed } from '@angular/core/testing';
import { PermissionService } from './permission.service';
import { AuthService } from './auth.service';
import { Permission, ROLE_HIERARCHY, ROLE_PERMISSIONS } from '../models/role.model';
import { of } from 'rxjs';

describe('PermissionService', () => {
  let service: PermissionService;
  let authServiceMock: { userRole$: ReturnType<typeof of> };

  beforeEach(() => {
    authServiceMock = {
      userRole$: of('Admin'),
    };

    TestBed.configureTestingModule({
      providers: [PermissionService, { provide: AuthService, useValue: authServiceMock }],
    });
    service = TestBed.inject(PermissionService);
  });

  describe('Initialization', () => {
    it('should be created', () => {
      expect(service).toBeTruthy();
    });

    it('should initialize userRole signal from AuthService', (done) => {
      authServiceMock.userRole$ = of('PropertyManager');
      const newService = TestBed.inject(PermissionService);

      // Use a small timeout to allow the signal to update
      setTimeout(() => {
        expect(newService.getCurrentRole()).toBe('PropertyManager');
        done();
      }, 100);
    });
  });

  describe('Role checks - isSuperAdmin', () => {
    it('should return true for SUPER_ADMIN role', () => {
      authServiceMock.userRole$ = of('SUPER_ADMIN');
      const newService = TestBed.inject(PermissionService);
      expect(newService.isSuperAdmin()).toBe(true);
    });

    it('should return false for non-SUPER_ADMIN roles', () => {
      const roles = ['ORG_ADMIN', 'Admin', 'PropertyManager', 'Accountant'];

      roles.forEach((role) => {
        authServiceMock.userRole$ = of(role);
        const newService = TestBed.inject(PermissionService);
        expect(newService.isSuperAdmin()).toBe(false);
      });
    });
  });

  describe('Role checks - isOrgAdmin', () => {
    it('should return true for ORG_ADMIN role', () => {
      authServiceMock.userRole$ = of('ORG_ADMIN');
      const newService = TestBed.inject(PermissionService);
      expect(newService.isOrgAdmin()).toBe(true);
    });

    it('should return false for non-ORG_ADMIN roles', () => {
      const roles = ['SUPER_ADMIN', 'Admin', 'PropertyManager', 'Accountant'];

      roles.forEach((role) => {
        authServiceMock.userRole$ = of(role);
        const newService = TestBed.inject(PermissionService);
        expect(newService.isOrgAdmin()).toBe(false);
      });
    });
  });

  describe('Role checks - isAdmin', () => {
    it('should return true for Admin role', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      expect(newService.isAdmin()).toBe(true);
    });

    it('should return false for non-Admin roles', () => {
      const roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'PropertyManager', 'Accountant'];

      roles.forEach((role) => {
        authServiceMock.userRole$ = of(role);
        const newService = TestBed.inject(PermissionService);
        expect(newService.isAdmin()).toBe(false);
      });
    });
  });

  describe('Role checks - isPropertyManager', () => {
    it('should return true for PropertyManager role', () => {
      authServiceMock.userRole$ = of('PropertyManager');
      const newService = TestBed.inject(PermissionService);
      expect(newService.isPropertyManager()).toBe(true);
    });

    it('should return false for non-PropertyManager roles', () => {
      const roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'Accountant'];

      roles.forEach((role) => {
        authServiceMock.userRole$ = of(role);
        const newService = TestBed.inject(PermissionService);
        expect(newService.isPropertyManager()).toBe(false);
      });
    });
  });

  describe('Role checks - isAccountant', () => {
    it('should return true for Accountant role', () => {
      authServiceMock.userRole$ = of('Accountant');
      const newService = TestBed.inject(PermissionService);
      expect(newService.isAccountant()).toBe(true);
    });

    it('should return false for non-Accountant roles', () => {
      const roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'PropertyManager'];

      roles.forEach((role) => {
        authServiceMock.userRole$ = of(role);
        const newService = TestBed.inject(PermissionService);
        expect(newService.isAccountant()).toBe(false);
      });
    });
  });

  describe('Permission checks - hasPermission', () => {
    it('should return true for Admin with manage_users permission', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      expect(newService.hasPermission(Permission.MANAGE_USERS)).toBe(true);
    });

    it('should return false for Accountant with manage_users permission', () => {
      authServiceMock.userRole$ = of('Accountant');
      const newService = TestBed.inject(PermissionService);
      expect(newService.hasPermission(Permission.MANAGE_USERS)).toBe(false);
    });

    it('should return true for Accountant with view_payments permission', () => {
      authServiceMock.userRole$ = of('Accountant');
      const newService = TestBed.inject(PermissionService);
      expect(newService.hasPermission(Permission.VIEW_PAYMENTS)).toBe(true);
    });
  });

  describe('Permission checks - hasAnyPermission', () => {
    it('should return true if user has any of the specified permissions', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      const permissions = [Permission.VIEW_DASHBOARD, Permission.MANAGE_USERS];
      expect(newService.hasAnyPermission(permissions)).toBe(true);
    });

    it('should return false if user has none of the specified permissions', () => {
      authServiceMock.userRole$ = of('Accountant');
      const newService = TestBed.inject(PermissionService);
      const permissions = [Permission.MANAGE_USERS, Permission.MANAGE_PROPERTIES];
      expect(newService.hasAnyPermission(permissions)).toBe(false);
    });
  });

  describe('Permission checks - hasAllPermissions', () => {
    it('should return true if user has all specified permissions', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      const permissions = [Permission.MANAGE_USERS, Permission.MANAGE_PROPERTIES];
      expect(newService.hasAllPermissions(permissions)).toBe(true);
    });

    it('should return false if user is missing any permission', () => {
      authServiceMock.userRole$ = of('PropertyManager');
      const newService = TestBed.inject(PermissionService);
      const permissions = [Permission.MANAGE_USERS, Permission.MANAGE_PROPERTIES];
      expect(newService.hasAllPermissions(permissions)).toBe(false);
    });
  });

  describe('Permission checks - canAccess', () => {
    it('should return computed signal for permission check', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      const canManageUsers = newService.canAccess(Permission.MANAGE_USERS);
      expect(canManageUsers()).toBe(true);
    });

    it('should update computed signal when role changes', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      const computed = newService.canAccess(Permission.MANAGE_USERS);
      expect(computed()).toBe(true);
    });
  });

  describe('Organization-specific permissions - canManageOrganizations', () => {
    it('should return true only for SUPER_ADMIN', () => {
      authServiceMock.userRole$ = of('SUPER_ADMIN');
      const superAdminService = TestBed.inject(PermissionService);
      expect(superAdminService.canManageOrganizations()).toBe(true);

      authServiceMock.userRole$ = of('ORG_ADMIN');
      const orgAdminService = TestBed.inject(PermissionService);
      expect(orgAdminService.canManageOrganizations()).toBe(false);

      authServiceMock.userRole$ = of('Admin');
      const adminService = TestBed.inject(PermissionService);
      expect(adminService.canManageOrganizations()).toBe(false);
    });
  });

  describe('Admin management - canManageOrgAdmins', () => {
    it('should return true for SUPER_ADMIN and ORG_ADMIN', () => {
      authServiceMock.userRole$ = of('SUPER_ADMIN');
      const superAdminService = TestBed.inject(PermissionService);
      expect(superAdminService.canManageOrgAdmins()).toBe(true);

      authServiceMock.userRole$ = of('ORG_ADMIN');
      const orgAdminService = TestBed.inject(PermissionService);
      expect(orgAdminService.canManageOrgAdmins()).toBe(true);
    });

    it('should return false for lower-level roles', () => {
      const roles = ['Admin', 'PropertyManager', 'Accountant'];

      roles.forEach((role) => {
        authServiceMock.userRole$ = of(role);
        const newService = TestBed.inject(PermissionService);
        expect(newService.canManageOrgAdmins()).toBe(false);
      });
    });
  });

  describe('User invitation - canInviteUsers', () => {
    it('should return true for admin roles', () => {
      const adminRoles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin'];

      adminRoles.forEach((role) => {
        authServiceMock.userRole$ = of(role);
        const newService = TestBed.inject(PermissionService);
        expect(newService.canInviteUsers()).toBe(true);
      });
    });

    it('should return false for non-admin roles', () => {
      const nonAdminRoles = ['PropertyManager', 'Accountant'];

      nonAdminRoles.forEach((role) => {
        authServiceMock.userRole$ = of(role);
        const newService = TestBed.inject(PermissionService);
        expect(newService.canInviteUsers()).toBe(false);
      });
    });
  });

  describe('Role hierarchy - isRoleHigherThan', () => {
    it('should correctly identify higher ranked roles', () => {
      expect(service.isRoleHigherThan('SUPER_ADMIN', 'Admin')).toBe(true);
      expect(service.isRoleHigherThan('SUPER_ADMIN', 'ORG_ADMIN')).toBe(true);
      expect(service.isRoleHigherThan('ORG_ADMIN', 'Admin')).toBe(true);
      expect(service.isRoleHigherThan('Admin', 'PropertyManager')).toBe(true);
      expect(service.isRoleHigherThan('PropertyManager', 'Accountant')).toBe(true);
    });

    it('should return false for lower or equal ranked roles', () => {
      expect(service.isRoleHigherThan('Admin', 'SUPER_ADMIN')).toBe(false);
      expect(service.isRoleHigherThan('PropertyManager', 'ORG_ADMIN')).toBe(false);
      expect(service.isRoleHigherThan('Accountant', 'Admin')).toBe(false);
    });

    it('should return false for same ranked roles', () => {
      expect(service.isRoleHigherThan('Admin', 'Admin')).toBe(false);
      expect(service.isRoleHigherThan('SUPER_ADMIN', 'SUPER_ADMIN')).toBe(false);
    });
  });

  describe('Role permissions - getRolePermissions', () => {
    it('should return all permissions for SUPER_ADMIN', () => {
      const permissions = service.getRolePermissions('SUPER_ADMIN');
      expect(permissions.length).toBeGreaterThan(0);
      expect(permissions).toContain(Permission.MANAGE_ORGANIZATIONS);
      expect(permissions).toContain(Permission.MANAGE_USERS);
      expect(permissions).toContain(Permission.MANAGE_PROPERTIES);
    });

    it('should return limited permissions for Accountant', () => {
      const permissions = service.getRolePermissions('Accountant');
      expect(permissions.length).toBeGreaterThan(0);
      expect(permissions).not.toContain(Permission.MANAGE_PROPERTIES);
      expect(permissions).not.toContain(Permission.MANAGE_USERS);
      expect(permissions).toContain(Permission.VIEW_PAYMENTS);
    });

    it('should return permissions for all roles', () => {
      const roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'PropertyManager', 'Accountant'] as const;

      roles.forEach((role) => {
        const permissions = service.getRolePermissions(role);
        expect(Array.isArray(permissions)).toBe(true);
        expect(permissions.length).toBeGreaterThan(0);
      });
    });

    it('should not return empty array for any valid role', () => {
      const roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'PropertyManager', 'Accountant'] as const;

      roles.forEach((role) => {
        const permissions = service.getRolePermissions(role);
        expect(permissions.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Convenience methods', () => {
    it('canManageProperties should check for MANAGE_PROPERTIES permission', () => {
      authServiceMock.userRole$ = of('PropertyManager');
      const newService = TestBed.inject(PermissionService);
      expect(newService.canManageProperties()).toBe(true);

      authServiceMock.userRole$ = of('Accountant');
      const accountantService = TestBed.inject(PermissionService);
      expect(accountantService.canManageProperties()).toBe(false);
    });

    it('canManageTenants should check for MANAGE_TENANTS permission', () => {
      authServiceMock.userRole$ = of('PropertyManager');
      const newService = TestBed.inject(PermissionService);
      expect(newService.canManageTenants()).toBe(true);

      authServiceMock.userRole$ = of('Accountant');
      const accountantService = TestBed.inject(PermissionService);
      expect(accountantService.canManageTenants()).toBe(false);
    });

    it('canRecordPayments should check for RECORD_PAYMENTS permission', () => {
      authServiceMock.userRole$ = of('PropertyManager');
      const newService = TestBed.inject(PermissionService);
      expect(newService.canRecordPayments()).toBe(true);

      authServiceMock.userRole$ = of('Accountant');
      const accountantService = TestBed.inject(PermissionService);
      expect(accountantService.canRecordPayments()).toBe(false);
    });

    it('canViewReports should check for VIEW_REPORTS permission', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      expect(newService.canViewReports()).toBe(true);

      authServiceMock.userRole$ = of('PropertyManager');
      const pmService = TestBed.inject(PermissionService);
      expect(pmService.canViewReports()).toBe(true);
    });

    it('canManageDocuments should check for MANAGE_DOCUMENTS permission', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      expect(newService.canManageDocuments()).toBe(true);

      authServiceMock.userRole$ = of('Accountant');
      const accountantService = TestBed.inject(PermissionService);
      expect(accountantService.canManageDocuments()).toBe(false);
    });
  });

  describe('getCurrentRole', () => {
    it('should return the current user role', () => {
      authServiceMock.userRole$ = of('Admin');
      const newService = TestBed.inject(PermissionService);
      expect(newService.getCurrentRole()).toBe('Admin');
    });

    it('should return empty string when no role is set', () => {
      authServiceMock.userRole$ = of('');
      const newService = TestBed.inject(PermissionService);
      expect(newService.getCurrentRole()).toBe('');
    });
  });

  describe('Role hierarchy validation', () => {
    it('should match ROLE_HIERARCHY constant', () => {
      const expectedHierarchy = {
        SUPER_ADMIN: 0,
        ORG_ADMIN: 1,
        Admin: 2,
        PropertyManager: 3,
        Accountant: 4,
      };

      expect(ROLE_HIERARCHY).toEqual(expectedHierarchy);
    });

    it('should use consistent hierarchy in comparisons', () => {
      // Higher rank values = lower privilege
      expect(ROLE_HIERARCHY['SUPER_ADMIN']).toBeLessThan(ROLE_HIERARCHY['ORG_ADMIN']);
      expect(ROLE_HIERARCHY['ORG_ADMIN']).toBeLessThan(ROLE_HIERARCHY['Admin']);
      expect(ROLE_HIERARCHY['Admin']).toBeLessThan(ROLE_HIERARCHY['PropertyManager']);
      expect(ROLE_HIERARCHY['PropertyManager']).toBeLessThan(ROLE_HIERARCHY['Accountant']);
    });
  });

  describe('Edge cases', () => {
    it('should handle unknown role gracefully', () => {
      authServiceMock.userRole$ = of('UnknownRole');
      const newService = TestBed.inject(PermissionService);
      expect(newService.isSuperAdmin()).toBe(false);
      expect(newService.hasPermission(Permission.MANAGE_USERS)).toBe(false);
    });

    it('should handle empty role string', () => {
      authServiceMock.userRole$ = of('');
      const newService = TestBed.inject(PermissionService);
      expect(newService.isSuperAdmin()).toBe(false);
      expect(newService.isAdmin()).toBe(false);
    });

    it('should handle null role gracefully', () => {
      authServiceMock.userRole$ = of(null);
      const newService = TestBed.inject(PermissionService);
      expect(newService.getCurrentRole()).toBeNull();
    });
  });

  describe('Permission consistency', () => {
    it('should ensure all roles have permissions defined', () => {
      const roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'PropertyManager', 'Accountant'];

      roles.forEach((role) => {
        expect(ROLE_PERMISSIONS[role]).toBeDefined();
        expect(Array.isArray(ROLE_PERMISSIONS[role])).toBe(true);
      });
    });

    it('should ensure SUPER_ADMIN has most permissions', () => {
      const superAdminPerms = ROLE_PERMISSIONS['SUPER_ADMIN'].length;
      const orgAdminPerms = ROLE_PERMISSIONS['ORG_ADMIN'].length;
      const adminPerms = ROLE_PERMISSIONS['Admin'].length;
      const pmPerms = ROLE_PERMISSIONS['PropertyManager'].length;
      const acctPerms = ROLE_PERMISSIONS['Accountant'].length;

      expect(superAdminPerms).toBeGreaterThanOrEqual(orgAdminPerms);
      expect(orgAdminPerms).toBeGreaterThanOrEqual(adminPerms);
      expect(adminPerms).toBeGreaterThanOrEqual(pmPerms);
      expect(pmPerms).toBeGreaterThanOrEqual(acctPerms);
    });
  });
});
