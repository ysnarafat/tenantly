import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { Building, BuildingType, Property } from '../../../core/models';

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
        MatCheckboxModule
    ],
    templateUrl: './building-form-dialog.html',
    styleUrls: ['./building-form-dialog.scss']
})
export class BuildingFormDialogComponent implements OnInit {
    private fb = inject(FormBuilder);
    private dialogRef = inject(MatDialogRef<BuildingFormDialogComponent>);
    public data = inject<BuildingFormDialogData>(MAT_DIALOG_DATA);

    buildingForm!: FormGroup;
    buildingTypes: BuildingType[] = ['Residential', 'Commercial', 'Mixed'];
    currentYear = new Date().getFullYear();

    ngOnInit() {
        this.initializeForm();
    }

    private initializeForm() {
        const building = this.data.building;

        this.buildingForm = this.fb.group({
            building_name: [
                building?.building_name || '',
                [Validators.required, Validators.minLength(3), Validators.maxLength(100)]
            ],
            building_code: [
                building?.building_code || this.generateBuildingCode(),
                [Validators.required, Validators.pattern(/^[A-Z0-9-]+$/)]
            ],
            building_type: [
                building?.building_type || 'Residential',
                [Validators.required]
            ],
            total_floors: [
                building?.total_floors || null,
                [Validators.min(1), Validators.max(200)]
            ],
            has_elevator: [
                building?.has_elevator || false
            ],
            construction_year: [
                building?.construction_year || null,
                [Validators.min(1800), Validators.max(this.currentYear + 5)]
            ]
        });

        // Disable building_code in edit mode
        if (this.data.mode === 'edit') {
            this.buildingForm.get('building_code')?.disable();
        }
    }

    private generateBuildingCode(): string {
        const timestamp = Date.now().toString().slice(-6);
        return `BLD-${timestamp}`;
    }

    onSubmit() {
        if (this.buildingForm.valid) {
            const formValue = this.buildingForm.getRawValue();

            // Remove building_code from updates (it's immutable)
            if (this.data.mode === 'edit') {
                const { building_code, ...updateData } = formValue;
                this.dialogRef.close(updateData);
            } else {
                // Add property_id for creation
                this.dialogRef.close({
                    ...formValue,
                    property_id: this.data.property.id
                });
            }
        } else {
            // Mark all fields as touched to show validation errors
            Object.keys(this.buildingForm.controls).forEach(key => {
                this.buildingForm.get(key)?.markAsTouched();
            });
        }
    }

    onCancel() {
        this.dialogRef.close();
    }

    getErrorMessage(fieldName: string): string {
        const control = this.buildingForm.get(fieldName);
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
        if (control.errors['min']) {
            return `Minimum value is ${control.errors['min'].min}`;
        }
        if (control.errors['max']) {
            return `Maximum value is ${control.errors['max'].max}`;
        }
        if (control.errors['pattern']) {
            if (fieldName === 'building_code') {
                return 'Only uppercase letters, numbers, and hyphens allowed';
            }
        }
        return 'Invalid value';
    }

    private getFieldLabel(fieldName: string): string {
        const labels: { [key: string]: string } = {
            building_name: 'Building Name',
            building_code: 'Building Code',
            building_type: 'Building Type',
            total_floors: 'Total Floors',
            has_elevator: 'Has Elevator',
            construction_year: 'Construction Year'
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
