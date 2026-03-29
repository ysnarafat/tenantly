import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TenantFormDialogComponent } from './tenant-form-dialog';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { Tenant } from '../../../core/models/tenant.model';

describe('TenantFormDialogComponent', () => {
  let component: TenantFormDialogComponent;
  let fixture: ComponentFixture<TenantFormDialogComponent>;
  let dialogRefSpy: jasmine.SpyObj<MatDialogRef<TenantFormDialogComponent>>;

  const mockTenant: Tenant = {
    id: 1,
    name: 'John Doe',
    tenant_type: 'Individual',
    email: 'john@example.com',
    phone_number: '1234567890',
    nid_number: 'NID123',
    address: '123 Main St',
    active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  beforeEach(async () => {
    dialogRefSpy = jasmine.createSpyObj('MatDialogRef', ['close']);

    await TestBed.configureTestingModule({
      imports: [
        TenantFormDialogComponent,
        ReactiveFormsModule,
        MatDialogModule,
        NoopAnimationsModule,
      ],
      providers: [
        FormBuilder,
        { provide: MatDialogRef, useValue: dialogRefSpy },
        { provide: MAT_DIALOG_DATA, useValue: { mode: 'create' } },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(TenantFormDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize form in create mode', () => {
    expect(component.tenantForm.get('name')?.value).toBe('');
    expect(component.tenantForm.get('tenant_type')?.value).toBe('Individual');
    expect(component.isEditMode).toBeFalse();
    expect(component.dialogTitle).toBe('Add New Tenant');
  });

  it('should initialize form in edit mode with data', () => {
    // Re-configure for edit mode
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [
        TenantFormDialogComponent,
        ReactiveFormsModule,
        MatDialogModule,
        NoopAnimationsModule,
      ],
      providers: [
        FormBuilder,
        { provide: MatDialogRef, useValue: dialogRefSpy },
        { provide: MAT_DIALOG_DATA, useValue: { mode: 'edit', tenant: mockTenant } },
      ],
    });

    const editFixture = TestBed.createComponent(TenantFormDialogComponent);
    const editComponent = editFixture.componentInstance;
    editFixture.detectChanges();

    expect(editComponent.tenantForm.get('name')?.value).toBe(mockTenant.name);
    expect(editComponent.tenantForm.get('email')?.value).toBe(mockTenant.email);
    expect(editComponent.isEditMode).toBeTrue();
    expect(editComponent.dialogTitle).toBe('Edit Tenant');
  });

  it('should mark form as invalid when required fields are empty', () => {
    component.tenantForm.patchValue({
      name: '',
      tenant_type: '',
    });
    expect(component.tenantForm.valid).toBeFalse();
  });

  it('should validate email format', () => {
    const emailControl = component.tenantForm.get('email');
    emailControl?.setValue('invalid-email');
    expect(emailControl?.valid).toBeFalse();
    expect(emailControl?.errors?.['email']).toBeTruthy();

    emailControl?.setValue('valid@example.com');
    expect(emailControl?.valid).toBeTrue();
  });

  it('should close dialog with form value on valid submit', () => {
    const formData = {
      name: 'Jane Doe',
      tenant_type: 'Individual',
      email: 'jane@example.com',
      phone_number: '9876543210',
      nid_number: 'NID456',
      address: '456 Elm St',
    };

    component.tenantForm.patchValue(formData);
    component.onSubmit();

    expect(dialogRefSpy.close).toHaveBeenCalledWith(jasmine.objectContaining(formData));
  });

  it('should not close dialog on invalid submit', () => {
    component.tenantForm.patchValue({ name: '' });
    component.onSubmit();

    expect(dialogRefSpy.close).not.toHaveBeenCalled();
    expect(component.tenantForm.get('name')?.touched).toBeTrue();
  });
});
