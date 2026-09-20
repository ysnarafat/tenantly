import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatDialog, MatDialogRef } from '@angular/material/dialog';
import { of, throwError, Subject } from 'rxjs';
import { UserList } from './user-list';
import { UserService } from '../../../core/services/user.service';
import { PermissionService } from '../../../core/services/permission.service';
import { User } from '../../../core/services/auth.service';

// The activate/deactivate/reset-password confirm dialogs are real
// MatDialogRef instances in the component; stub just the afterClosed()
// result the component reads.
function dialogRefStub(result: unknown): MatDialogRef<unknown> {
  return { afterClosed: () => of(result) } as unknown as MatDialogRef<unknown>;
}

// actWithUndo drives a real MatSnackBarRef's onAction()/afterDismissed() —
// this stub lets tests simulate "undo clicked" vs. "toast timed out" without
// a real timer.
function createSnackBarRefStub() {
  const action = new Subject<void>();
  const dismissed = new Subject<{ dismissedByAction: boolean }>();
  return {
    ref: { onAction: () => action.asObservable(), afterDismissed: () => dismissed.asObservable() },
    clickUndo: () => {
      action.next();
      dismissed.next({ dismissedByAction: true });
    },
    timeOut: () => dismissed.next({ dismissedByAction: false }),
  };
}

describe('UserList Component', () => {
  let component: UserList;
  let fixture: ComponentFixture<UserList>;
  let userService: jasmine.SpyObj<UserService>;
  let permissionService: jasmine.SpyObj<PermissionService>;
  let router: jasmine.SpyObj<Router>;
  let snackBar: jasmine.SpyObj<MatSnackBar>;
  let dialog: jasmine.SpyObj<MatDialog>;

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

  beforeEach(async () => {
    const userServiceSpy = jasmine.createSpyObj('UserService', [
      'getUsers',
      'updateUser',
      'deleteUser',
    ]);
    const permissionServiceSpy = jasmine.createSpyObj('PermissionService', ['canManageOrgAdmins']);
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);
    const snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);
    const dialogSpy = jasmine.createSpyObj('MatDialog', ['open']);

    await TestBed.configureTestingModule({
      imports: [UserList],
      providers: [
        { provide: UserService, useValue: userServiceSpy },
        { provide: PermissionService, useValue: permissionServiceSpy },
        { provide: Router, useValue: routerSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
        { provide: MatDialog, useValue: dialogSpy },
      ],
    }).compileComponents();

    userService = TestBed.inject(UserService) as jasmine.SpyObj<UserService>;
    permissionService = TestBed.inject(PermissionService) as jasmine.SpyObj<PermissionService>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;
    snackBar = TestBed.inject(MatSnackBar) as jasmine.SpyObj<MatSnackBar>;
    dialog = TestBed.inject(MatDialog) as jasmine.SpyObj<MatDialog>;

    userService.getUsers.and.returnValue(of({ users: mockUsers, total: 3 }));

    fixture = TestBed.createComponent(UserList);
    component = fixture.componentInstance;
  });

  describe('Initialization', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should load users on init', () => {
      fixture.detectChanges();

      expect(userService.getUsers).toHaveBeenCalledWith(true);
      expect(component.dataSource.data).toEqual(mockUsers);
    });

    it('should handle error when loading users', () => {
      userService.getUsers.and.returnValue(throwError(() => new Error('Load failed')));

      component.loadUsers();

      expect(snackBar.open).toHaveBeenCalledWith(
        'Failed to load users',
        'Close',
        jasmine.objectContaining({ duration: 5000 })
      );
      expect(component.loading()).toBe(false);
    });
  });

  describe('Data loading', () => {
    it('should populate table with users', () => {
      component.loadUsers();

      expect(component.dataSource.data).toEqual(mockUsers);
      expect(component.dataSource.data.length).toBe(3);
    });

    it('should handle empty users list', () => {
      userService.getUsers.and.returnValue(of({ users: [], total: 0 }));

      component.loadUsers();

      expect(component.dataSource.data).toEqual([]);
      expect(component.loading()).toBe(false);
    });
  });

  describe('showInactive toggle', () => {
    it('should request all users (including inactive) when toggled on', () => {
      component.toggleShowInactive(true);

      expect(userService.getUsers).toHaveBeenCalledWith(false);
    });

    it('should request only active users when toggled off', () => {
      component.toggleShowInactive(false);

      expect(userService.getUsers).toHaveBeenCalledWith(true);
    });
  });

  describe('Search functionality', () => {
    beforeEach(() => {
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
      component.loadUsers();
    });

    it('should ask for confirmation and deactivate the user when confirmed', () => {
      dialog.open.and.returnValue(dialogRefStub(true));
      userService.updateUser.and.returnValue(of({ message: 'ok' }));
      userService.getUsers.calls.reset();
      userService.getUsers.and.returnValue(of({ users: mockUsers, total: 3 }));
      const user = mockUsers[0];

      component.deactivateUser(user);

      expect(dialog.open).toHaveBeenCalled();
      expect(userService.updateUser).toHaveBeenCalledWith(user.id, { active: false });
      expect(snackBar.open).toHaveBeenCalledWith(
        `${user.username} deactivated`,
        'Close',
        jasmine.objectContaining({ duration: 3000 })
      );
    });

    it('should not deactivate the user when the confirm dialog is cancelled', () => {
      dialog.open.and.returnValue(dialogRefStub(false));
      const user = mockUsers[0];

      component.deactivateUser(user);

      expect(userService.updateUser).not.toHaveBeenCalled();
    });

    it('should show an error and not touch local state when deactivate fails', () => {
      dialog.open.and.returnValue(dialogRefStub(true));
      userService.updateUser.and.returnValue(
        throwError(() => ({ error: { error: 'Failed to deactivate user' } }))
      );
      const user = mockUsers[0];

      component.deactivateUser(user);

      expect(snackBar.open).toHaveBeenCalledWith(
        'Failed to deactivate user',
        'Close',
        jasmine.objectContaining({ duration: 5000 })
      );
    });

    it('should ask for confirmation and activate the user when confirmed', () => {
      dialog.open.and.returnValue(dialogRefStub(true));
      userService.updateUser.and.returnValue(of({ message: 'ok' }));
      const user = mockUsers[2];

      component.activateUser(user);

      expect(dialog.open).toHaveBeenCalled();
      expect(userService.updateUser).toHaveBeenCalledWith(user.id, { active: true });
      expect(component.allUsers().find((u) => u.id === user.id)?.active).toBe(true);
      expect(snackBar.open).toHaveBeenCalledWith(
        `${user.username} activated`,
        'Close',
        jasmine.objectContaining({ duration: 3000 })
      );
    });

    it('should not activate the user when the confirm dialog is cancelled', () => {
      dialog.open.and.returnValue(dialogRefStub(false));
      const user = mockUsers[2];

      component.activateUser(user);

      expect(userService.updateUser).not.toHaveBeenCalled();
    });

    it('should optimistically remove the user and show an undo toast on delete', () => {
      const stub = createSnackBarRefStub();
      snackBar.open.and.returnValue(stub.ref as never);
      const user = mockUsers[0];

      component.deleteUser(user);

      expect(component.allUsers().find((u) => u.id === user.id)).toBeUndefined();
      expect(snackBar.open).toHaveBeenCalledWith(
        `${user.username} deleted`,
        'Undo',
        jasmine.objectContaining({ duration: 5000 })
      );
      expect(userService.deleteUser).not.toHaveBeenCalled();
    });

    it('should call deleteUser once the delete toast times out', () => {
      const stub = createSnackBarRefStub();
      snackBar.open.and.returnValue(stub.ref as never);
      userService.deleteUser.and.returnValue(of({ message: 'ok' }));
      const user = mockUsers[0];

      component.deleteUser(user);
      stub.timeOut();

      expect(userService.deleteUser).toHaveBeenCalledWith(user.id);
    });

    it('should restore the user and skip deleteUser when Undo is clicked on delete', () => {
      const stub = createSnackBarRefStub();
      snackBar.open.and.returnValue(stub.ref as never);
      const user = mockUsers[0];

      component.deleteUser(user);
      stub.clickUndo();

      expect(component.allUsers().find((u) => u.id === user.id)).toEqual(user);
      expect(userService.deleteUser).not.toHaveBeenCalled();
    });

    it('should navigate to the promote page', () => {
      permissionService.canManageOrgAdmins.and.returnValue(true);
      const user = mockUsers[0];

      component.promoteUser(user);

      expect(router.navigate).toHaveBeenCalledWith(['/admin/users/promote'], {
        queryParams: { userId: user.id },
      });
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
  });

  describe('Error handling', () => {
    it('should handle http errors when loading', () => {
      userService.getUsers.and.returnValue(
        throwError(() => ({ status: 500, statusText: 'Server Error' }))
      );

      component.loadUsers();

      expect(snackBar.open).toHaveBeenCalledWith(
        'Failed to load users',
        'Close',
        jasmine.objectContaining({ duration: 5000 })
      );
    });

    it('should handle 403 forbidden error', () => {
      userService.getUsers.and.returnValue(
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
