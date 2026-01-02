import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatSnackBar } from '@angular/material/snack-bar';
import { AuthService, User, ChangePasswordRequest } from '../../../core/services/auth.service';

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

      this.authService.changePassword(request).subscribe({
        next: () => {
          this.snackBar.open('Password changed successfully!', 'Close', { duration: 3000 });
          this.changePasswordForm.reset();
          this.showChangePassword = false;
          this.isChangingPassword = false;
        },
        error: (error) => {
          console.error('Change password error:', error);
          this.snackBar.open(
            error.error?.error || 'Failed to change password. Please try again.',
            'Close',
            { duration: 5000 }
          );
          this.isChangingPassword = false;
        },
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
