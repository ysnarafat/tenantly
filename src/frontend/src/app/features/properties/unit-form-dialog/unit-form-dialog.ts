import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { Unit, UnitType, Building, Property } from '../../../core/models';

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
  ],
  templateUrl: './unit-form-dialog.html',
  styleUrls: ['./unit-form-dialog.scss'],
})
export class UnitFormDialogComponent implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<UnitFormDialogComponent>);
  public data = inject<UnitFormDialogData>(MAT_DIALOG_DATA);

  unitForm!: FormGroup;
  unitTypes: UnitType[] = ['Shop', 'Apartment', 'Office', 'Parking', 'Storage', 'Other'];

  ngOnInit() {
    this.initializeForm();
  }

  private initializeForm() {
    const unit = this.data.unit;

    this.unitForm = this.fb.group({
      unit_number: [unit?.unit_number || '', [Validators.required, Validators.maxLength(50)]],
      unit_name: [unit?.unit_name || '', [Validators.maxLength(100)]],
      unit_type: [unit?.unit_type || 'Apartment', [Validators.required]],
      floor: [unit?.floor || null, [Validators.min(0), Validators.max(200)]],
      section: [unit?.section || '', [Validators.maxLength(50)]],
      monthly_rent: [unit?.monthly_rent || null, [Validators.required, Validators.min(0)]],
    });

    // Disable unit_number in edit mode (it's the identifier)
    if (this.data.mode === 'edit') {
      this.unitForm.get('unit_number')?.disable();
    }
  }

  onSubmit() {
    if (this.unitForm.valid) {
      const formValue = this.unitForm.getRawValue();

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
    return 'Invalid value';
  }

  private getFieldLabel(fieldName: string): string {
    const labels: Record<string, string> = {
      unit_number: 'Unit Number',
      unit_name: 'Unit Name',
      unit_type: 'Unit Type',
      floor: 'Floor',
      section: 'Section',
      monthly_rent: 'Monthly Rent',
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
    switch (type) {
      case 'Shop':
        return 'store';
      case 'Apartment':
        return 'home';
      case 'Office':
        return 'business';
      case 'Parking':
        return 'local_parking';
      case 'Storage':
        return 'inventory_2';
      default:
        return 'meeting_room';
    }
  }
}
