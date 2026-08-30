import { Component, inject, signal } from '@angular/core';
import {
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  ValidationErrors,
  ValidatorFn,
  Validators,
} from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

export interface ResetPasswordDialogData {
  username: string;
}

// Mirrors the backend's ValidatePassword rules (user_service.go) closely
// enough to catch obvious misses before a round-trip — the backend stays
// the source of truth, so a mismatch here just means one extra toast, not
// a security gap.
const strongPassword: ValidatorFn = (control): ValidationErrors | null => {
  const value: string = control.value || '';
  const hasUpper = /[A-Z]/.test(value);
  const hasLower = /[a-z]/.test(value);
  const hasDigit = /[0-9]/.test(value);
  const hasSpecial = /[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]/.test(value);
  return hasUpper && hasLower && hasDigit && hasSpecial ? null : { weak: true };
};

const passwordsMatch: ValidatorFn = (group): ValidationErrors | null => {
  const newPassword = group.get('newPassword')?.value;
  const confirmPassword = group.get('confirmPassword')?.value;
  return newPassword === confirmPassword ? null : { mismatch: true };
};

@Component({
  selector: 'app-reset-password-dialog',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
  ],
  templateUrl: './reset-password-dialog.html',
  styleUrl: './reset-password-dialog.scss',
})
export class ResetPasswordDialogComponent {
  private dialogRef = inject(MatDialogRef<ResetPasswordDialogComponent>);
  private fb = inject(FormBuilder);
  data = inject<ResetPasswordDialogData>(MAT_DIALOG_DATA);

  showPassword = signal(false);

  form: FormGroup = this.fb.group(
    {
      newPassword: ['', [Validators.required, Validators.minLength(8), strongPassword]],
      confirmPassword: ['', Validators.required],
    },
    { validators: passwordsMatch }
  );

  onConfirm(): void {
    if (this.form.valid) {
      this.dialogRef.close(this.form.value.newPassword);
    }
  }

  onCancel(): void {
    this.dialogRef.close(null);
  }
}
