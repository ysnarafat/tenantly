import { Component, ElementRef, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import {
  MatAutocompleteModule,
  MatAutocompleteSelectedEvent,
} from '@angular/material/autocomplete';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import {
  AbstractControl,
  ReactiveFormsModule,
  FormBuilder,
  FormControl,
  FormGroup,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import {
  LeaseService,
  CreateLeaseRequest,
  LeaseCustomFields,
} from '../../../core/services/lease.service';
import { PropertyService } from '../../../core/services/property.service';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import { TenantService } from '../../../core/services/tenant.service';
import { Property } from '../../../core/models/property.model';
import { Building } from '../../../core/models/building.model';
import { UnitWithDetails } from '../../../core/models/unit.model';
import { Tenant } from '../../../core/models/tenant.model';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';
import { LeaseChargesEditor, LeaseChargeDraft } from '../lease-charges-editor/lease-charges-editor';
import { LeaseCustomFieldsEditor } from '../lease-custom-fields-editor/lease-custom-fields-editor';

function integerValidator(control: AbstractControl): ValidationErrors | null {
  const value = control.value;
  if (value === null || value === undefined || value === '') return null;
  return Number.isInteger(Number(value)) ? null : { notInteger: true };
}

@Component({
  selector: 'app-create-lease-dialog',
  standalone: true,
  imports: [
    CommonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatAutocompleteModule,
    MatButtonModule,
    MatIconModule,
    MatDatepickerModule,
    MatProgressSpinnerModule,
    ReactiveFormsModule,
    TranslateModule,
    LeaseChargesEditor,
    LeaseCustomFieldsEditor,
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
  private translate = inject(TranslateService);
  private elRef: ElementRef<HTMLElement> = inject(ElementRef);

  leaseForm: FormGroup;
  tenantSearch = new FormControl<Tenant | string | null>('');

  loading = false;
  submitLoading = false;
  buildingsLoading = false;
  unitsLoading = false;

  properties: Property[] = [];
  buildings: Building[] = [];
  units: UnitWithDetails[] = [];
  tenants: Tenant[] = [];
  filteredTenants: Tenant[] = [];

  propertiesLoadFailed = false;
  tenantsLoadFailed = false;

  minDate: Date = new Date();

  // The lease doesn't exist yet, so charges/custom fields are held here as a
  // draft and only persisted once the lease itself is created — see onSubmit.
  pendingCharges: LeaseChargeDraft[] = [];
  customFields: LeaseCustomFields = {};

  constructor() {
    this.leaseForm = this.fb.group({
      tenant_id: [null, Validators.required],
      property_id: [null, Validators.required],
      building_id: [{ value: null, disabled: true }, Validators.required],
      unit_id: [{ value: null, disabled: true }, Validators.required],
      lease_type: ['Commercial', Validators.required],
      start_date: [new Date(), Validators.required],
      duration_months: [
        12,
        [Validators.required, Validators.min(1), Validators.max(60), integerValidator],
      ],
      monthly_rent: [null, [Validators.required, Validators.min(0)]],
      security_deposit: [null, [Validators.min(0)]],
    });
  }

  ngOnInit() {
    this.loadData();

    this.tenantSearch.valueChanges.subscribe((value) => {
      this.filterTenants(value);
      // A manual edit after a selection invalidates the previous choice —
      // otherwise a stale tenant_id could be submitted alongside new text.
      if (typeof value === 'string') {
        this.leaseForm.patchValue({ tenant_id: null });
      }
    });
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
    this.propertiesLoadFailed = false;
    return new Promise<void>((resolve) => {
      this.propertyService.getProperties({ active: true }).subscribe({
        next: (response) => {
          this.properties = response.properties;
          resolve();
        },
        error: (error: any) => {
          console.error('Error loading properties:', safeErrorMessage(error));
          this.propertiesLoadFailed = true;
          notifyError(
            this.snackBar,
            this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.LOAD_PROPERTIES')
          );
          resolve();
        },
      });
    });
  }

  loadTenants() {
    this.tenantsLoadFailed = false;
    return new Promise<void>((resolve) => {
      this.tenantService.getAllTenants(1, 100).subscribe({
        next: (response) => {
          this.tenants = response.tenants.filter((t) => t.active);
          this.filteredTenants = [...this.tenants];
          resolve();
        },
        error: (error) => {
          console.error('Error loading tenants:', safeErrorMessage(error));
          this.tenantsLoadFailed = true;
          notifyError(
            this.snackBar,
            this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.LOAD_TENANTS')
          );
          resolve();
        },
      });
    });
  }

  onPropertyChange(propertyId: number) {
    this.leaseForm.patchValue({ building_id: null, unit_id: null });
    this.buildings = [];
    this.units = [];
    this.leaseForm.get('building_id')?.disable();
    this.leaseForm.get('unit_id')?.disable();

    if (!propertyId) {
      return;
    }

    this.buildingsLoading = true;
    this.buildingService.getBuildingsByProperty(propertyId).subscribe({
      next: (response: any) => {
        this.buildings = response.buildings;
        this.buildingsLoading = false;
        if (this.buildings.length > 0) {
          this.leaseForm.get('building_id')?.enable();
        }
      },
      error: (error: any) => {
        console.error('Error loading buildings:', safeErrorMessage(error));
        this.buildingsLoading = false;
        notifyError(
          this.snackBar,
          this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.LOAD_BUILDINGS')
        );
      },
    });
  }

  // A unit already carrying an active lease can't be assigned another one
  // (enforced backend-side too — see LeaseService.CreateLease /
  // idx_leases_unit_active_unique) — disabled in the dropdown rather than
  // hidden, so it's still clear the unit exists, just unavailable right now.
  get allUnitsInBuildingLeased(): boolean {
    return this.units.length > 0 && this.units.every((u) => u.lease_active);
  }

  onBuildingChange(buildingId: number) {
    this.leaseForm.patchValue({ unit_id: null });
    this.units = [];
    this.leaseForm.get('unit_id')?.disable();

    if (!buildingId) {
      return;
    }

    this.unitsLoading = true;
    this.unitService.getUnitsByBuilding(buildingId).subscribe({
      next: (response: any) => {
        this.units = response.units;
        this.unitsLoading = false;
        if (this.units.length > 0) {
          this.leaseForm.get('unit_id')?.enable();
        }
      },
      error: (error: any) => {
        console.error('Error loading units:', safeErrorMessage(error));
        this.unitsLoading = false;
        notifyError(this.snackBar, this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.LOAD_UNITS'));
      },
    });
  }

  filterTenants(value: Tenant | string | null) {
    if (!value || typeof value !== 'string') {
      this.filteredTenants = [...this.tenants];
      return;
    }

    const searchTerm = value.toLowerCase();
    this.filteredTenants = this.tenants.filter(
      (tenant) =>
        tenant.name.toLowerCase().includes(searchTerm) ||
        tenant.email?.toLowerCase().includes(searchTerm) ||
        tenant.phone_number?.toLowerCase().includes(searchTerm)
    );
  }

  onTenantSelected(event: MatAutocompleteSelectedEvent) {
    const tenant = event.option.value as Tenant;
    this.leaseForm.patchValue({ tenant_id: tenant.id });
    this.leaseForm.get('tenant_id')?.markAsTouched();
  }

  onTenantSearchFocus(event: FocusEvent) {
    (event.target as HTMLInputElement).select();
  }

  displayTenantWith = (tenant: Tenant | string | null): string => {
    if (!tenant || typeof tenant === 'string') return typeof tenant === 'string' ? tenant : '';
    const noPhone = this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.NO_PHONE');
    return `${tenant.name} (${tenant.phone_number || noPhone})`;
  };

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
      notifyError(this.snackBar, this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.FORM_INVALID'));
      this.focusFirstInvalidField();
      return;
    }

    this.submitLoading = true;
    // Block accidental backdrop/Escape dismissal while the request is in
    // flight — closing now would abandon the request without the caller
    // ever knowing whether the lease was actually created.
    this.dialogRef.disableClose = true;

    const formValue = this.leaseForm.value;
    const request: CreateLeaseRequest = {
      tenant_id: formValue.tenant_id,
      unit_id: formValue.unit_id,
      lease_type: formValue.lease_type,
      start_date: this.formatDateForApi(formValue.start_date),
      duration_months: formValue.duration_months,
      monthly_rent: formValue.monthly_rent,
      security_deposit: formValue.security_deposit,
      custom_fields: this.customFields,
    };

    this.leaseService.createLease(request).subscribe({
      next: (lease) => {
        this.persistPendingCharges(lease.id);
      },
      error: (error) => {
        console.error('Error creating lease:', safeErrorMessage(error));
        const fallback = this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.CREATE_FAILED');
        const errorMessage = error.error?.error || error.message || fallback;
        notifyError(this.snackBar, errorMessage);
        this.submitLoading = false;
        this.dialogRef.disableClose = false;
      },
    });
  }

  // The lease has to exist before charges can be attached to it (they're
  // keyed by lease ID on the backend), so any charges added while building
  // the form are only sent now, after creation succeeds.
  private persistPendingCharges(leaseId: number): void {
    if (this.pendingCharges.length === 0) {
      notifySuccess(this.snackBar, this.translate.instant('CREATE_LEASE_DIALOG.SUCCESS.CREATED'));
      this.dialogRef.close(true);
      this.submitLoading = false;
      return;
    }

    const requests = this.pendingCharges.map((c) =>
      this.leaseService
        .addLeaseCharge(leaseId, { charge_type: c.charge_type, label: c.label, amount: c.amount })
        .pipe(catchError(() => of(null)))
    );

    forkJoin(requests).subscribe((results) => {
      const failed = results.filter((r) => r === null).length;
      if (failed > 0) {
        notifyError(
          this.snackBar,
          `Lease created, but ${failed} charge(s) failed to save — add them from the lease's edit dialog`
        );
      } else {
        notifySuccess(this.snackBar, this.translate.instant('CREATE_LEASE_DIALOG.SUCCESS.CREATED'));
      }
      this.dialogRef.close(true);
      this.submitLoading = false;
    });
  }

  onCancel() {
    if (this.submitLoading) return;
    this.dialogRef.close(false);
  }

  formatDateForApi(date: Date): string {
    return date.toISOString().split('T')[0];
  }

  getErrorMessage(controlName: string): string {
    const control = this.leaseForm.get(controlName);
    if (control?.errors) {
      if (control.errors['required'])
        return this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.REQUIRED');
      if (control.errors['min'])
        return this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.MIN_VALUE');
      if (control.errors['max'])
        return this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.MAX_VALUE');
      if (control.errors['notInteger'])
        return this.translate.instant('CREATE_LEASE_DIALOG.ERRORS.NOT_INTEGER');
    }
    return '';
  }

  private focusFirstInvalidField() {
    setTimeout(() => {
      const invalidEl = this.elRef.nativeElement.querySelector<HTMLElement>(
        '.lease-form .ng-invalid input, .lease-form .ng-invalid mat-select'
      );
      invalidEl?.scrollIntoView({ behavior: 'smooth', block: 'center' });
      invalidEl?.focus();
    });
  }
}
