import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { Property, PropertyType } from '../../../core/models';

export interface PropertyFormDialogData {
  property?: Property;
  mode: 'create' | 'edit';
}

@Component({
  selector: 'app-property-form-dialog',
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
  templateUrl: './property-form-dialog.html',
  styleUrls: ['./property-form-dialog.scss'],
})
export class PropertyFormDialogComponent implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<PropertyFormDialogComponent>);
  public data = inject<PropertyFormDialogData>(MAT_DIALOG_DATA);

  propertyForm!: FormGroup;
  propertyTypes: PropertyType[] = ['Residential', 'Commercial', 'Mixed'];

  ngOnInit() {
    this.initializeForm();
  }

  private initializeForm() {
    const property = this.data.property;

    this.propertyForm = this.fb.group({
      property_name: [
        property?.property_name || '',
        [Validators.required, Validators.minLength(3), Validators.maxLength(100)],
      ],
      property_code: [
        property?.property_code || this.generatePropertyCode(),
        [Validators.required, Validators.pattern(/^[A-Z0-9-]+$/)],
      ],
      address: [property?.address || '', [Validators.required, Validators.minLength(5)]],
      city: [property?.city || '', [Validators.maxLength(50)]],
      postal_code: [property?.postal_code || '', [Validators.pattern(/^\d{4}$/)]],
      property_type: [property?.property_type || 'Residential', [Validators.required]],
    });

    // Disable property_code in edit mode
    if (this.data.mode === 'edit') {
      this.propertyForm.get('property_code')?.disable();
    }
  }

  private generatePropertyCode(): string {
    const timestamp = Date.now().toString().slice(-6);
    return `PROP-${timestamp}`;
  }

  onSubmit() {
    if (this.propertyForm.valid) {
      const formValue = this.propertyForm.getRawValue();

      // Remove property_code from updates (it's immutable)
      if (this.data.mode === 'edit') {
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const { property_code, ...updateData } = formValue;
        this.dialogRef.close(updateData);
      } else {
        this.dialogRef.close(formValue);
      }
    } else {
      // Mark all fields as touched to show validation errors
      Object.keys(this.propertyForm.controls).forEach((key) => {
        this.propertyForm.get(key)?.markAsTouched();
      });
    }
  }

  onCancel() {
    this.dialogRef.close();
  }

  getErrorMessage(fieldName: string): string {
    const control = this.propertyForm.get(fieldName);
    if (!control || !control.errors || !control.touched) {
      return '';
    }

    if (control.errors['required']) {
      return `${this.getFieldLabel(fieldName)} is required`;
    }
    if (control.errors['minlength']) {
      return `Minimum length is ${control.errors['minlength'].requiredLength}`;
    }
    if (control.errors['maxlength']) {
      return `Maximum length is ${control.errors['maxlength'].requiredLength}`;
    }
    if (control.errors['pattern']) {
      if (fieldName === 'property_code') {
        return 'Only uppercase letters, numbers, and hyphens allowed';
      }
      if (fieldName === 'postal_code') {
        return 'Must be a 4-digit postal code';
      }
    }
    return 'Invalid value';
  }

  private getFieldLabel(fieldName: string): string {
    const labels: Record<string, string> = {
      property_name: 'Property Name',
      property_code: 'Property Code',
      address: 'Address',
      city: 'City',
      postal_code: 'Postal Code',
      property_type: 'Property Type',
    };
    return labels[fieldName] || fieldName;
  }

  get isEditMode(): boolean {
    return this.data.mode === 'edit';
  }

  get dialogTitle(): string {
    return this.isEditMode ? 'Edit Property' : 'Add New Property';
  }
}
