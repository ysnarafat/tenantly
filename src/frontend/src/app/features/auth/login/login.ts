import { Component, inject, OnInit, signal } from '@angular/core';

import {
  ReactiveFormsModule,
  FormsModule,
  FormBuilder,
  FormGroup,
  Validators,
} from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { filter } from 'rxjs';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService, LoginRequest } from '../../../core/services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    FormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    TranslateModule,
  ],
  templateUrl: './login.html',
  styleUrls: ['./login.scss'],
})
export class Login implements OnInit {
  private fb = inject(FormBuilder);
  private authService = inject(AuthService);
  private snackBar = inject(MatSnackBar);

  loginForm: FormGroup;
  loading = signal(false);
  error = signal<unknown>(null);
  showPassword = false;
  rememberMe = false;

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

    // Listen for authentication success — navigation is handled by auth effects
    this.authService.isAuthenticated$
      .pipe(
        takeUntilDestroyed(),
        filter((isAuth) => isAuth)
      )
      .subscribe(() => {
        this.snackBar.open('Login successful!', 'Close', { duration: 3000 });
      });

    // Listen for errors
    this.authService.error$
      .pipe(
        takeUntilDestroyed(),
        filter((error) => !!error)
      )
      .subscribe((error: unknown) => {
        console.error('Login error:', error);
        const errorMessage =
          (error as { message?: string })?.message || 'Login failed. Please check your credentials.';
        this.snackBar.open(errorMessage, 'Close', { duration: 5000 });
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
