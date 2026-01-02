import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { Store } from '@ngrx/store';
import { of, Subject } from 'rxjs';
import { Login } from './login';
import { AuthService } from '../../../core/services/auth.service';
import { AppState } from '../../../store';

describe('Login Component', () => {
  let component: Login;
  let fixture: ComponentFixture<Login>;
  let authService: jasmine.SpyObj<AuthService>;
  let store: jasmine.SpyObj<Store<AppState>>;
  let router: jasmine.SpyObj<Router>;
  let snackBar: jasmine.SpyObj<MatSnackBar>;

  const mockAuthService = {
    login: jasmine.createSpy('login'),
    clearError: jasmine.createSpy('clearError'),
    loading$: of(false),
    error$: of(null),
    isAuthenticated$: of(false),
  };

  beforeEach(async () => {
    const storeSpy = jasmine.createSpyObj('Store', ['select', 'dispatch']);
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);
    const snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);

    await TestBed.configureTestingModule({
      imports: [Login, ReactiveFormsModule, NoopAnimationsModule],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        { provide: Store, useValue: storeSpy },
        { provide: Router, useValue: routerSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(Login);
    component = fixture.componentInstance;
    authService = TestBed.inject(AuthService) as jasmine.SpyObj<AuthService>;
    store = TestBed.inject(Store) as jasmine.SpyObj<Store<AppState>>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;
    snackBar = TestBed.inject(MatSnackBar) as jasmine.SpyObj<MatSnackBar>;

    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize form with empty values', () => {
    expect(component.loginForm.get('username')?.value).toBe('');
    expect(component.loginForm.get('password')?.value).toBe('');
  });

  it('should clear errors on initialization', () => {
    expect(authService.clearError).toHaveBeenCalled();
  });

  it('should call authService.login when form is valid and submitted', () => {
    component.loginForm.patchValue({
      username: 'testuser',
      password: 'testpass',
    });

    component.onSubmit();

    expect(authService.login).toHaveBeenCalledWith({
      username: 'testuser',
      password: 'testpass',
    });
  });

  it('should not call authService.login when form is invalid', () => {
    component.loginForm.patchValue({
      username: '',
      password: '',
    });

    component.onSubmit();

    expect(authService.login).not.toHaveBeenCalled();
  });

  it('should navigate to dashboard and show success message on successful authentication', () => {
    const isAuthenticatedSubject = new Subject<boolean>();
    authService.isAuthenticated$ = isAuthenticatedSubject.asObservable();

    component.ngOnInit();

    isAuthenticatedSubject.next(true);

    expect(snackBar.open).toHaveBeenCalledWith('Login successful!', 'Close', { duration: 3000 });
    expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
  });

  it('should show error message when authentication fails', () => {
    const errorSubject = new Subject<any>();
    authService.error$ = errorSubject.asObservable();

    component.ngOnInit();

    const error = { error: { error: 'Invalid credentials' } };
    errorSubject.next(error);

    expect(snackBar.open).toHaveBeenCalledWith('Invalid credentials', 'Close', { duration: 5000 });
  });

  it('should show default error message when error has no specific message', () => {
    const errorSubject = new Subject<any>();
    authService.error$ = errorSubject.asObservable();

    component.ngOnInit();

    const error = {};
    errorSubject.next(error);

    expect(snackBar.open).toHaveBeenCalledWith(
      'Login failed. Please check your credentials.',
      'Close',
      { duration: 5000 }
    );
  });

  it('should have loading$ observable from authService', () => {
    expect(component.loading$).toBe(authService.loading$);
  });

  it('should have error$ observable from authService', () => {
    expect(component.error$).toBe(authService.error$);
  });

  it('should unsubscribe on destroy', () => {
    spyOn(component['destroy$'], 'next');
    spyOn(component['destroy$'], 'complete');

    component.ngOnDestroy();

    expect(component['destroy$'].next).toHaveBeenCalled();
    expect(component['destroy$'].complete).toHaveBeenCalled();
  });
});
