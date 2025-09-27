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
    path: 'shops',
    loadComponent: () => import('./features/shops/shop-list/shop-list').then((m) => m.ShopList),
    canActivate: [AuthGuard],
  },
  {
    path: 'tenants',
    loadComponent: () =>
      import('./features/tenants/tenant-list/tenant-list').then((m) => m.TenantList),
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
    path: 'users',
    loadComponent: () => import('./features/users/user-list/user-list').then((m) => m.UserList),
    canActivate: [AuthGuard],
  },
  {
    path: '**',
    redirectTo: '/dashboard',
  },
];
