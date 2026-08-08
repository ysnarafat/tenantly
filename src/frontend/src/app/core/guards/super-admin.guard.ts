import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { map, take } from 'rxjs/operators';
import { selectIsSuperAdmin } from '../../store/auth/auth.selectors';

export const superAdminGuard = () => {
  const store = inject(Store);
  const router = inject(Router);

  return store.select(selectIsSuperAdmin).pipe(
    take(1),
    map((isSuperAdmin) => {
      if (isSuperAdmin) {
        return true;
      }
      router.navigate(['/401']);
      return false;
    })
  );
};
