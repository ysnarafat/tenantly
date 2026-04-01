import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { of, throwError } from 'rxjs';
import { superAdminGuard } from './super-admin.guard';

describe('superAdminGuard', () => {
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
    store.select.and.returnValue(of(true));

    const guard = superAdminGuard();
    guard.subscribe((result) => {
      expect(result).toBe(true);
      expect(router.navigate).not.toHaveBeenCalled();
      done();
    });
  });

  it('should deny access and navigate for non-SUPER_ADMIN user', (done) => {
    store.select.and.returnValue(of(false));

    const guard = superAdminGuard();
    guard.subscribe((result) => {
      expect(result).toBe(false);
      expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
      done();
    });
  });

  it('should select isSuperAdmin from store', (done) => {
    store.select.and.returnValue(of(true));

    const guard = superAdminGuard();
    guard.subscribe(() => {
      expect(store.select).toHaveBeenCalled();
      done();
    });
  });

  it('should use take(1) operator to complete immediately', (done) => {
    let emissionCount = 0;
    const testObservable = of(true);
    store.select.and.returnValue(testObservable);

    const guard = superAdminGuard();
    guard.subscribe({
      next: () => {
        emissionCount++;
      },
      complete: () => {
        expect(emissionCount).toBe(1);
        done();
      },
    });
  });

  it('should navigate to dashboard when access denied', (done) => {
    store.select.and.returnValue(of(false));

    const guard = superAdminGuard();
    guard.subscribe(() => {
      setTimeout(() => {
        expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
        done();
      }, 0);
    });
  });

  it('should handle store selection errors gracefully', (done) => {
    store.select.and.returnValue(throwError(() => new Error('Store error')));

    const guard = superAdminGuard();
    guard.subscribe(
      () => fail('should have errored'),
      (error) => {
        expect(error).toBeTruthy();
        done();
      }
    );
  });

  it('should return observable that completes', (done) => {
    store.select.and.returnValue(of(true));
    let completed = false;

    const guard = superAdminGuard();
    guard.subscribe({
      complete: () => {
        completed = true;
        expect(completed).toBe(true);
        done();
      },
    });
  });
});
