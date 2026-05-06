import { Component, inject, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { LeaseService, LeaseWithDetails } from '../../../core/services/lease.service';

@Component({
  selector: 'app-edit-lease-dialog',
  standalone: true,
  imports: [
    CommonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatSlideToggleModule,
    ReactiveFormsModule,
    TranslateModule,
  ],
  templateUrl: './edit-lease-dialog.html',
  styleUrls: ['./edit-lease-dialog.scss'],
})
export class EditLeaseDialog {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<EditLeaseDialog>);
  private leaseService = inject(LeaseService);
  private snackBar = inject(MatSnackBar);

  lease: LeaseWithDetails;
  submitLoading = false;
  editForm: FormGroup;

  constructor(@Inject(MAT_DIALOG_DATA) data: { lease: LeaseWithDetails }) {
    this.lease = data.lease;
    this.editForm = this.fb.group({
      lease_type: [this.lease.lease_type, Validators.required],
      monthly_rent: [this.lease.monthly_rent, [Validators.required, Validators.min(1)]],
      security_deposit: [this.lease.security_deposit ?? 0, [Validators.min(0)]],
      active: [this.lease.active],
    });
  }

  onSubmit() {
    if (this.editForm.invalid) {
      this.editForm.markAllAsTouched();
      return;
    }

    this.submitLoading = true;
    const v = this.editForm.value;

    this.leaseService
      .updateLease(this.lease.id, {
        lease_type: v.lease_type,
        monthly_rent: v.monthly_rent,
        security_deposit: v.security_deposit,
        active: v.active,
      })
      .subscribe({
        next: () => {
          this.snackBar.open('Lease updated successfully', 'Close', { duration: 3000 });
          this.dialogRef.close(true);
          this.submitLoading = false;
        },
        error: (error) => {
          this.snackBar.open(error.error?.message || 'Failed to update lease', 'Close', {
            duration: 5000,
          });
          this.submitLoading = false;
        },
      });
  }

  onCancel() {
    this.dialogRef.close(false);
  }
}
