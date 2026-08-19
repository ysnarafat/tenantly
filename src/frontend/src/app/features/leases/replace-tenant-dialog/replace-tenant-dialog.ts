import { Component, Inject, OnInit, inject } from '@angular/core';
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
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { LeaseService, LeaseWithDetails } from '../../../core/services/lease.service';
import { TenantService } from '../../../core/services/tenant.service';
import { Tenant } from '../../../core/models/tenant.model';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

export interface ReplaceTenantDialogData {
  lease: LeaseWithDetails;
}

/**
 * Hands a unit over from its current tenant to a new one in a single step.
 *
 * The outgoing lease is closed at the handover date and a successor lease opens
 * the same day — the API does both in one transaction. Terms default to the
 * outgoing lease's, so the usual turnover needs only a tenant and a date.
 */
@Component({
  selector: 'app-replace-tenant-dialog',
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
    ReactiveFormsModule,
    TranslateModule,
  ],
  templateUrl: './replace-tenant-dialog.html',
  styleUrls: ['./replace-tenant-dialog.scss'],
})
export class ReplaceTenantDialog implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<ReplaceTenantDialog>);
  private leaseService = inject(LeaseService);
  private tenantService = inject(TenantService);
  private snackBar = inject(MatSnackBar);
  private translate = inject(TranslateService);

  lease: LeaseWithDetails;
  form: FormGroup;

  loadingTenants = false;
  submitLoading = false;

  /** Active tenants in the org, minus the one already on this lease. */
  candidateTenants: Tenant[] = [];

  /** The handover cannot predate the lease it is ending. */
  minHandoverDate: Date;

  constructor(@Inject(MAT_DIALOG_DATA) data: ReplaceTenantDialogData) {
    this.lease = data.lease;
    this.minHandoverDate = new Date(this.lease.start_date);

    this.form = this.fb.group({
      new_tenant_id: [null, Validators.required],
      handover_date: [new Date(), Validators.required],
      // Seeded from the outgoing lease: turnover usually keeps the same terms,
      // and anything the user does not touch is sent as-is.
      lease_type: [this.lease.lease_type, Validators.required],
      duration_months: [
        this.lease.duration_months,
        [Validators.required, Validators.min(1), Validators.max(60)],
      ],
      monthly_rent: [this.lease.monthly_rent, [Validators.required, Validators.min(1)]],
      security_deposit: [this.lease.security_deposit ?? 0, [Validators.min(0)]],
    });
  }

  ngOnInit(): void {
    this.loadTenants();
  }

  private loadTenants(): void {
    this.loadingTenants = true;
    this.tenantService.getAllTenants(1, 100).subscribe({
      next: (response) => {
        // The outgoing tenant is excluded: the API rejects replacing a tenant
        // with itself, so offering it would only produce an error.
        this.candidateTenants = response.tenants.filter(
          (tenant) => tenant.active && tenant.id !== this.lease.tenant_id
        );
        this.loadingTenants = false;
      },
      error: (error) => {
        console.error('Error loading tenants:', safeErrorMessage(error));
        notifyError(
          this.snackBar,
          this.translate.instant('REPLACE_TENANT_DIALOG.ERRORS.LOAD_TENANTS')
        );
        this.loadingTenants = false;
      },
    });
  }

  /** End date the successor lease will get, shown so the dates are not a surprise. */
  get successorEndDate(): Date | null {
    const handover = this.form.get('handover_date')?.value as Date | null;
    const months = this.form.get('duration_months')?.value as number | null;
    if (!handover || !months) return null;

    const end = new Date(handover);
    end.setMonth(end.getMonth() + months);
    return end;
  }

  onSubmit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.submitLoading = true;
    const value = this.form.value;

    this.leaseService
      .replaceTenant(this.lease.id, {
        new_tenant_id: value.new_tenant_id,
        handover_date: this.formatDateForApi(value.handover_date),
        lease_type: value.lease_type,
        duration_months: value.duration_months,
        monthly_rent: value.monthly_rent,
        security_deposit: value.security_deposit ?? 0,
      })
      .subscribe({
        next: (result) => {
          notifySuccess(
            this.snackBar,
            this.translate.instant('REPLACE_TENANT_DIALOG.SUCCESS', {
              tenant: result.new_lease.tenant_name,
              unit: result.new_lease.unit_number,
            })
          );
          this.dialogRef.close(true);
          this.submitLoading = false;
        },
        error: (error) => {
          console.error('Error replacing tenant:', safeErrorMessage(error));
          notifyError(
            this.snackBar,
            error.error?.message || this.translate.instant('REPLACE_TENANT_DIALOG.ERRORS.FAILED')
          );
          this.submitLoading = false;
        },
      });
  }

  onCancel(): void {
    this.dialogRef.close(false);
  }

  // Send the local calendar date the user picked. toISOString() would shift it
  // by the UTC offset, landing a Dhaka-evening pick on the previous day.
  private formatDateForApi(date: Date): string {
    const year = date.getFullYear();
    const month = `${date.getMonth() + 1}`.padStart(2, '0');
    const day = `${date.getDate()}`.padStart(2, '0');
    return `${year}-${month}-${day}`;
  }
}
