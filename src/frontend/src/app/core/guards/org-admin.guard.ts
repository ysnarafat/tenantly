import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { map, take } from 'rxjs/operators';
import { selectIsSuperAdmin, selectIsOrgAdmin } from '../../store/auth/auth.selectors';
import { combineLatest } from 'rxjs';

export const orgAdminGuard = () => {
  const store = inject(Store);
  const router = inject(Router);

  return combineLatest([store.select(selectIsSuperAdmin), store.select(selectIsOrgAdmin)]).pipe(
    take(1),
    map(([isSuperAdmin, isOrgAdmin]) => {
      if (isSuperAdmin || isOrgAdmin) {
        return true;
      }
      router.navigate(['/dashboard']);
      return false;
    })
  );
};
