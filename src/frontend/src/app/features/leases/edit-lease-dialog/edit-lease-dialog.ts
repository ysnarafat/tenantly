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
import {
  FormBuilder,
  FormControl,
  FormGroup,
  ReactiveFormsModule,
  Validators,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { TranslateModule } from '@ngx-translate/core';
import {
  LeaseService,
  LeaseWithDetails,
  UpdateLeaseRequest,
  LeaseCustomFields,
} from '../../../core/services/lease.service';
import { LeaseChargesEditor, LeaseChargeDraft } from '../lease-charges-editor/lease-charges-editor';
import { LeaseCustomFieldsEditor } from '../lease-custom-fields-editor/lease-custom-fields-editor';

function endDateAfterStartDate(group: AbstractControl): ValidationErrors | null {
  const start = group.get('start_date')?.value as Date | null;
  const end = group.get('end_date')?.value as Date | null;
  if (!start || !end) return null;
  return end > start ? null : { endBeforeStart: true };
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
    MatDatepickerModule,
    MatProgressSpinnerModule,
    ReactiveFormsModule,
    TranslateModule,
    LeaseChargesEditor,
    LeaseCustomFieldsEditor,
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

  // Ending a tenancy replaces the normal edit view with a focused confirm
  // step, rather than being a silent toggle among the other fields — the
  // move-out date has to be explicit since it's what closes out the lease's
  // historical record (see LeaseService.terminateLease).
  endingTenancy = false;
  moveOutDate = new FormControl<Date>(new Date(), Validators.required);
  minMoveOutDate: Date;

  // The lease passed in from the list view doesn't carry its charges/custom
  // fields (the list endpoint omits them to avoid an N+1 query per row) —
  // fetched separately below once the dialog opens.
  charges: LeaseChargeDraft[] = [];
  chargesLoading = true;
  customFields: LeaseCustomFields = {};

  constructor(@Inject(MAT_DIALOG_DATA) data: { lease: LeaseWithDetails }) {
    this.lease = data.lease;
    this.minMoveOutDate = new Date(this.lease.start_date);
    this.customFields = this.lease.custom_fields ?? {};
    this.loadFullDetails();
    this.editForm = this.fb.group(
      {
        lease_type: [this.lease.lease_type, Validators.required],
        start_date: [new Date(this.lease.start_date), Validators.required],
        end_date: [new Date(this.lease.end_date), Validators.required],
        monthly_rent: [this.lease.monthly_rent, [Validators.required, Validators.min(1)]],
        security_deposit: [this.lease.security_deposit ?? 0, [Validators.min(0)]],
      },
      { validators: endDateAfterStartDate }
    );
  }

  private loadFullDetails(): void {
    this.leaseService.getLeaseById(this.lease.id).subscribe({
      next: (details) => {
        this.charges = details.charges ?? [];
        this.customFields = details.custom_fields ?? {};
        this.chargesLoading = false;
      },
      error: (error) => {
        console.error('Error loading lease charges:', safeErrorMessage(error));
        this.chargesLoading = false;
      },
    });
  }

  formatDate(d: Date): string {
    return d.toLocaleDateString('en-BD', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  onSubmit() {
    if (this.editForm.invalid) {
      this.editForm.markAllAsTouched();
      return;
    }

    this.submitLoading = true;
    const v = this.editForm.value;

    const payload: UpdateLeaseRequest = {
      lease_type: v.lease_type,
      start_date: this.formatDateForApi(v.start_date),
      end_date: this.formatDateForApi(v.end_date),
      monthly_rent: v.monthly_rent,
      security_deposit: v.security_deposit,
      custom_fields: this.customFields,
    };

    this.leaseService.updateLease(this.lease.id, payload).subscribe({
      next: () => {
        notifySuccess(this.snackBar, 'Lease updated successfully');
        this.dialogRef.close(true);
        this.submitLoading = false;
      },
      error: (error) => {
        notifyError(this.snackBar, error.error?.error || error.message || 'Failed to update lease');
        this.submitLoading = false;
      },
    });
  }

  startEndingTenancy() {
    this.endingTenancy = true;
  }

  cancelEndingTenancy() {
    this.endingTenancy = false;
  }

  confirmEndTenancy() {
    this.moveOutDate.markAsTouched();
    if (this.moveOutDate.invalid || !this.moveOutDate.value) {
      return;
    }

    this.submitLoading = true;
    const terminationDate = this.moveOutDate.value.toISOString().split('T')[0];
    this.leaseService
      .terminateLease(this.lease.id, { termination_date: terminationDate })
      .subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Tenancy ended');
          this.dialogRef.close(true);
          this.submitLoading = false;
        },
        error: (error) => {
          notifyError(
            this.snackBar,
            error.error?.error || error.message || 'Failed to end tenancy'
          );
          this.submitLoading = false;
        },
      });
  }

  onCancel() {
    this.dialogRef.close(false);
  }

  private formatDateForApi(date: Date): string {
    return date.toISOString().split('T')[0];
  }
}
