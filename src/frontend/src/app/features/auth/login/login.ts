import { Component, inject, OnInit, signal } from '@angular/core';

import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { filter } from 'rxjs';
import { AuthService, LoginRequest } from '../../../core/services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './login.html',
  styleUrls: ['./login.scss'],
})
export class Login implements OnInit {
  private fb = inject(FormBuilder);
  private authService = inject(AuthService);
  private router = inject(Router);
  private snackBar = inject(MatSnackBar);

  loginForm: FormGroup;
  loading = signal(false);
  error = signal<any>(null);

  constructor() {
    this.loginForm = this.fb.group({
      username: ['', Validators.required],
      password: ['', Validators.required],
    });

    // Subscribe to auth state and update signals
    this.authService.loading$
      .pipe(takeUntilDestroyed())
      .subscribe((loading) => this.loading.set(loading));

    this.authService.error$.pipe(takeUntilDestroyed()).subscribe((error) => this.error.set(error));

    // Listen for authentication success
    this.authService.isAuthenticated$
      .pipe(
        takeUntilDestroyed(),
        filter((isAuth) => isAuth)
      )
      .subscribe(() => {
        this.snackBar.open('Login successful!', 'Close', { duration: 3000 });
        this.router.navigate(['/dashboard']);
      });

    // Listen for errors
    this.authService.error$
      .pipe(
        takeUntilDestroyed(),
        filter((error) => !!error)
      )
      .subscribe((error: any) => {
        console.error('Login error:', error);
        this.snackBar.open(
          error.error?.error || 'Login failed. Please check your credentials.',
          'Close',
          { duration: 5000 }
        );
      });
  }

  ngOnInit() {
    // Clear any previous errors when component initializes
    this.authService.clearError();
  }

  onSubmit() {
    if (this.loginForm.valid) {
      const loginRequest: LoginRequest = {
        username: this.loginForm.value.username,
        password: this.loginForm.value.password,
      };

      // Dispatch login action through AuthService (which uses NgRx)
      this.authService.login(loginRequest);
    }
  }
}
