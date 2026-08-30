import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { Building, BuildingType, CreateBuildingRequest, Property } from '../../../core/models';
import { BuildingService } from '../../../core/services/building.service';
import { notifyError } from '../../../shared/utils/notify.utils';

export interface BuildingFormDialogData {
  building?: Building;
  property: Property;
  mode: 'create' | 'edit';
}

@Component({
  selector: 'app-building-form-dialog',
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
    MatCheckboxModule,
    MatProgressSpinnerModule,
    TranslateModule,
  ],
  templateUrl: './building-form-dialog.html',
  styleUrls: ['./building-form-dialog.scss'],
})
export class BuildingFormDialogComponent implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<BuildingFormDialogComponent>);
  private buildingService = inject(BuildingService);
  private snackBar = inject(MatSnackBar);
  public data = inject<BuildingFormDialogData>(MAT_DIALOG_DATA);

  buildingForm!: FormGroup;
  buildingTypes: BuildingType[] = ['Residential', 'Commercial', 'Mixed'];
  currentYear = new Date().getFullYear();
  loading = signal(false);

  private codeUserEdited = false;

  ngOnInit() {
    this.initializeForm();

    if (!this.isEditMode) {
      this.buildingForm.get('building_name')!.valueChanges.subscribe((name: string) => {
        if (!this.codeUserEdited && name) {
          const code = this.generateBuildingCode(name);
          this.buildingForm.get('building_code')!.setValue(code, { emitEvent: false });
        }
      });

      this.buildingForm.get('building_code')!.valueChanges.subscribe(() => {
        this.codeUserEdited = true;
      });
    }
  }

  private initializeForm() {
    const building = this.data.building;

    this.buildingForm = this.fb.group({
      building_name: [
        building?.building_name || '',
        [Validators.required, Validators.minLength(3), Validators.maxLength(100)],
      ],
      building_code: [
        building?.building_code || '',
        [Validators.required, Validators.pattern(/^[A-Z0-9-]+$/)],
      ],
      building_type: [building?.building_type || 'Residential', [Validators.required]],
      total_floors: [building?.total_floors || null, [Validators.min(1), Validators.max(200)]],
      has_elevator: [building?.has_elevator || false],
      construction_year: [
        building?.construction_year || null,
        [Validators.min(1800), Validators.max(this.currentYear + 5)],
      ],
    });

    if (this.isEditMode) {
      this.buildingForm.get('building_code')?.disable();
    }
  }

  private generateBuildingCode(name: string): string {
    const initials = name
      .split(/\s+/)
      .filter((w) => w.length > 0)
      .map((w) => w[0].toUpperCase())
      .join('');
    return initials ? `${initials}-001` : '';
  }

  onSubmit() {
    if (this.buildingForm.invalid) {
      Object.keys(this.buildingForm.controls).forEach((key) => {
        this.buildingForm.get(key)?.markAsTouched();
      });
      return;
    }

    if (this.isEditMode) {
      const formValue = this.buildingForm.getRawValue();
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { building_code, ...updateData } = formValue;
      this.dialogRef.close(updateData);
      return;
    }

    const payload: CreateBuildingRequest = {
      ...this.buildingForm.getRawValue(),
      property_id: this.data.property.id,
    };

    this.loading.set(true);
    this.buildingService.createBuilding(payload).subscribe({
      next: (building) => {
        this.loading.set(false);
        this.dialogRef.close(building);
      },
      error: (err) => {
        this.loading.set(false);
        if (err?.error?.code === 'BUILDING_CODE_EXISTS') {
          const codeControl = this.buildingForm.get('building_code')!;
          codeControl.setErrors({
            serverError: 'Building code already exists — try a different one',
          });
          codeControl.markAsTouched();
          this.codeUserEdited = true;
        } else {
          notifyError(this.snackBar, 'Failed to create building');
        }
      },
    });
  }

  onCancel() {
    this.dialogRef.close();
  }

  getErrorMessage(fieldName: string): string {
    const control = this.buildingForm.get(fieldName);
    if (!control || !control.errors || !control.touched) return '';

    if (control.errors['serverError']) return control.errors['serverError'];
    if (control.errors['required']) return `${this.getFieldLabel(fieldName)} is required`;
    if (control.errors['minlength'])
      return `Minimum length is ${control.errors['minlength'].requiredLength}`;
    if (control.errors['maxlength'])
      return `Maximum length is ${control.errors['maxlength'].requiredLength}`;
    if (control.errors['min']) return `Minimum value is ${control.errors['min'].min}`;
    if (control.errors['max']) return `Maximum value is ${control.errors['max'].max}`;
    if (control.errors['pattern'] && fieldName === 'building_code')
      return 'Only uppercase letters, numbers, and hyphens allowed';
    return 'Invalid value';
  }

  private getFieldLabel(fieldName: string): string {
    const labels: Record<string, string> = {
      building_name: 'Building Name',
      building_code: 'Building Code',
      building_type: 'Building Type',
      total_floors: 'Total Floors',
      has_elevator: 'Has Elevator',
      construction_year: 'Construction Year',
    };
    return labels[fieldName] || fieldName;
  }

  get isEditMode(): boolean {
    return this.data.mode === 'edit';
  }

  get dialogTitle(): string {
    return this.isEditMode ? 'Edit Building' : 'Add New Building';
  }

  get propertyName(): string {
    return this.data.property.property_name;
  }
}
