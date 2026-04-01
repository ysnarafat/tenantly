import { Injectable, inject, signal, computed } from '@angular/core';
import { AuthService } from './auth.service';
import {
  Permission,
  hasPermission,
  hasAnyPermission,
  hasAllPermissions,
  ROLE_HIERARCHY,
  ROLE_PERMISSIONS,
  UserRole,
} from '../models/role.model';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Injectable({
  providedIn: 'root',
})
export class PermissionService {
  private authService = inject(AuthService);

  // Current user role as signal
  private userRole = signal<string>('');

  constructor() {
    // Subscribe to role changes
    this.authService.userRole$
      .pipe(takeUntilDestroyed())
      .subscribe((role) => this.userRole.set(role));
  }

  /**
   * Check if current user has a specific permission
   */
  hasPermission(permission: Permission): boolean {
    return hasPermission(this.userRole() as UserRole, permission);
  }

  /**
   * Check if current user has any of the specified permissions
   */
  hasAnyPermission(permissions: Permission[]): boolean {
    return hasAnyPermission(this.userRole() as UserRole, permissions);
  }

  /**
   * Check if current user has all of the specified permissions
   */
  hasAllPermissions(permissions: Permission[]): boolean {
    return hasAllPermissions(this.userRole() as UserRole, permissions);
  }

  /**
   * Computed signal for checking permission
   */
  canAccess(permission: Permission) {
    return computed(() => hasPermission(this.userRole() as UserRole, permission));
  }

  /**
   * Check if user is SUPER_ADMIN
   */
  isSuperAdmin(): boolean {
    return this.userRole() === 'SUPER_ADMIN';
  }

  /**
   * Check if user is ORG_ADMIN
   */
  isOrgAdmin(): boolean {
    return this.userRole() === 'ORG_ADMIN';
  }

  /**
   * Check if user is Admin
   */
  isAdmin(): boolean {
    return this.userRole() === 'Admin';
  }

  /**
   * Check if user is Property Manager
   */
  isPropertyManager(): boolean {
    return this.userRole() === 'PropertyManager';
  }

  /**
   * Check if user is Accountant
   */
  isAccountant(): boolean {
    return this.userRole() === 'Accountant';
  }

  /**
   * Check if user can manage properties (Admin or PropertyManager)
   */
  canManageProperties(): boolean {
    return this.hasPermission(Permission.MANAGE_PROPERTIES);
  }

  /**
   * Check if user can manage tenants (Admin or PropertyManager)
   */
  canManageTenants(): boolean {
    return this.hasPermission(Permission.MANAGE_TENANTS);
  }

  /**
   * Check if user can record payments
   */
  canRecordPayments(): boolean {
    return this.hasPermission(Permission.RECORD_PAYMENTS);
  }

  /**
   * Check if user can view reports
   */
  canViewReports(): boolean {
    return this.hasPermission(Permission.VIEW_REPORTS);
  }

  /**
   * Check if user can manage documents
   */
  canManageDocuments(): boolean {
    return this.hasPermission(Permission.MANAGE_DOCUMENTS);
  }

  /**
   * Check if user can manage organizations (SUPER_ADMIN only)
   */
  canManageOrganizations(): boolean {
    return this.hasPermission(Permission.MANAGE_ORGANIZATIONS);
  }

  /**
   * Check if user can manage org admins
   */
  canManageOrgAdmins(): boolean {
    return this.hasPermission(Permission.MANAGE_ORG_ADMINS);
  }

  /**
   * Check if user can invite users
   */
  canInviteUsers(): boolean {
    return this.hasPermission(Permission.INVITE_USERS);
  }

  /**
   * Check if one role is higher than another in the hierarchy
   * (lower number = higher privilege)
   */
  isRoleHigherThan(role1: UserRole, role2: UserRole): boolean {
    const rank1 = ROLE_HIERARCHY[role1];
    const rank2 = ROLE_HIERARCHY[role2];
    return rank1 < rank2;
  }

  /**
   * Get all permissions for a role
   */
  getRolePermissions(role: UserRole): Permission[] {
    return ROLE_PERMISSIONS[role] || [];
  }

  /**
   * Get current user role
   */
  getCurrentRole(): string {
    return this.userRole();
  }
}
