// Role types
export type UserRole = 'SUPER_ADMIN' | 'ORG_ADMIN' | 'Admin' | 'PropertyManager' | 'Accountant';

// Role hierarchy (lower number = higher privilege)
export const ROLE_HIERARCHY: Record<UserRole, number> = {
  SUPER_ADMIN: 0,
  ORG_ADMIN: 1,
  Admin: 2,
  PropertyManager: 3,
  Accountant: 4,
};

// Permission types
export enum Permission {
  // Organization Management
  MANAGE_ORGANIZATIONS = 'manage_organizations',
  MANAGE_ORG_ADMINS = 'manage_org_admins',

  // User Management
  MANAGE_USERS = 'manage_users',
  VIEW_USERS = 'view_users',
  INVITE_USERS = 'invite_users',

  // Property Management
  MANAGE_PROPERTIES = 'manage_properties',
  VIEW_PROPERTIES = 'view_properties',
  MANAGE_BUILDINGS = 'manage_buildings',
  MANAGE_UNITS = 'manage_units',

  // Tenant Management
  MANAGE_TENANTS = 'manage_tenants',
  VIEW_TENANTS = 'view_tenants',
  MANAGE_LEASES = 'manage_leases',
  VIEW_LEASES = 'view_leases',

  // Financial
  RECORD_PAYMENTS = 'record_payments',
  VIEW_PAYMENTS = 'view_payments',
  VIEW_REPORTS = 'view_reports',
  EXPORT_REPORTS = 'export_reports',

  // Documents
  MANAGE_DOCUMENTS = 'manage_documents',
  VIEW_DOCUMENTS = 'view_documents',

  // System
  VIEW_DASHBOARD = 'view_dashboard',
  MANAGE_SETTINGS = 'manage_settings',
}

// Role-Permission mapping
export const ROLE_PERMISSIONS: Record<string, Permission[]> = {
  SUPER_ADMIN: [
    // Organization
    Permission.MANAGE_ORGANIZATIONS,
    Permission.MANAGE_ORG_ADMINS,
    // Users
    Permission.MANAGE_USERS,
    Permission.VIEW_USERS,
    Permission.INVITE_USERS,
    // Properties
    Permission.MANAGE_PROPERTIES,
    Permission.VIEW_PROPERTIES,
    Permission.MANAGE_BUILDINGS,
    Permission.MANAGE_UNITS,
    // Tenants
    Permission.MANAGE_TENANTS,
    Permission.VIEW_TENANTS,
    Permission.MANAGE_LEASES,
    Permission.VIEW_LEASES,
    // Financial
    Permission.RECORD_PAYMENTS,
    Permission.VIEW_PAYMENTS,
    Permission.VIEW_REPORTS,
    Permission.EXPORT_REPORTS,
    // Documents
    Permission.MANAGE_DOCUMENTS,
    Permission.VIEW_DOCUMENTS,
    // System
    Permission.VIEW_DASHBOARD,
    Permission.MANAGE_SETTINGS,
  ],
  ORG_ADMIN: [
    // Users (no organization management)
    Permission.MANAGE_USERS,
    Permission.VIEW_USERS,
    Permission.INVITE_USERS,
    Permission.MANAGE_ORG_ADMINS,
    // Properties
    Permission.MANAGE_PROPERTIES,
    Permission.VIEW_PROPERTIES,
    Permission.MANAGE_BUILDINGS,
    Permission.MANAGE_UNITS,
    // Tenants
    Permission.MANAGE_TENANTS,
    Permission.VIEW_TENANTS,
    Permission.MANAGE_LEASES,
    Permission.VIEW_LEASES,
    // Financial
    Permission.RECORD_PAYMENTS,
    Permission.VIEW_PAYMENTS,
    Permission.VIEW_REPORTS,
    Permission.EXPORT_REPORTS,
    // Documents
    Permission.MANAGE_DOCUMENTS,
    Permission.VIEW_DOCUMENTS,
    // System
    Permission.VIEW_DASHBOARD,
    Permission.MANAGE_SETTINGS,
  ],
  Admin: [
    Permission.MANAGE_USERS,
    Permission.VIEW_USERS,
    Permission.INVITE_USERS,
    Permission.MANAGE_PROPERTIES,
    Permission.VIEW_PROPERTIES,
    Permission.MANAGE_BUILDINGS,
    Permission.MANAGE_UNITS,
    Permission.MANAGE_TENANTS,
    Permission.VIEW_TENANTS,
    Permission.MANAGE_LEASES,
    Permission.VIEW_LEASES,
    Permission.RECORD_PAYMENTS,
    Permission.VIEW_PAYMENTS,
    Permission.VIEW_REPORTS,
    Permission.EXPORT_REPORTS,
    Permission.MANAGE_DOCUMENTS,
    Permission.VIEW_DOCUMENTS,
    Permission.VIEW_DASHBOARD,
    Permission.MANAGE_SETTINGS,
  ],
  PropertyManager: [
    Permission.VIEW_USERS,
    Permission.VIEW_PROPERTIES,
    Permission.MANAGE_PROPERTIES,
    Permission.MANAGE_BUILDINGS,
    Permission.MANAGE_UNITS,
    Permission.MANAGE_TENANTS,
    Permission.VIEW_TENANTS,
    Permission.MANAGE_LEASES,
    Permission.VIEW_LEASES,
    Permission.RECORD_PAYMENTS,
    Permission.VIEW_PAYMENTS,
    Permission.VIEW_REPORTS,
    Permission.MANAGE_DOCUMENTS,
    Permission.VIEW_DOCUMENTS,
    Permission.VIEW_DASHBOARD,
  ],
  Accountant: [
    Permission.VIEW_USERS,
    Permission.VIEW_PROPERTIES,
    Permission.VIEW_TENANTS,
    Permission.VIEW_LEASES,
    Permission.VIEW_PAYMENTS,
    Permission.VIEW_REPORTS,
    Permission.EXPORT_REPORTS,
    Permission.VIEW_DOCUMENTS,
    Permission.VIEW_DASHBOARD,
  ],
};

// Helper function to check if role has permission
export function hasPermission(role: UserRole, permission: Permission): boolean {
  const permissions = ROLE_PERMISSIONS[role] || [];
  return permissions.includes(permission);
}

// Helper function to check if role has any of the permissions
export function hasAnyPermission(role: UserRole, permissions: Permission[]): boolean {
  return permissions.some((permission) => hasPermission(role, permission));
}

// Helper function to check if role has all permissions
export function hasAllPermissions(role: UserRole, permissions: Permission[]): boolean {
  return permissions.every((permission) => hasPermission(role, permission));
}
