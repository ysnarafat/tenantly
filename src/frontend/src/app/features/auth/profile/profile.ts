import { Component, OnInit, inject, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Actions, ofType } from '@ngrx/effects';
import { take } from 'rxjs/operators';
import { AuthService, User, ChangePasswordRequest } from '../../../core/services/auth.service';
import * as AuthActions from '../../../store/auth/auth.actions';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
  ],
  templateUrl: './profile.html',
  styleUrls: ['./profile.scss'],
})
export class Profile implements OnInit {
  private fb = inject(FormBuilder);
  private authService = inject(AuthService);
  private snackBar = inject(MatSnackBar);
  private actions$ = inject(Actions);
  private destroyRef = inject(DestroyRef);

  user: User | null = null;
  changePasswordForm: FormGroup;
  showChangePassword = false;
  isChangingPassword = false;

  constructor() {
    this.changePasswordForm = this.fb.group(
      {
        currentPassword: ['', Validators.required],
        newPassword: ['', [Validators.required, Validators.minLength(8)]],
        confirmPassword: ['', Validators.required],
      },
      { validators: this.passwordMatchValidator }
    );
  }

  ngOnInit() {
    this.user = this.authService.getUser();
  }

  passwordMatchValidator(form: FormGroup) {
    const newPassword = form.get('newPassword');
    const confirmPassword = form.get('confirmPassword');

    if (newPassword && confirmPassword && newPassword.value !== confirmPassword.value) {
      confirmPassword.setErrors({ passwordMismatch: true });
      return { passwordMismatch: true };
    }

    return null;
  }

  toggleChangePassword() {
    this.showChangePassword = !this.showChangePassword;
    if (!this.showChangePassword) {
      this.changePasswordForm.reset();
    }
  }

  onChangePassword() {
    if (this.changePasswordForm.valid) {
      this.isChangingPassword = true;

      const request: ChangePasswordRequest = {
        current_password: this.changePasswordForm.value.currentPassword,
        new_password: this.changePasswordForm.value.newPassword,
      };

      this.authService.changePassword(request);

      this.actions$
        .pipe(
          ofType(AuthActions.changePasswordSuccess, AuthActions.changePasswordFailure),
          take(1),
          takeUntilDestroyed(this.destroyRef)
        )
        .subscribe((action) => {
          this.isChangingPassword = false;
          if (action.type === AuthActions.changePasswordSuccess.type) {
            notifySuccess(this.snackBar, 'Password changed successfully!');
            this.changePasswordForm.reset();
            this.showChangePassword = false;
          } else {
            const failure = action as ReturnType<typeof AuthActions.changePasswordFailure>;
            const msg =
              (failure.error as { message?: string })?.message ||
              'Failed to change password. Please try again.';
            notifyError(this.snackBar, msg);
          }
        });
    }
  }

  getRoleColor(role: string): string {
    switch (role) {
      case 'Admin':
        return 'primary';
      case 'PropertyManager':
        return 'accent';
      case 'Accountant':
        return 'warn';
      default:
        return 'basic';
    }
  }

  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-BD', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }
}
