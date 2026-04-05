import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { of, throwError } from 'rxjs';
import { orgAdminGuard } from './org-admin.guard';

describe('orgAdminGuard', () => {
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

  it('should allow access for SUPER_ADMIN user', (done) => {
    store.select.and.returnValues(of(true), of(false)); // isSuperAdmin=true, isOrgAdmin=false

    const guard = orgAdminGuard();
    guard.subscribe((result) => {
      expect(result).toBe(true);
      expect(router.navigate).not.toHaveBeenCalled();
      done();
    });
  });

  it('should allow access for ORG_ADMIN user', (done) => {
    store.select.and.returnValues(of(false), of(true)); // isSuperAdmin=false, isOrgAdmin=true

    const guard = orgAdminGuard();
    guard.subscribe((result) => {
      expect(result).toBe(true);
      expect(router.navigate).not.toHaveBeenCalled();
      done();
    });
  });

  it('should allow access for both SUPER_ADMIN and ORG_ADMIN', (done) => {
    store.select.and.returnValues(of(true), of(true)); // both true

    const guard = orgAdminGuard();
    guard.subscribe((result) => {
      expect(result).toBe(true);
      expect(router.navigate).not.toHaveBeenCalled();
      done();
    });
  });

  it('should deny access and navigate for non-admin user', (done) => {
    store.select.and.returnValues(of(false), of(false)); // both false

    const guard = orgAdminGuard();
    guard.subscribe((result) => {
      expect(result).toBe(false);
      expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
      done();
    });
  });

  it('should deny access and navigate for Admin user', (done) => {
    store.select.and.returnValues(of(false), of(false));

    const guard = orgAdminGuard();
    guard.subscribe((result) => {
      expect(result).toBe(false);
      expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
      done();
    });
  });

  it('should select both isSuperAdmin and isOrgAdmin from store', (done) => {
    store.select.and.returnValues(of(true), of(false));

    const guard = orgAdminGuard();
    guard.subscribe(() => {
      expect(store.select).toHaveBeenCalledTimes(2);
      done();
    });
  });

  it('should use take(1) operator to complete immediately', (done) => {
    store.select.and.returnValues(of(true), of(false));

    const guard = orgAdminGuard();
    let completionCalled = false;
    guard.subscribe({
      complete: () => {
        completionCalled = true;
        expect(completionCalled).toBe(true);
        done();
      },
    });
  });

  it('should navigate to dashboard when access denied', (done) => {
    store.select.and.returnValues(of(false), of(false));

    const guard = orgAdminGuard();
    guard.subscribe(() => {
      setTimeout(() => {
        expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
        done();
      }, 0);
    });
  });

  it('should handle store selection errors gracefully', (done) => {
    store.select.and.returnValue(throwError(() => new Error('Store error')));

    const guard = orgAdminGuard();
    guard.subscribe(
      () => fail('should have errored'),
      (error) => {
        expect(error).toBeTruthy();
        done();
      }
    );
  });

  it('should use combineLatest to wait for both selectors', (done) => {
    store.select.and.returnValues(of(true), of(true));

    const guard = orgAdminGuard();
    guard.subscribe((result) => {
      expect(result).toBe(true);
      done();
    });
  });

  describe('Access control matrix', () => {
    const testCases = [
      {
        isSuperAdmin: true,
        isOrgAdmin: true,
        shouldAllow: true,
        description: 'Both SUPER_ADMIN and ORG_ADMIN',
      },
      { isSuperAdmin: true, isOrgAdmin: false, shouldAllow: true, description: 'Only SUPER_ADMIN' },
      { isSuperAdmin: false, isOrgAdmin: true, shouldAllow: true, description: 'Only ORG_ADMIN' },
      {
        isSuperAdmin: false,
        isOrgAdmin: false,
        shouldAllow: false,
        description: 'Neither SUPER_ADMIN nor ORG_ADMIN',
      },
    ];

    testCases.forEach(({ isSuperAdmin, isOrgAdmin, shouldAllow, description }) => {
      it(`should ${shouldAllow ? 'allow' : 'deny'} access for: ${description}`, (done) => {
        store.select.and.returnValues(of(isSuperAdmin), of(isOrgAdmin));

        const guard = orgAdminGuard();
        guard.subscribe((result) => {
          expect(result).toBe(shouldAllow);

          if (!shouldAllow) {
            expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
          } else {
            expect(router.navigate).not.toHaveBeenCalled();
          }

          done();
        });
      });
    });
  });

  it('should return observable that completes', (done) => {
    store.select.and.returnValues(of(true), of(false));
    let completed = false;

    const guard = orgAdminGuard();
    guard.subscribe({
      complete: () => {
        completed = true;
        expect(completed).toBe(true);
        done();
      },
    });
  });
});
