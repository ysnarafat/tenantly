import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  FormBuilder,
  FormGroup,
  Validators,
  ReactiveFormsModule,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { TranslateModule } from '@ngx-translate/core';
import { Unit, UnitType, Building, Property, LeaseType } from '../../../core/models';
import { getUnitTypesForBuilding, getUnitTypeIcon } from '../unit-type.utils';

export interface UnitFormDialogData {
  unit?: Unit;
  building: Building;
  property: Property;
  mode: 'create' | 'edit';
}

@Component({
  selector: 'app-unit-form-dialog',
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
    TranslateModule,
  ],
  templateUrl: './unit-form-dialog.html',
  styleUrls: ['./unit-form-dialog.scss'],
})
export class UnitFormDialogComponent implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<UnitFormDialogComponent>);
  public data = inject<UnitFormDialogData>(MAT_DIALOG_DATA);

  unitForm!: FormGroup;
  allowedUnitTypes: UnitType[] = [];

  readonly leaseTypes: LeaseType[] = ['Residential', 'Commercial'];

  ngOnInit() {
    this.updateAllowedUnitTypes();
    this.initializeForm();
  }

  private updateAllowedUnitTypes() {
    this.allowedUnitTypes = getUnitTypesForBuilding(this.data.building.building_type);
  }

  private initializeForm() {
    const unit = this.data.unit;
    const defaultUnitType = this.allowedUnitTypes.includes(unit?.unit_type as UnitType)
      ? unit?.unit_type
      : this.allowedUnitTypes[0];

    this.unitForm = this.fb.group({
      unit_number: [unit?.unit_number || '', [Validators.required, Validators.maxLength(50)]],
      unit_name: [unit?.unit_name || '', [Validators.maxLength(100)]],
      unit_type: [
        defaultUnitType || 'Apartment',
        [Validators.required, this.unitTypeValidator.bind(this)],
      ],
      floor: [unit?.floor || null, [Validators.min(0), Validators.max(200)]],
      section: [unit?.section || '', [Validators.maxLength(50)]],

      // Lease defaults — optional. They only pre-fill the lease form later, so
      // an empty value here is a normal state, never a validation failure.
      default_lease_type: [unit?.default_lease_type ?? this.suggestedLeaseType()],
      default_monthly_rent: [unit?.default_monthly_rent ?? null, [Validators.min(0)]],
      default_security_deposit: [unit?.default_security_deposit ?? null, [Validators.min(0)]],
      // Capped at 60 months to match what the lease form accepts, so a stored
      // default is always usable there. The DB and API allow a wider range as an
      // outer sanity bound.
      default_duration_months: [
        unit?.default_duration_months ?? null,
        [Validators.min(1), Validators.max(60)],
      ],
    });

    // Disable unit_number in edit mode (it's the identifier)
    if (this.data.mode === 'edit') {
      this.unitForm.get('unit_number')?.disable();
    }
  }

  /**
   * Residential buildings hold residential tenancies and commercial ones hold
   * commercial tenancies, so seed the default from the building rather than
   * making the user pick the obvious answer. Mixed buildings get no guess.
   */
  private suggestedLeaseType(): LeaseType | null {
    switch (this.data.building.building_type) {
      case 'Residential':
        return 'Residential';
      case 'Commercial':
        return 'Commercial';
      default:
        return null;
    }
  }

  private unitTypeValidator(control: AbstractControl): ValidationErrors | null {
    if (!control.value) {
      return null;
    }
    const isAllowed = this.allowedUnitTypes.includes(control.value as UnitType);
    return isAllowed ? null : { invalidUnitType: { value: control.value } };
  }

  onSubmit() {
    if (this.unitForm.valid) {
      const formValue = this.stripEmptyLeaseDefaults(this.unitForm.getRawValue());

      // Remove unit_number from updates (it's immutable)
      if (this.data.mode === 'edit') {
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const { unit_number, ...updateData } = formValue;
        this.dialogRef.close(updateData);
      } else {
        // Add building_id and property_id for creation
        this.dialogRef.close({
          ...formValue,
          building_id: this.data.building.id,
          property_id: this.data.property.id,
        });
      }
    } else {
      // Mark all fields as touched to show validation errors
      Object.keys(this.unitForm.controls).forEach((key) => {
        this.unitForm.get(key)?.markAsTouched();
      });
    }
  }

  /**
   * Drops blank lease defaults from the payload. The API binds them as optional
   * pointers, so an omitted key means "no default"; sending an empty string or
   * null instead trips the oneof/gte validators.
   */
  private stripEmptyLeaseDefaults(formValue: Record<string, unknown>): Record<string, unknown> {
    const leaseDefaultKeys = [
      'default_lease_type',
      'default_monthly_rent',
      'default_security_deposit',
      'default_duration_months',
    ];
    const cleaned = { ...formValue };
    for (const key of leaseDefaultKeys) {
      if (cleaned[key] === null || cleaned[key] === '' || cleaned[key] === undefined) {
        delete cleaned[key];
      }
    }
    return cleaned;
  }

  onCancel() {
    this.dialogRef.close();
  }

  getErrorMessage(fieldName: string): string {
    const control = this.unitForm.get(fieldName);
    if (!control || !control.errors || !control.touched) {
      return '';
    }

    if (control.errors['required']) {
      return `${this.getFieldLabel(fieldName)} is required`;
    }
    if (control.errors['maxlength']) {
      return `Maximum length is ${control.errors['maxlength'].requiredLength}`;
    }
    if (control.errors['min']) {
      return `Minimum value is ${control.errors['min'].min}`;
    }
    if (control.errors['max']) {
      return `Maximum value is ${control.errors['max'].max}`;
    }
    if (control.errors['invalidUnitType']) {
      const buildingType = this.data.building.building_type;
      const allowedTypes = this.allowedUnitTypes.join(', ');
      return `${control.value} is not allowed in ${buildingType} buildings. Allowed types: ${allowedTypes}`;
    }
    return 'Invalid value';
  }

  private getFieldLabel(fieldName: string): string {
    const labels: Record<string, string> = {
      unit_number: 'Unit Number',
      unit_name: 'Unit Name',
      unit_type: 'Unit Type',
      floor: 'Floor',
      section: 'Section',
      default_monthly_rent: 'Default Monthly Rent',
      default_security_deposit: 'Default Security Deposit',
      default_duration_months: 'Default Lease Duration',
    };
    return labels[fieldName] || fieldName;
  }

  get isEditMode(): boolean {
    return this.data.mode === 'edit';
  }

  get dialogTitle(): string {
    return this.isEditMode ? 'Edit Unit' : 'Add New Unit';
  }

  get buildingContext(): string {
    return `${this.data.property.property_name} › ${this.data.building.building_name}`;
  }

  getUnitTypeIcon(type: UnitType): string {
    return getUnitTypeIcon(type);
  }
}
