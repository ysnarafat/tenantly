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
import {
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  Validators,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';
import { TranslateModule } from '@ngx-translate/core';
import {
  LeaseService,
  LeaseWithDetails,
  UpdateLeaseRequest,
} from '../../../core/services/lease.service';

function atLeastOnePositive(group: AbstractControl): ValidationErrors | null {
  const years = Number(group.get('extend_years')?.value) || 0;
  const months = Number(group.get('extend_months')?.value) || 0;
  return years + months > 0 ? null : { extensionZero: true };
}

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
      extend_lease: [false],
      extension: this.fb.group(
        {
          extend_years: [0, [Validators.min(0), Validators.max(50)]],
          extend_months: [0, [Validators.min(0), Validators.max(11)]],
        },
        { validators: atLeastOnePositive }
      ),
    });

    // Enable/disable extension sub-group based on toggle
    this.editForm.get('extend_lease')!.valueChanges.subscribe((on: boolean) => {
      const ext = this.editForm.get('extension')!;
      if (on) {
        ext.enable();
      } else {
        ext.disable();
      }
    });

    // Start with extension disabled
    this.editForm.get('extension')!.disable();
  }

  get isExtending(): boolean {
    return !!this.editForm.get('extend_lease')?.value;
  }

  get newEndDate(): Date | null {
    if (!this.isExtending) return null;
    const years = Number(this.editForm.get('extension.extend_years')?.value) || 0;
    const months = Number(this.editForm.get('extension.extend_months')?.value) || 0;
    if (years + months === 0) return null;
    const d = new Date(this.lease.end_date);
    d.setFullYear(d.getFullYear() + years);
    d.setMonth(d.getMonth() + months);
    return d;
  }

  get newDurationMonths(): number {
    const years = Number(this.editForm.get('extension.extend_years')?.value) || 0;
    const months = Number(this.editForm.get('extension.extend_months')?.value) || 0;
    return this.lease.duration_months + years * 12 + months;
  }

  get extensionInvalid(): boolean {
    const ext = this.editForm.get('extension');
    return !!ext?.enabled && !!ext?.errors?.['extensionZero'] && ext?.touched;
  }

  formatDate(d: Date): string {
    return d.toLocaleDateString('en-BD', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  onSubmit() {
    if (this.isExtending) {
      this.editForm.get('extension')!.markAllAsTouched();
    }
    if (this.editForm.invalid) {
      this.editForm.markAllAsTouched();
      return;
    }
    if (this.isExtending && this.editForm.get('extension')?.errors?.['extensionZero']) {
      return;
    }

    this.submitLoading = true;
    const v = this.editForm.value;

    const payload: UpdateLeaseRequest = {
      lease_type: v.lease_type,
      monthly_rent: v.monthly_rent,
      security_deposit: v.security_deposit,
      active: v.active,
    };

    if (this.isExtending) {
      payload.duration_months = this.newDurationMonths;
      // Re-activate expired leases when their end date is being pushed forward
      payload.active = true;
    }

    this.leaseService.updateLease(this.lease.id, payload).subscribe({
      next: () => {
        const msg = this.isExtending ? 'Lease extended and updated' : 'Lease updated successfully';
        notifySuccess(this.snackBar, msg);
        this.dialogRef.close(true);
        this.submitLoading = false;
      },
      error: (error) => {
        notifyError(this.snackBar, error.error?.message || 'Failed to update lease');
        this.submitLoading = false;
      },
    });
  }

  onCancel() {
    this.dialogRef.close(false);
  }
}
