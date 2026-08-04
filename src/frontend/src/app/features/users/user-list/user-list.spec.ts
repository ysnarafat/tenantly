import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatSnackBar } from '@angular/material/snack-bar';
import { of, throwError } from 'rxjs';
import { UserList } from './user-list';
import { OrganizationService } from '../../../core/services/organization.service';
import { PermissionService } from '../../../core/services/permission.service';
import { User } from '../../../core/services/auth.service';

describe('UserList Component', () => {
  let component: UserList;
  let fixture: ComponentFixture<UserList>;
  let organizationService: jasmine.SpyObj<OrganizationService>;
  let permissionService: jasmine.SpyObj<PermissionService>;
  let snackBar: jasmine.SpyObj<MatSnackBar>;

  const mockUsers: User[] = [
    {
      id: 1,
      username: 'john_doe',
      email: 'john@example.com',
      role: 'Admin',
      active: true,
      created_at: '2026-04-01T10:00:00Z',
      updated_at: '2026-04-01T10:00:00Z',
    },
    {
      id: 2,
      username: 'jane_smith',
      email: 'jane@example.com',
      role: 'PropertyManager',
      active: true,
      created_at: '2026-04-01T10:00:00Z',
      updated_at: '2026-04-01T10:00:00Z',
    },
    {
      id: 3,
      username: 'bob_accountant',
      email: 'bob@example.com',
      role: 'Accountant',
      active: false,
      created_at: '2026-04-01T10:00:00Z',
      updated_at: '2026-04-01T10:00:00Z',
    },
  ];

  const mockOrganization = {
    id: 1,
    name: 'Test Org',
    slug: 'test-org',
    subscriptionTier: 'professional' as const,
    maxUsers: 50,
    active: true,
    createdAt: new Date(),
    updatedAt: new Date(),
  };

  beforeEach(async () => {
    const organizationServiceSpy = jasmine.createSpyObj('OrganizationService', [
      'getOrganizationUsers',
      'getCurrentOrganization',
    ]);
    const permissionServiceSpy = jasmine.createSpyObj('PermissionService', ['canManageOrgAdmins']);
    const snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);

    await TestBed.configureTestingModule({
      imports: [UserList],
      providers: [
        { provide: OrganizationService, useValue: organizationServiceSpy },
        { provide: PermissionService, useValue: permissionServiceSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
      ],
    }).compileComponents();

    organizationService = TestBed.inject(
      OrganizationService
    ) as jasmine.SpyObj<OrganizationService>;
    permissionService = TestBed.inject(PermissionService) as jasmine.SpyObj<PermissionService>;
    snackBar = TestBed.inject(MatSnackBar) as jasmine.SpyObj<MatSnackBar>;

    fixture = TestBed.createComponent(UserList);
    component = fixture.componentInstance;
  });

  describe('Initialization', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should load users on init', () => {
      organizationService.getCurrentOrganization.and.returnValue(mockOrganization);
      organizationService.getOrganizationUsers.and.returnValue(of({ users: mockUsers, total: 3 }));

      fixture.detectChanges();

      expect(organizationService.getOrganizationUsers).toHaveBeenCalledWith(1);
      expect(component.dataSource.data).toEqual(mockUsers);
    });

    it('should show error when no organization selected', () => {
      organizationService.getCurrentOrganization.and.returnValue(null);

      component.loadUsers();

      expect(snackBar.open).toHaveBeenCalledWith('No organization selected', 'Close', {
        duration: 3000,
      });
    });
  });

  describe('Data loading', () => {
    beforeEach(() => {
      organizationService.getCurrentOrganization.and.returnValue(mockOrganization);
    });

    it('should populate table with users', () => {
      organizationService.getOrganizationUsers.and.returnValue(of({ users: mockUsers, total: 3 }));

      component.loadUsers();

      expect(component.dataSource.data).toEqual(mockUsers);
      expect(component.dataSource.data.length).toBe(3);
    });

    it('should handle empty users list', () => {
      organizationService.getOrganizationUsers.and.returnValue(of({ users: [], total: 0 }));

      component.loadUsers();

      expect(component.dataSource.data).toEqual([]);
      expect(component.loading()).toBe(false);
    });

    it('should handle error when loading users', () => {
      organizationService.getOrganizationUsers.and.returnValue(
        throwError(() => new Error('Load failed'))
      );

      component.loadUsers();

      expect(snackBar.open).toHaveBeenCalledWith('Failed to load users', 'Close', {
        duration: 3000,
      });
    });
  });

  describe('Search functionality', () => {
    beforeEach(() => {
      organizationService.getCurrentOrganization.and.returnValue(mockOrganization);
      organizationService.getOrganizationUsers.and.returnValue(of({ users: mockUsers, total: 3 }));
      component.loadUsers();
    });

    it('should filter by username', () => {
      component.onSearchChange('john');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
      expect(filtered[0].username).toBe('john_doe');
    });

    it('should filter by email', () => {
      component.onSearchChange('jane@example.com');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
      expect(filtered[0].email).toBe('jane@example.com');
    });

    it('should be case insensitive', () => {
      component.onSearchChange('JOHN');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
    });

    it('should return all users when search is empty', () => {
      component.onSearchChange('');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(3);
    });

    it('should return empty array when no matches', () => {
      component.onSearchChange('nonexistent@example.com');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(0);
    });
  });

  describe('Role filtering', () => {
    beforeEach(() => {
      organizationService.getCurrentOrganization.and.returnValue(mockOrganization);
      organizationService.getOrganizationUsers.and.returnValue(of({ users: mockUsers, total: 3 }));
      component.loadUsers();
    });

    it('should filter by Admin role', () => {
      component.onRoleFilterChange('Admin');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
      expect(filtered[0].role).toBe('Admin');
    });

    it('should filter by PropertyManager role', () => {
      component.onRoleFilterChange('PropertyManager');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
      expect(filtered[0].role).toBe('PropertyManager');
    });

    it('should show all users when filter is cleared', () => {
      component.onRoleFilterChange(null);

      const filtered = component.filteredData();
      expect(filtered.length).toBe(3);
    });

    it('should return empty when filtering for non-existent role', () => {
      component.onRoleFilterChange('NonExistentRole');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(0);
    });
  });

  describe('Combined search and filter', () => {
    beforeEach(() => {
      organizationService.getCurrentOrganization.and.returnValue(mockOrganization);
      organizationService.getOrganizationUsers.and.returnValue(of({ users: mockUsers, total: 3 }));
      component.loadUsers();
    });

    it('should apply both search and role filter', () => {
      component.onSearchChange('jane');
      component.onRoleFilterChange('PropertyManager');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
      expect(filtered[0].username).toBe('jane_smith');
      expect(filtered[0].role).toBe('PropertyManager');
    });

    it('should return empty when search and filter have no intersection', () => {
      component.onSearchChange('john');
      component.onRoleFilterChange('PropertyManager');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(0);
    });
  });

  describe('User actions', () => {
    beforeEach(() => {
      organizationService.getCurrentOrganization.and.returnValue(mockOrganization);
      organizationService.getOrganizationUsers.and.returnValue(of({ users: mockUsers, total: 3 }));
      component.loadUsers();
    });

    it('should call deactivateUser', () => {
      spyOn(window, 'confirm').and.returnValue(true);
      const user = mockUsers[0];

      component.deactivateUser(user);

      expect(window.confirm).toHaveBeenCalled();
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should not deactivate when cancelled', () => {
      spyOn(window, 'confirm').and.returnValue(false);
      const user = mockUsers[0];
      snackBar.open.calls.reset();

      component.deactivateUser(user);

      expect(snackBar.open).not.toHaveBeenCalled();
    });

    it('should call deleteUser', () => {
      spyOn(window, 'confirm').and.returnValue(true);
      const user = mockUsers[0];

      component.deleteUser(user);

      expect(window.confirm).toHaveBeenCalled();
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should call promoteUser', () => {
      permissionService.canManageOrgAdmins.and.returnValue(true);
      const user = mockUsers[0];

      component.promoteUser(user);

      expect(snackBar.open).toHaveBeenCalled();
    });
  });

  describe('Role colors', () => {
    it('should return warn color for SUPER_ADMIN', () => {
      expect(component.getRoleColor('SUPER_ADMIN')).toBe('warn');
    });

    it('should return accent color for ORG_ADMIN', () => {
      expect(component.getRoleColor('ORG_ADMIN')).toBe('accent');
    });

    it('should return primary color for Admin', () => {
      expect(component.getRoleColor('Admin')).toBe('primary');
    });

    it('should return primary color for PropertyManager', () => {
      expect(component.getRoleColor('PropertyManager')).toBe('primary');
    });

    it('should return accent color for Accountant', () => {
      expect(component.getRoleColor('Accountant')).toBe('accent');
    });

    it('should return empty string for unknown role', () => {
      expect(component.getRoleColor('UnknownRole')).toBe('');
    });
  });

  describe('Permissions', () => {
    it('should check if user can promote users', () => {
      permissionService.canManageOrgAdmins.and.returnValue(true);

      expect(component.canPromoteUsers()).toBe(true);
    });

    it('should show promote button only when allowed', () => {
      permissionService.canManageOrgAdmins.and.returnValue(false);

      expect(component.canPromoteUsers()).toBe(false);
    });
  });

  describe('DisplayedColumns', () => {
    it('should have all required columns', () => {
      expect(component.displayedColumns).toContain('username');
      expect(component.displayedColumns).toContain('email');
      expect(component.displayedColumns).toContain('role');
      expect(component.displayedColumns).toContain('active');
      expect(component.displayedColumns).toContain('created_at');
      expect(component.displayedColumns).toContain('actions');
    });

    it('should have correct number of columns', () => {
      expect(component.displayedColumns.length).toBe(6);
    });
  });

  describe('User data display', () => {
    beforeEach(() => {
      organizationService.getCurrentOrganization.and.returnValue(mockOrganization);
      organizationService.getOrganizationUsers.and.returnValue(of({ users: mockUsers, total: 3 }));
      component.loadUsers();
    });

    it('should display user with correct properties', () => {
      const user = component.dataSource.data[0];
      expect(user.id).toBe(1);
      expect(user.username).toBe('john_doe');
      expect(user.email).toBe('john@example.com');
      expect(user.role).toBe('Admin');
    });

    it('should display inactive users', () => {
      const inactiveUser = component.dataSource.data.find((u) => !u.active);
      expect(inactiveUser).toBeDefined();
      expect(inactiveUser?.active).toBe(false);
    });

    it('should display first_name when available', () => {
      const userWithName = { ...mockUsers[0], first_name: 'John' };
      organizationService.getOrganizationUsers.and.returnValue(
        of({ users: [userWithName], total: 1 })
      );

      component.loadUsers();

      // The component uses first_name || username pattern
      expect(component.dataSource.data[0].first_name).toBe('John');
    });
  });

  describe('Error handling', () => {
    beforeEach(() => {
      organizationService.getCurrentOrganization.and.returnValue(mockOrganization);
    });

    it('should handle http errors when loading', () => {
      organizationService.getOrganizationUsers.and.returnValue(
        throwError(() => ({ status: 500, statusText: 'Server Error' }))
      );

      component.loadUsers();

      expect(snackBar.open).toHaveBeenCalledWith('Failed to load users', 'Close', {
        duration: 3000,
      });
    });

    it('should handle 403 forbidden error', () => {
      organizationService.getOrganizationUsers.and.returnValue(
        throwError(() => ({ status: 403, statusText: 'Forbidden' }))
      );

      component.loadUsers();

      expect(snackBar.open).toHaveBeenCalled();
    });
  });

  describe('Roles list', () => {
    it('should include all available roles', () => {
      expect(component.roles).toContain('SUPER_ADMIN');
      expect(component.roles).toContain('ORG_ADMIN');
      expect(component.roles).toContain('Admin');
      expect(component.roles).toContain('PropertyManager');
      expect(component.roles).toContain('Accountant');
    });

    it('should have correct number of roles', () => {
      expect(component.roles.length).toBe(5);
    });
  });
});
