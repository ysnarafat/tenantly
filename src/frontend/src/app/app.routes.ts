import { Routes } from '@angular/router';
import { AuthGuard } from './core/guards/auth.guard';
import { GuestGuard } from './core/guards/guest.guard';
import { superAdminGuard } from './core/guards/super-admin.guard';
import { orgAdminGuard } from './core/guards/org-admin.guard';
import { userManagementGuard } from './core/guards/user-management.guard';
import { provideState } from '@ngrx/store';
import { provideEffects } from '@ngrx/effects';
import { propertyReducer } from './features/properties/store/property.reducer';
import { buildingReducer } from './features/properties/store/building.reducer';
import { unitReducer } from './features/properties/store/unit.reducer';
import { PropertyEffects } from './features/properties/store/property.effects';
import { BuildingEffects } from './features/properties/store/building.effects';
import { UnitEffects } from './features/properties/store/unit.effects';

// `title` sets the document title via AppTitleStrategy (see core/seo). Kept as
// plain strings so the tab title is correct on first paint, independent of the
// app's async translations.
export const routes: Routes = [
  {
    path: '',
    redirectTo: '/dashboard',
    pathMatch: 'full',
  },
  {
    path: 'login',
    title: 'Sign in',
    loadComponent: () => import('./features/auth/login/login').then((m) => m.Login),
    canActivate: [GuestGuard],
  },
  {
    path: 'select-organization',
    title: 'Select organization',
    loadComponent: () =>
      import('./features/auth/organization-picker/organization-picker').then(
        (m) => m.OrganizationPicker
      ),
    canActivate: [AuthGuard],
  },
  {
    path: 'dashboard',
    title: 'Dashboard',
    loadComponent: () => import('./features/dashboard/dashboard').then((m) => m.Dashboard),
    canActivate: [AuthGuard],
  },
  {
    path: 'properties',
    title: 'Properties',
    loadComponent: () =>
      import('./features/properties/property-list/property-list.component').then(
        (m) => m.PropertyListComponent
      ),
    canActivate: [AuthGuard],
    providers: [
      provideState('properties', propertyReducer),
      provideState('buildings', buildingReducer),
      provideState('units', unitReducer),
      provideEffects([PropertyEffects, BuildingEffects, UnitEffects]),
    ],
  },
  {
    path: 'tenants',
    title: 'Tenants',
    loadComponent: () =>
      import('./features/tenants/tenant-list/tenant-list').then((m) => m.TenantList),
    canActivate: [AuthGuard],
  },
  {
    path: 'leases',
    title: 'Lease Management',
    loadComponent: () => import('./features/leases/lease-list/lease-list').then((m) => m.LeaseList),
    canActivate: [AuthGuard],
  },
  {
    path: 'leases/due',
    title: 'Rent Due',
    loadComponent: () => import('./features/leases/due-list/due-list').then((m) => m.DueList),
    canActivate: [AuthGuard],
  },
  {
    path: 'payments',
    title: 'Payments',
    loadComponent: () =>
      import('./features/payments/payment-list/payment-list').then((m) => m.PaymentList),
    canActivate: [AuthGuard],
  },
  {
    path: 'reports',
    title: 'Reports & Analysis',
    loadComponent: () => import('./features/reports/report-analysis').then((m) => m.ReportAnalysis),
    canActivate: [AuthGuard],
  },
  {
    path: 'documents',
    title: 'Documents',
    loadComponent: () =>
      import('./features/attachments/attachment-list/attachment-list').then(
        (m) => m.AttachmentList
      ),
    canActivate: [AuthGuard],
  },
  {
    path: 'users',
    title: 'User Management',
    loadComponent: () => import('./features/users/user-list/user-list').then((m) => m.UserList),
    canActivate: [AuthGuard, userManagementGuard],
  },
  {
    path: 'admin',
    canActivate: [AuthGuard, orgAdminGuard],
    children: [
      {
        path: 'organizations',
        title: 'Organizations',
        loadComponent: () =>
          import('./features/admin/organization-management/organization-list').then(
            (m) => m.OrganizationListComponent
          ),
        canActivate: [superAdminGuard],
      },
      {
        path: 'organizations/new',
        title: 'New organization',
        loadComponent: () =>
          import('./features/admin/organization-management/organization-create').then(
            (m) => m.OrganizationCreateComponent
          ),
        canActivate: [superAdminGuard],
      },
      {
        path: 'invitations',
        title: 'Invitations',
        loadComponent: () =>
          import('./features/admin/user-onboarding/pending-invitations').then(
            (m) => m.PendingInvitationsComponent
          ),
      },
      {
        path: 'invitations/new',
        title: 'Invite user',
        loadComponent: () =>
          import('./features/admin/user-onboarding/invite-user').then((m) => m.InviteUserComponent),
      },
      {
        path: 'users',
        title: 'User Management',
        loadComponent: () => import('./features/users/user-list/user-list').then((m) => m.UserList),
        canActivate: [userManagementGuard],
      },
      {
        path: 'users/new',
        title: 'New user',
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
  // 403 — reached when a role guard blocks access
  {
    path: 'unauthorized',
    title: 'Access denied',
    loadComponent: () =>
      import('./features/errors/unauthorized/unauthorized').then((m) => m.Unauthorized),
    canActivate: [AuthGuard],
  },
  // 404 — render in place so the mistyped address is preserved in the URL bar
  {
    path: '**',
    title: 'Page not found',
    loadComponent: () => import('./features/errors/not-found/not-found').then((m) => m.NotFound),
  },
];
