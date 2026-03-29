import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { CreateTenantRequest, Tenant, TenantType } from '../../../core/models/tenant.model';

export interface TenantFormDialogData {
  tenant?: Tenant;
  mode: 'create' | 'edit';
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
  ],
  templateUrl: './tenant-form-dialog.html',
  styleUrls: ['./tenant-form-dialog.scss'],
})
export class TenantFormDialogComponent implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<TenantFormDialogComponent>);
  public data = inject<TenantFormDialogData>(MAT_DIALOG_DATA);

  tenantForm!: FormGroup;
  tenantTypes: TenantType[] = ['Individual', 'Business'];

  ngOnInit() {
    this.initializeForm();
  }

  private initializeForm() {
    const tenant = this.data.tenant;

    this.tenantForm = this.fb.group({
      name: [tenant?.name || '', [Validators.required, Validators.maxLength(100)]],
      tenant_type: [tenant?.tenant_type || 'Individual', [Validators.required]],
      nid_number: [tenant?.nid_number || '', [Validators.required, Validators.maxLength(20)]],
      phone_number: [tenant?.phone_number || '', [Validators.required, Validators.maxLength(20)]],
      email: [tenant?.email || '', [Validators.email]],
      address: [tenant?.address || ''],
    });
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
    const labels: { [key: string]: string } = {
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

  get dialogTitle(): string {
    return this.isEditMode ? 'Edit Tenant' : 'Add New Tenant';
  }
}
