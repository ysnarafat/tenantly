import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { map, take } from 'rxjs/operators';
import { selectUserRole } from '../../store/auth/auth.selectors';
import { Permission, UserRole, hasPermission } from '../models/role.model';

// Blocks direct-URL access to a route the caller's role lacks permission for
// (the same permission that gates the matching sidenav link in app.ts), and
// sends them to the 401 page instead of letting them load data-bearing views
// the nav intentionally hides for their role.
export const permissionGuard = (permission: Permission) => () => {
  const store = inject(Store);
  const router = inject(Router);

  return store.select(selectUserRole).pipe(
    take(1),
    map((role) => {
      if (hasPermission(role as UserRole, permission)) {
        return true;
      }
      router.navigate(['/401']);
      return false;
    })
  );
};
