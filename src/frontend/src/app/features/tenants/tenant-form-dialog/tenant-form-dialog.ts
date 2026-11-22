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
import { switchMap } from 'rxjs/operators';
import { Tenant, TenantType } from '../../../core/models/tenant.model';
import { PermissionService } from '../../../core/services/permission.service';
import { TenantService } from '../../../core/services/tenant.service';
import { MfaService } from '../../../core/services/mfa.service';
import { maskFromLastFour, maskPhone } from '../../../shared/utils/pii-mask.utils';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { DocumentAttachmentsComponent } from '../../../shared/components/document-attachments/document-attachments.component';

export interface TenantFormDialogData {
  tenant?: Tenant;
  mode: 'create' | 'edit' | 'view';
  attachments?: {
    id: number;
    file_name: string;
    content_type: string;
    file_size: number;
    uploaded_at: string;
  }[];
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
    DocumentAttachmentsComponent,
  ],
  templateUrl: './tenant-form-dialog.html',
  styleUrls: ['./tenant-form-dialog.scss'],
})
export class TenantFormDialogComponent implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<TenantFormDialogComponent>);
  private permissionService = inject(PermissionService);
  private tenantService = inject(TenantService);
  private mfaService = inject(MfaService);
  public data = inject<TenantFormDialogData>(MAT_DIALOG_DATA);

  tenantForm!: FormGroup;
  attachments = this.data.attachments || [];
  isUploading = false;

  tenantTypes: TenantType[] = ['Individual', 'Business'];

  nidRevealed = false;
  phoneRevealed = false;
  private _realNid = '';
  private _nidLastFour = '';
  private _realPhone = '';

  get hasNid(): boolean {
    return !!this._nidLastFour || !!this._realNid;
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
    const willReveal = !this.nidRevealed;

    // Fetch the full NID on demand the first time it is revealed, gated by an
    // MFA step-up challenge.
    if (willReveal && !this._realNid && this.data.tenant?.id) {
      const id = this.data.tenant.id;
      this.mfaService
        .ensureStepUp()
        .pipe(switchMap((token) => this.tenantService.getTenantNid(id, token)))
        .subscribe({
          next: (res) => {
            this._realNid = res.nid_number;
            this.nidRevealed = true;
            if (this.isViewMode) {
              this.tenantForm.get('nid_number')?.setValue(this._realNid, { emitEvent: false });
            }
          },
          error: (err) => {
            this.mfaService.clearStepUp();
            console.error('Error revealing NID:', safeErrorMessage(err));
          },
        });
      return;
    }

    this.nidRevealed = willReveal;
    if (this.isViewMode) {
      this.tenantForm
        .get('nid_number')
        ?.setValue(this.nidRevealed ? this._realNid : maskFromLastFour(this._nidLastFour), {
          emitEvent: false,
        });
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

    this._nidLastFour = tenant?.nid_last_four || '';
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
        this.isViewMode ? maskFromLastFour(this._nidLastFour) : '',
        // NID is required only when creating. In edit mode it is left blank and
        // an empty value is dropped on submit so the existing NID is preserved
        // (changing it is optional and does not force an MFA reveal).
        this.isViewMode
          ? []
          : this.isEditMode
            ? [Validators.maxLength(20)]
            : [Validators.required, Validators.maxLength(20)],
      ],
      phone_number: [
        this.isViewMode ? maskPhone(this._realPhone) : this._realPhone,
        this.isViewMode ? [] : [Validators.required, Validators.maxLength(20)],
      ],
      email: [tenant?.email || '', this.isViewMode ? [] : [Validators.email]],
      address: [tenant?.address || ''],
      attachments: [],
    });

    if (this.isViewMode) {
      this.tenantForm.disable();
    }

    // Load existing attachments for view mode or edit
    if (!this.isEditMode && !this.data.tenant?.id && this.attachments.length > 0) {
      // This is create mode with pre-loaded attachments (e.g., from dialog data)
    } else if (tenant?.id || this.isViewMode) {
      // In edit or view mode, load existing attachments
      const attachmentIds: number[] = [];
      this.attachments.forEach((att: any) => {
        attachmentIds.push(att.id);
      });
      this.tenantForm.get('attachments')?.setValue(attachmentIds);
    }
  }

  handleUploadProgress(event: any): void {
    if (event.progress) {
      console.log('Upload progress:', event.progress, event.status);
    }
  }

  handleFileSelected(event: any): void {
    const file = event.file;
    console.log('File selected:', file);

    // Update the form control with attachment data
    const attachmentsData: any[] = [];

    // Check if there's an existing attachments value
    const currentAttachments = this.tenantForm.get('attachments')?.value || [];

    // Add new file info - we'll store the file object initially
    // The document component manages this, we just need to track it
    if (event.selected) {
      attachmentsData.push({
        id: null, // Will be set after upload completes
        file: event.file,
        status: 'uploading',
      });
    } else if (event.deleted) {
      // Remove the deleted attachment from our tracking
      const filtered = currentAttachments.filter((att: any) => att.id !== event.attachmentId);
      this.tenantForm.get('attachments')?.setValue(filtered);
    }

    this.tenantForm.get('attachments')?.setValue(attachmentsData);
  }

  handleFilesDeleted(event: any): void {
    const deletedIds = event.deletedIds;
    console.log('Files deleted:', deletedIds);

    // Update form control to remove deleted attachments
    const currentAttachments = this.tenantForm.get('attachments')?.value || [];
    const filtered = currentAttachments.filter((att: any) => !deletedIds.includes(att.id));
    this.tenantForm.get('attachments')?.setValue(filtered);
  }

  onSubmit() {
    if (this.tenantForm.valid) {
      const value = { ...this.tenantForm.value };

      // In edit mode a blank NID means "keep the existing value" — omit it so
      // the backend does not overwrite the stored NID with an empty string.
      if (this.isEditMode && !value.nid_number?.trim()) {
        delete value.nid_number;
      }

      // Handle attachments - extract attachment IDs or file data
      const attachmentsData = this.getAttachmentData();

      // Only include attachments if they exist and are not empty
      if (attachmentsData && attachmentsData.length > 0) {
        value.attachments = attachmentsData;
      }

      this.dialogRef.close(value);
    } else {
      Object.keys(this.tenantForm.controls).forEach((key) => {
        this.tenantForm.get(key)?.markAsTouched();
      });
    }
  }

  private getAttachmentData(): any[] | null {
    const attachments = this.tenantForm.get('attachments')?.value || [];

    if (!attachments || attachments.length === 0) {
      return null;
    }

    // Return attachments as-is (they may contain file objects or ID references)
    // The backend will handle both cases: existing attachment IDs or new file uploads
    return attachments;
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
