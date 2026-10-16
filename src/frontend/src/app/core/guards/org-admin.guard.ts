import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { map, take } from 'rxjs/operators';
import { selectUserRole } from '../../store/auth/auth.selectors';

export const orgAdminGuard = () => {
  const store = inject(Store);
  const router = inject(Router);

  return store.select(selectUserRole).pipe(
    take(1),
    map((role) => {
      if (role === 'SUPER_ADMIN' || role === 'ORG_ADMIN' || role === 'Admin') {
        return true;
      }
      router.navigate(['/unauthorized']);
      return false;
    })
  );
};
