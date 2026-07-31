import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { LeaseService, CreateLeaseRequest } from '../../../core/services/lease.service';
import { PropertyService } from '../../../core/services/property.service';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import { TenantService } from '../../../core/services/tenant.service';
import { Property } from '../../../core/models/property.model';
import { Building } from '../../../core/models/building.model';
import { Unit } from '../../../core/models/unit.model';
import { Tenant } from '../../../core/models/tenant.model';
import { LeaseType } from '../../../core/models/lease.model';
import { safeErrorMessage } from '../../../shared/utils/error.utils';

@Component({
  selector: 'app-create-lease-dialog',
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
  templateUrl: './create-lease-dialog.html',
  styleUrls: ['./create-lease-dialog.scss'],
})
export class CreateLeaseDialog implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<CreateLeaseDialog>);
  private leaseService = inject(LeaseService);
  private propertyService = inject(PropertyService);
  private buildingService = inject(BuildingService);
  private unitService = inject(UnitService);
  private tenantService = inject(TenantService);
  private snackBar = inject(MatSnackBar);

  leaseForm: FormGroup;
  loading = false;
  submitLoading = false;

  properties: Property[] = [];
  buildings: Building[] = [];
  units: Unit[] = [];
  tenants: Tenant[] = [];

  filteredTenants: Tenant[] = [];
  selectedTenantInput = '';

  minDate: Date = new Date();

  constructor() {
    this.leaseForm = this.fb.group({
      tenant_id: [null, Validators.required],
      property_id: [null, Validators.required],
      building_id: [{ value: null, disabled: true }, Validators.required],
      unit_id: [{ value: null, disabled: true }, Validators.required],
      lease_type: ['Commercial', Validators.required],
      start_date: [new Date(), Validators.required],
      duration_months: [12, [Validators.required, Validators.min(1), Validators.max(60)]],
      monthly_rent: [null, [Validators.required, Validators.min(0)]],
      security_deposit: [null, [Validators.min(0)]],
    });
  }

  ngOnInit() {
    this.loadData();
  }

  onTenantSelectOpened() {
    // Reset filtered tenants when dropdown opens
    this.filteredTenants = [...this.tenants];
  }

  compareById(option1: any, option2: any): boolean {
    // Handle null/undefined values
    if (!option1 && !option2) return true;
    if (!option1 || !option2) return false;
    // Convert both to strings for comparison to handle string/number mismatches
    return String(option1) === String(option2);
  }

  loadData() {
    this.loading = true;

    // Load all data in parallel
    Promise.all([this.loadProperties(), this.loadTenants()]).finally(() => {
      this.loading = false;
    });
  }

  loadProperties() {
    return new Promise<void>((resolve) => {
      this.propertyService.getProperties({ active: true }).subscribe({
        next: (response) => {
          this.properties = response.properties;
          resolve();
        },
        error: (error: any) => {
          console.error('Error loading properties:', safeErrorMessage(error));
          this.snackBar.open('Error loading properties', 'Close', { duration: 3000 });
          resolve();
        },
      });
    });
  }

  loadTenants() {
    return new Promise<void>((resolve) => {
      this.tenantService.getAllTenants(1, 100).subscribe({
        next: (response) => {
          this.tenants = response.tenants.filter((t) => t.active);
          this.filteredTenants = [...this.tenants];
          resolve();
        },
        error: (error) => {
          console.error('Error loading tenants:', safeErrorMessage(error));
          this.snackBar.open('Error loading tenants', 'Close', { duration: 3000 });
          resolve();
        },
      });
    });
  }

  onPropertyChange(propertyId: number) {
    this.leaseForm.patchValue({ building_id: null, unit_id: null });
    this.buildings = [];
    this.units = [];

    if (!propertyId) {
      this.leaseForm.get('building_id')?.disable();
      this.leaseForm.get('unit_id')?.disable();
      return;
    }

    this.buildingService.getBuildingsByProperty(propertyId).subscribe({
      next: (response: any) => {
        this.buildings = response.buildings;
        if (this.buildings.length > 0) {
          this.leaseForm.get('building_id')?.enable();
        } else {
          this.leaseForm.get('building_id')?.disable();
        }
      },
      error: (error: any) => {
        console.error('Error loading buildings:', safeErrorMessage(error));
        this.snackBar.open('Error loading buildings', 'Close', { duration: 3000 });
        this.leaseForm.get('building_id')?.disable();
      },
    });
  }

  onBuildingChange(buildingId: number) {
    this.leaseForm.patchValue({ unit_id: null });
    this.units = [];

    if (!buildingId) {
      this.leaseForm.get('unit_id')?.disable();
      return;
    }

    this.unitService.getUnitsByBuilding(buildingId).subscribe({
      next: (response: any) => {
        this.units = response.units;
        if (this.units.length > 0) {
          this.leaseForm.get('unit_id')?.enable();
        } else {
          this.leaseForm.get('unit_id')?.disable();
        }
      },
      error: (error: any) => {
        console.error('Error loading units:', safeErrorMessage(error));
        this.snackBar.open('Error loading units', 'Close', { duration: 3000 });
        this.leaseForm.get('unit_id')?.disable();
      },
    });
  }

  filterTenants(value: any) {
    if (!value) {
      this.filteredTenants = [...this.tenants];
      return;
    }

    // Handle case where value is a Tenant object (when selected from dropdown)
    if (typeof value === 'object' && value.id) {
      this.filteredTenants = [...this.tenants];
      return;
    }

    // Handle case where value is a string (when typing)
    const searchTerm = value.toString().toLowerCase();
    this.filteredTenants = this.tenants.filter(
      (tenant) =>
        tenant.name.toLowerCase().includes(searchTerm) ||
        tenant.email?.toLowerCase().includes(searchTerm) ||
        tenant.phone_number?.toLowerCase().includes(searchTerm)
    );
  }

  displayTenantWith(tenantId: number | string | null): string {
    if (!tenantId) return '';
    const id = typeof tenantId === 'string' ? Number(tenantId) : tenantId;
    const tenant = this.tenants.find((t) => t.id === id);
    return tenant ? `${tenant.name} (${tenant.phone_number || 'No phone'})` : '';
  }

  displayBuildingWith(buildingId: number | string | null): string {
    if (!buildingId) return '';
    const id = typeof buildingId === 'string' ? Number(buildingId) : buildingId;
    const building = this.buildings.find((b) => b.id === id);
    return building ? `${building.building_name} (${building.building_code})` : '';
  }

  displayUnitWith(unitId: number | string | null): string {
    if (!unitId) return '';
    const id = typeof unitId === 'string' ? Number(unitId) : unitId;
    const unit = this.units.find((u) => u.id === id);
    return unit ? `Unit ${unit.unit_number} - ${unit.unit_type}` : '';
  }

  calculateEndDate() {
    const startDate = this.leaseForm.get('start_date')?.value;
    const durationMonths = this.leaseForm.get('duration_months')?.value;

    if (startDate && durationMonths) {
      const endDate = new Date(startDate);
      endDate.setMonth(endDate.getMonth() + durationMonths);
      return endDate;
    }
    return null;
  }

  onSubmit() {
    if (this.leaseForm.invalid) {
      this.leaseForm.markAllAsTouched();
      this.snackBar.open('Please fill all required fields', 'Close', { duration: 3000 });
      return;
    }

    this.submitLoading = true;

    const formValue = this.leaseForm.value;
    const request: CreateLeaseRequest = {
      tenant_id: formValue.tenant_id,
      unit_id: formValue.unit_id,
      lease_type: formValue.lease_type,
      start_date: this.formatDateForApi(formValue.start_date),
      duration_months: formValue.duration_months,
      monthly_rent: formValue.monthly_rent,
      security_deposit: formValue.security_deposit,
    };

    this.leaseService.createLease(request).subscribe({
      next: () => {
        this.snackBar.open('Lease created successfully', 'Close', { duration: 3000 });
        this.dialogRef.close(true);
        this.submitLoading = false;
      },
      error: (error) => {
        console.error('Error creating lease:', safeErrorMessage(error));
        const errorMessage = error.error?.message || 'Failed to create lease';
        this.snackBar.open(errorMessage, 'Close', { duration: 5000 });
        this.submitLoading = false;
      },
    });
  }

  onCancel() {
    this.dialogRef.close(false);
  }

  formatDateForApi(date: Date): string {
    return date.toISOString().split('T')[0];
  }

  getErrorMessage(controlName: string): string {
    const control = this.leaseForm.get(controlName);
    if (control?.errors) {
      if (control.errors['required']) return 'This field is required';
      if (control.errors['min']) return 'Value must be positive';
      if (control.errors['max']) return 'Value is too large';
    }
    return '';
  }
}
