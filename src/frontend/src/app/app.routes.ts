import { Routes } from '@angular/router';
import { AuthGuard } from './core/guards/auth.guard';
import { GuestGuard } from './core/guards/guest.guard';
import { superAdminGuard } from './core/guards/super-admin.guard';
import { orgAdminGuard } from './core/guards/org-admin.guard';
import { userManagementGuard } from './core/guards/user-management.guard';
import { permissionGuard } from './core/guards/permission.guard';
import { Permission } from './core/models/role.model';
import { provideState } from '@ngrx/store';
import { provideEffects } from '@ngrx/effects';
import { propertyReducer } from './features/properties/store/property.reducer';
import { buildingReducer } from './features/properties/store/building.reducer';
import { unitReducer } from './features/properties/store/unit.reducer';
import { PropertyEffects } from './features/properties/store/property.effects';
import { BuildingEffects } from './features/properties/store/building.effects';
import { UnitEffects } from './features/properties/store/unit.effects';

export const routes: Routes = [
  {
    path: '',
    redirectTo: '/home',
    pathMatch: 'full',
  },
  {
    path: 'home',
    loadComponent: () => import('./features/marketing/homepage/homepage').then((m) => m.Homepage),
    canActivate: [GuestGuard],
  },
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login/login').then((m) => m.Login),
    canActivate: [GuestGuard],
  },
  {
    path: 'select-organization',
    loadComponent: () =>
      import('./features/auth/organization-picker/organization-picker').then(
        (m) => m.OrganizationPicker
      ),
    canActivate: [AuthGuard],
  },
  {
    path: 'dashboard',
    loadComponent: () => import('./features/dashboard/dashboard').then((m) => m.Dashboard),
    canActivate: [AuthGuard],
  },
  {
    path: 'properties',
    loadComponent: () =>
      import('./features/properties/property-list/property-list.component').then(
        (m) => m.PropertyListComponent
      ),
    canActivate: [AuthGuard, permissionGuard(Permission.MANAGE_PROPERTIES)],
    providers: [
      provideState('properties', propertyReducer),
      provideState('buildings', buildingReducer),
      provideState('units', unitReducer),
      provideEffects([PropertyEffects, BuildingEffects, UnitEffects]),
    ],
  },
  {
    path: 'tenants',
    loadComponent: () =>
      import('./features/tenants/tenant-list/tenant-list').then((m) => m.TenantList),
    canActivate: [AuthGuard, permissionGuard(Permission.MANAGE_TENANTS)],
  },
  {
    path: 'leases',
    loadComponent: () => import('./features/leases/lease-list/lease-list').then((m) => m.LeaseList),
    canActivate: [AuthGuard, permissionGuard(Permission.MANAGE_TENANTS)],
  },
  {
    path: 'leases/due',
    loadComponent: () => import('./features/leases/due-list/due-list').then((m) => m.DueList),
    canActivate: [AuthGuard, permissionGuard(Permission.MANAGE_TENANTS)],
  },
  {
    path: 'payments',
    loadComponent: () =>
      import('./features/payments/payment-list/payment-list').then((m) => m.PaymentList),
    canActivate: [AuthGuard, permissionGuard(Permission.VIEW_PAYMENTS)],
  },
  {
    path: 'reports',
    loadComponent: () => import('./features/reports/report-analysis').then((m) => m.ReportAnalysis),
    canActivate: [AuthGuard, permissionGuard(Permission.VIEW_REPORTS)],
  },
  {
    path: 'documents',
    loadComponent: () =>
      import('./features/attachments/attachment-list/attachment-list').then(
        (m) => m.AttachmentList
      ),
    canActivate: [AuthGuard, permissionGuard(Permission.MANAGE_DOCUMENTS)],
  },
  {
    path: 'users',
    loadComponent: () => import('./features/users/user-list/user-list').then((m) => m.UserList),
    canActivate: [AuthGuard, userManagementGuard],
  },
  {
    path: 'admin',
    canActivate: [AuthGuard, orgAdminGuard],
    children: [
      {
        path: 'organizations',
        loadComponent: () =>
          import('./features/admin/organization-management/organization-list').then(
            (m) => m.OrganizationListComponent
          ),
        canActivate: [superAdminGuard],
      },
      {
        path: 'organizations/new',
        loadComponent: () =>
          import('./features/admin/organization-management/organization-create').then(
            (m) => m.OrganizationCreateComponent
          ),
        canActivate: [superAdminGuard],
      },
      {
        path: 'invitations',
        loadComponent: () =>
          import('./features/admin/user-onboarding/pending-invitations').then(
            (m) => m.PendingInvitationsComponent
          ),
      },
      {
        path: 'invitations/new',
        loadComponent: () =>
          import('./features/admin/user-onboarding/invite-user').then((m) => m.InviteUserComponent),
      },
      {
        path: 'users',
        loadComponent: () => import('./features/users/user-list/user-list').then((m) => m.UserList),
        canActivate: [userManagementGuard],
      },
      {
        path: 'users/new',
        loadComponent: () =>
          import('./features/admin/user-management/create-user').then((m) => m.CreateUserComponent),
        canActivate: [userManagementGuard],
      },
    ],
  },
  // Legacy routes for backward compatibility
  {
    path: 'shops',
    redirectTo: 'properties',
    pathMatch: 'full',
  },
  {
    path: 'attachments',
    redirectTo: 'documents',
    pathMatch: 'full',
  },
  {
    path: '401',
    loadComponent: () =>
      import('./features/errors/unauthorized/unauthorized').then((m) => m.Unauthorized),
  },
  {
    path: '404',
    loadComponent: () => import('./features/errors/not-found/not-found').then((m) => m.NotFound),
  },
  {
    path: '**',
    redirectTo: '/404',
  },
];
