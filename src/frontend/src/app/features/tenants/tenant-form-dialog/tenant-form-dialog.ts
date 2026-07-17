import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { TranslateModule } from '@ngx-translate/core';
import { Tenant, TenantType } from '../../../core/models/tenant.model';
import { PermissionService } from '../../../core/services/permission.service';
import { maskNid, maskPhone } from '../../../shared/utils/pii-mask.utils';

export interface TenantFormDialogData {
  tenant?: Tenant;
  mode: 'create' | 'edit' | 'view';
}

@Component({
  selector: 'app-tenant-form-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatSelectModule,
    MatIconModule,
    MatTooltipModule,
    TranslateModule,
  ],
  templateUrl: './tenant-form-dialog.html',
  styleUrls: ['./tenant-form-dialog.scss'],
})
export class TenantFormDialogComponent implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<TenantFormDialogComponent>);
  private permissionService = inject(PermissionService);
  public data = inject<TenantFormDialogData>(MAT_DIALOG_DATA);

  tenantForm!: FormGroup;
  tenantTypes: TenantType[] = ['Individual', 'Business'];

  nidRevealed = false;
  phoneRevealed = false;
  private _realNid = '';
  private _realPhone = '';

  get hasNid(): boolean {
    return !!this._realNid;
  }
  get hasPhone(): boolean {
    return !!this._realPhone;
  }

  get canRevealPii(): boolean {
    return (
      this.permissionService.isSuperAdmin() ||
      this.permissionService.isOrgAdmin() ||
      this.permissionService.isAdmin() ||
      this.permissionService.isPropertyManager()
    );
  }

  toggleNidReveal(): void {
    this.nidRevealed = !this.nidRevealed;
    if (this.isViewMode) {
      this.tenantForm
        .get('nid_number')
        ?.setValue(this.nidRevealed ? this._realNid : maskNid(this._realNid), { emitEvent: false });
    }
  }

  togglePhoneReveal(): void {
    this.phoneRevealed = !this.phoneRevealed;
    if (this.isViewMode) {
      this.tenantForm
        .get('phone_number')
        ?.setValue(this.phoneRevealed ? this._realPhone : maskPhone(this._realPhone), {
          emitEvent: false,
        });
    }
  }

  ngOnInit() {
    this.initializeForm();
  }

  private initializeForm() {
    const tenant = this.data.tenant;

    this._realNid = tenant?.nid_number || '';
    this._realPhone = tenant?.phone_number || '';

    this.tenantForm = this.fb.group({
      name: [
        tenant?.name || '',
        this.isViewMode ? [] : [Validators.required, Validators.maxLength(100)],
      ],
      tenant_type: [
        tenant?.tenant_type || 'Individual',
        this.isViewMode ? [] : [Validators.required],
      ],
      nid_number: [
        this.isViewMode ? maskNid(this._realNid) : this._realNid,
        this.isViewMode ? [] : [Validators.required, Validators.maxLength(20)],
      ],
      phone_number: [
        this.isViewMode ? maskPhone(this._realPhone) : this._realPhone,
        this.isViewMode ? [] : [Validators.required, Validators.maxLength(20)],
      ],
      email: [tenant?.email || '', this.isViewMode ? [] : [Validators.email]],
      address: [tenant?.address || ''],
    });

    if (this.isViewMode) {
      this.tenantForm.disable();
    }
  }

  onSubmit() {
    if (this.tenantForm.valid) {
      this.dialogRef.close(this.tenantForm.value);
    } else {
      Object.keys(this.tenantForm.controls).forEach((key) => {
        this.tenantForm.get(key)?.markAsTouched();
      });
    }
  }

  onCancel() {
    this.dialogRef.close();
  }

  getErrorMessage(fieldName: string): string {
    const control = this.tenantForm.get(fieldName);
    if (!control || !control.errors || !control.touched) {
      return '';
    }

    if (control.errors['required']) {
      return `${this.getFieldLabel(fieldName)} is required`;
    }
    if (control.errors['email']) {
      return 'Invalid email address';
    }
    if (control.errors['maxlength']) {
      return `Maximum length is ${control.errors['maxlength'].requiredLength}`;
    }
    return 'Invalid value';
  }

  private getFieldLabel(fieldName: string): string {
    type Labels = Record<string, string>;
    const labels: Labels = {
      name: 'Full Name',
      tenant_type: 'Tenant Type',
      email: 'Email',
      phone_number: 'Phone Number',
      nid_number: 'NID/Registration No.',
      address: 'Address',
    };
    return labels[fieldName] || fieldName;
  }

  get isEditMode(): boolean {
    return this.data.mode === 'edit';
  }

  get isViewMode(): boolean {
    return this.data.mode === 'view';
  }

  get dialogTitle(): string {
    if (this.isViewMode) {
      return 'Tenant Details';
    }
    return this.isEditMode ? 'Edit Tenant' : 'Add New Tenant';
  }
}
