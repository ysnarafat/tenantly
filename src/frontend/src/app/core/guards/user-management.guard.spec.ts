import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { of } from 'rxjs';
import { userManagementGuard } from './user-management.guard';

describe('userManagementGuard', () => {
  let store: jasmine.SpyObj<Store>;
  let router: jasmine.SpyObj<Router>;

  beforeEach(() => {
    const storeSpy = jasmine.createSpyObj('Store', ['select']);
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    TestBed.configureTestingModule({
      providers: [
        { provide: Store, useValue: storeSpy },
        { provide: Router, useValue: routerSpy },
      ],
    });

    store = TestBed.inject(Store) as jasmine.SpyObj<Store>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;
  });

  // Functional guards use inject(), so they must run inside an injection context.
  const run = () => TestBed.runInInjectionContext(() => userManagementGuard());

  (['SUPER_ADMIN', 'ORG_ADMIN', 'Admin'] as const).forEach((role) => {
    it(`allows access for ${role}`, (done) => {
      store.select.and.returnValue(of(role));

      run().subscribe((result) => {
        expect(result).toBeTrue();
        expect(router.navigate).not.toHaveBeenCalled();
        done();
      });
    });
  });

  (['PropertyManager', 'Accountant', '', 'Unknown'] as const).forEach((role) => {
    it(`denies "${role}" and redirects to /unauthorized`, (done) => {
      store.select.and.returnValue(of(role));

      run().subscribe((result) => {
        expect(result).toBeFalse();
        expect(router.navigate).toHaveBeenCalledWith(['/401']);
        done();
      });
    });
  });
});
