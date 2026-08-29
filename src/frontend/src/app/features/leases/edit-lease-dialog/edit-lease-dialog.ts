import { Component, inject, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
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
  RenewLeaseRequest,
  LeaseCustomFields,
} from '../../../core/services/lease.service';
import { LeaseChargesEditor, LeaseChargeDraft } from '../lease-charges-editor/lease-charges-editor';
import { LeaseCustomFieldsEditor } from '../lease-custom-fields-editor/lease-custom-fields-editor';

function atLeastOnePositive(group: AbstractControl): ValidationErrors | null {
  const years = Number(group.get('renew_years')?.value) || 0;
  const months = Number(group.get('renew_months')?.value) || 0;
  return years + months > 0 ? null : { extensionZero: true };
}

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
    MatNativeDateModule,
    MatProgressSpinnerModule,
    MatSlideToggleModule,
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
        // Corrects the CURRENT lease's own term (e.g. a data-entry mistake
        // made at creation) — distinct from "Renew lease" below, which
        // starts a new lease term rather than editing this one's dates.
        start_date: [new Date(this.lease.start_date), Validators.required],
        end_date: [new Date(this.lease.end_date), Validators.required],
        monthly_rent: [this.lease.monthly_rent, [Validators.required, Validators.min(1)]],
        security_deposit: [this.lease.security_deposit ?? 0, [Validators.min(0)]],
        renew_lease: [false],
        renewal: this.fb.group(
          {
            renew_years: [0, [Validators.min(0), Validators.max(50)]],
            renew_months: [0, [Validators.min(0), Validators.max(11)]],
          },
          { validators: atLeastOnePositive }
        ),
      },
      { validators: endDateAfterStartDate }
    );

    // Enable/disable the renewal sub-group based on the toggle
    this.editForm.get('renew_lease')!.valueChanges.subscribe((on: boolean) => {
      const renewal = this.editForm.get('renewal')!;
      if (on) {
        renewal.enable();
      } else {
        renewal.disable();
      }
    });

    // Start with renewal fields disabled
    this.editForm.get('renewal')!.disable();
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

  get isRenewing(): boolean {
    return !!this.editForm.get('renew_lease')?.value;
  }

  // Length of the NEW lease term being created — not cumulative with the
  // current lease's own duration, since renewing creates a separate lease.
  get renewDurationMonths(): number {
    const years = Number(this.editForm.get('renewal.renew_years')?.value) || 0;
    const months = Number(this.editForm.get('renewal.renew_months')?.value) || 0;
    return years * 12 + months;
  }

  get renewedEndDate(): Date | null {
    if (!this.isRenewing || this.renewDurationMonths === 0) return null;
    const d = new Date(this.lease.end_date);
    d.setMonth(d.getMonth() + this.renewDurationMonths);
    return d;
  }

  get renewalInvalid(): boolean {
    const renewal = this.editForm.get('renewal');
    return !!renewal?.enabled && !!renewal?.errors?.['extensionZero'] && renewal?.touched;
  }

  formatDate(d: Date): string {
    return d.toLocaleDateString('en-BD', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  onSubmit() {
    if (this.isRenewing) {
      this.editForm.get('renewal')!.markAllAsTouched();
    }
    if (this.editForm.invalid) {
      this.editForm.markAllAsTouched();
      return;
    }
    if (this.isRenewing && this.editForm.get('renewal')?.errors?.['extensionZero']) {
      return;
    }

    this.submitLoading = true;
    const v = this.editForm.value;

    if (this.isRenewing) {
      const req: RenewLeaseRequest = {
        duration_months: this.renewDurationMonths,
        monthly_rent: v.monthly_rent,
        security_deposit: v.security_deposit,
        lease_type: v.lease_type,
      };
      this.leaseService.renewLease(this.lease.id, req).subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Lease renewed');
          this.dialogRef.close(true);
          this.submitLoading = false;
        },
        error: (error) => {
          notifyError(
            this.snackBar,
            error.error?.error || error.message || 'Failed to renew lease'
          );
          this.submitLoading = false;
        },
      });
      return;
    }

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
