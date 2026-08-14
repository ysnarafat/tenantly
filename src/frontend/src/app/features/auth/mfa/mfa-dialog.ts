import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { TranslateModule } from '@ngx-translate/core';
import { MfaService } from '../../../core/services/mfa.service';
import { safeErrorMessage } from '../../../shared/utils/error.utils';

export interface MfaDialogData {
  enrolled: boolean;
}

/**
 * Step-up MFA dialog. When the user is not yet enrolled it first fetches a TOTP
 * secret to display for manual entry into an authenticator app, then verifies a
 * code. When already enrolled it simply prompts for the current code. Closes
 * with the step-up token on success, or undefined on cancel.
 */
@Component({
  selector: 'app-mfa-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    TranslateModule,
  ],
  templateUrl: './mfa-dialog.html',
  styleUrls: ['./mfa-dialog.scss'],
})
export class MfaDialogComponent implements OnInit {
  private dialogRef = inject(MatDialogRef<MfaDialogComponent>);
  private mfaService = inject(MfaService);
  data = inject<MfaDialogData>(MAT_DIALOG_DATA);

  code = '';
  loading = signal(false);
  error = signal('');
  secret = signal('');
  otpauthUri = signal('');

  get needsEnrollment(): boolean {
    return !this.data.enrolled;
  }

  ngOnInit(): void {
    if (this.needsEnrollment) {
      this.loading.set(true);
      this.mfaService.enroll().subscribe({
        next: (res) => {
          this.secret.set(res.secret);
          this.otpauthUri.set(res.otpauth_uri);
          this.loading.set(false);
        },
        error: (err) => {
          this.error.set(safeErrorMessage(err) || 'Failed to start MFA setup');
          this.loading.set(false);
        },
      });
    }
  }

  submit(): void {
    const code = this.code.trim();
    if (code.length < 6) {
      this.error.set('Enter the 6-digit code');
      return;
    }
    this.loading.set(true);
    this.error.set('');
    this.mfaService.verify(code).subscribe({
      next: (res) => {
        this.loading.set(false);
        this.dialogRef.close(res.step_up_token);
      },
      error: (err) => {
        this.loading.set(false);
        this.error.set(safeErrorMessage(err) || 'Invalid code');
      },
    });
  }

  cancel(): void {
    this.dialogRef.close();
  }
}
