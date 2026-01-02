import { Routes } from '@angular/router';
import { AuthGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  {
    path: '',
    redirectTo: '/dashboard',
    pathMatch: 'full',
  },
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login/login').then((m) => m.Login),
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
    canActivate: [AuthGuard],
  },
  {
    path: 'tenants',
    loadComponent: () =>
      import('./features/tenants/tenant-list/tenant-list').then((m) => m.TenantList),
    canActivate: [AuthGuard],
  },
  {
    path: 'leases',
    loadComponent: () => import('./features/leases/lease-list/lease-list').then((m) => m.LeaseList),
    canActivate: [AuthGuard],
  },
  {
    path: 'payments',
    loadComponent: () =>
      import('./features/payments/payment-list/payment-list').then((m) => m.PaymentList),
    canActivate: [AuthGuard],
  },
  {
    path: 'reports',
    loadComponent: () =>
      import('./features/reports/report-list/report-list').then((m) => m.ReportList),
    canActivate: [AuthGuard],
  },
  {
    path: 'documents',
    loadComponent: () =>
      import('./features/attachments/attachment-list/attachment-list').then(
        (m) => m.AttachmentList
      ),
    canActivate: [AuthGuard],
  },
  {
    path: 'users',
    loadComponent: () => import('./features/users/user-list/user-list').then((m) => m.UserList),
    canActivate: [AuthGuard],
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
    path: '**',
    redirectTo: '/dashboard',
  },
];
