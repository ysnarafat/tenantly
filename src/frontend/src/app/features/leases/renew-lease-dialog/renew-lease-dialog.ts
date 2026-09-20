import { Component, inject, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import {
  AbstractControl,
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';
import {
  LeaseService,
  LeaseWithDetails,
  RenewLeaseRequest,
} from '../../../core/services/lease.service';

function atLeastOnePositive(group: AbstractControl): ValidationErrors | null {
  const years = Number(group.get('renew_years')?.value) || 0;
  const months = Number(group.get('renew_months')?.value) || 0;
  return years + months > 0 ? null : { extensionZero: true };
}

// Renewing starts a brand new lease term for the same unit/tenant — the
// current lease is closed out (see LeaseService.renewLease) rather than
// mutated, so its rent/duration/dates remain an accurate historical record.
// This dialog is deliberately pre-populated with the current lease's own
// values (matching the edit dialog's feel) so renewing on unchanged terms is
// a single click, while still letting the landlord adjust anything before
// confirming — a rent increase, an early/late start, a different duration.
@Component({
  selector: 'app-renew-lease-dialog',
  standalone: true,
  imports: [
    CommonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIconModule,
    MatDatepickerModule,
    MatProgressSpinnerModule,
    MatSlideToggleModule,
    ReactiveFormsModule,
    TranslateModule,
  ],
  templateUrl: './renew-lease-dialog.html',
  styleUrl: './renew-lease-dialog.scss',
})
export class RenewLeaseDialog {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<RenewLeaseDialog>);
  private leaseService = inject(LeaseService);
  private snackBar = inject(MatSnackBar);

  lease: LeaseWithDetails;
  submitLoading = false;
  renewForm: FormGroup;

  constructor(@Inject(MAT_DIALOG_DATA) data: { lease: LeaseWithDetails }) {
    this.lease = data.lease;

    const currentYears = Math.floor(this.lease.duration_months / 12);
    const currentMonths = this.lease.duration_months % 12;

    this.renewForm = this.fb.group({
      lease_type: [this.lease.lease_type, Validators.required],
      // Back-to-back continuation by default (matches the backend's own
      // default when start_date is omitted) — editable for an early or late
      // renewal.
      start_date: [new Date(this.lease.end_date), Validators.required],
      duration: this.fb.group(
        {
          renew_years: [currentYears, [Validators.min(0), Validators.max(50)]],
          renew_months: [currentMonths, [Validators.min(0), Validators.max(11)]],
        },
        { validators: atLeastOnePositive }
      ),
      monthly_rent: [this.lease.monthly_rent, [Validators.required, Validators.min(1)]],
      security_deposit: [this.lease.security_deposit ?? 0, [Validators.min(0)]],
      carry_forward_charges: [true],
    });
  }

  get durationMonths(): number {
    const years = Number(this.renewForm.get('duration.renew_years')?.value) || 0;
    const months = Number(this.renewForm.get('duration.renew_months')?.value) || 0;
    return years * 12 + months;
  }

  get newEndDate(): Date | null {
    const start = this.renewForm.get('start_date')?.value as Date | null;
    if (!start || this.durationMonths === 0) return null;
    const d = new Date(start);
    d.setMonth(d.getMonth() + this.durationMonths);
    return d;
  }

  get durationInvalid(): boolean {
    const duration = this.renewForm.get('duration');
    return !!duration?.errors?.['extensionZero'] && duration?.touched;
  }

  formatDate(d: Date): string {
    return d.toLocaleDateString('en-BD', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  onSubmit(): void {
    this.renewForm.get('duration')!.markAllAsTouched();
    if (this.renewForm.invalid) {
      this.renewForm.markAllAsTouched();
      return;
    }

    this.submitLoading = true;
    const v = this.renewForm.value;
    const req: RenewLeaseRequest = {
      start_date: this.formatDateForApi(v.start_date),
      duration_months: this.durationMonths,
      monthly_rent: v.monthly_rent,
      security_deposit: v.security_deposit,
      lease_type: v.lease_type,
      carry_forward_charges: v.carry_forward_charges,
    };

    this.leaseService.renewLease(this.lease.id, req).subscribe({
      next: () => {
        notifySuccess(this.snackBar, 'Lease renewed');
        this.dialogRef.close(true);
        this.submitLoading = false;
      },
      error: (error) => {
        notifyError(this.snackBar, error.error?.error || error.message || 'Failed to renew lease');
        this.submitLoading = false;
      },
    });
  }

  onCancel(): void {
    this.dialogRef.close(false);
  }

  private formatDateForApi(date: Date): string {
    return date.toISOString().split('T')[0];
  }
}
