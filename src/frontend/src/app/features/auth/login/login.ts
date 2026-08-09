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
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { setRememberMe } from '../../../shared/utils/auth-storage.utils';
import { notifyError } from '../../../shared/utils/notify.utils';

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
  // Defaults checked: previously every login persisted via localStorage
  // regardless of this checkbox (it was never wired up), so defaulting to
  // checked keeps that behavior for anyone who doesn't touch it.
  rememberMe = true;

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

    // Listen for errors
    this.authService.error$
      .pipe(
        takeUntilDestroyed(),
        filter((error) => !!error)
      )
      .subscribe((error: unknown) => {
        console.error('Login error:', safeErrorMessage(error));
        const errorMessage =
          (error as { message?: string })?.message ||
          'Login failed. Please check your credentials.';
        notifyError(this.snackBar, errorMessage);
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

      // Set before dispatching: the login effect writes tokens synchronously
      // once the HTTP response arrives, so the flag must already be in
      // place by then to land in the right storage.
      setRememberMe(this.rememberMe);

      // Dispatch login action through AuthService (which uses NgRx)
      this.authService.login(loginRequest);
    }
  }
}
