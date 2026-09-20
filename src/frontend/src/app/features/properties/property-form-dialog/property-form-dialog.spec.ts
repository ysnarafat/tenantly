import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { PropertyFormDialogComponent, PropertyFormDialogData } from './property-form-dialog';
import { Property } from '../../../core/models';
import { ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDialogModule } from '@angular/material/dialog';
import { TranslateModule } from '@ngx-translate/core';
import { CommonModule } from '@angular/common';

describe('PropertyFormDialogComponent - Address Validation', () => {
  let component: PropertyFormDialogComponent;
  let fixture: ComponentFixture<PropertyFormDialogComponent>;
  let mockDialogRef: jasmine.SpyObj<MatDialogRef<PropertyFormDialogComponent>>;

  const mockProperty: Property = {
    id: 1,
    property_name: 'Test Property',
    property_code: 'PROP001',
    property_type: 'Mixed',
    address: '123 Test St',
    city: 'Test City',
    active: true,
    total_buildings: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  beforeEach(async () => {
    mockDialogRef = jasmine.createSpyObj<MatDialogRef<PropertyFormDialogComponent>>(
      'MatDialogRef',
      ['close']
    );

    await TestBed.configureTestingModule({
      imports: [
        PropertyFormDialogComponent,
        CommonModule,
        ReactiveFormsModule,
        MatFormFieldModule,
        MatInputModule,
        MatSelectModule,
        MatButtonModule,
        MatIconModule,
        MatDialogModule,
        TranslateModule.forRoot(),
      ],
      providers: [
        { provide: MatDialogRef, useValue: mockDialogRef },
        { provide: MAT_DIALOG_DATA, useValue: {} },
      ],
    }).compileComponents();
  });

  function createComponent(data: PropertyFormDialogData) {
    fixture = TestBed.createComponent(PropertyFormDialogComponent);
    component = fixture.componentInstance;
    (component as any).data = data;
    component.ngOnInit();
    fixture.detectChanges();
  }

  describe('Create mode', () => {
    beforeEach(() => {
      createComponent({ mode: 'create' });
    });

    it('should reject an address longer than 500 characters', () => {
      const addressControl = component.propertyForm.get('address');
      addressControl?.setValue('a'.repeat(501));

      expect(addressControl?.hasError('maxlength')).toBe(true);
      expect(component.propertyForm.valid).toBe(false);
    });

    it('should accept an address exactly 500 characters long', () => {
      const addressControl = component.propertyForm.get('address');
      addressControl?.setValue('a'.repeat(500));

      expect(addressControl?.hasError('maxlength')).toBe(false);
    });

    it('should accept a normal address', () => {
      const addressControl = component.propertyForm.get('address');
      addressControl?.setValue('742 Evergreen Terrace, Springfield');

      expect(addressControl?.valid).toBe(true);
    });

    it('should reject an empty address (required)', () => {
      const addressControl = component.propertyForm.get('address');
      addressControl?.setValue('');

      expect(addressControl?.hasError('required')).toBe(true);
    });

    it('should show a max-length error message once touched', () => {
      const addressControl = component.propertyForm.get('address');
      addressControl?.setValue('a'.repeat(501));
      addressControl?.markAsTouched();

      expect(component.getErrorMessage('address')).toBe('Maximum length is 500');
    });

    it('should not submit the dialog when address exceeds max length', () => {
      component.propertyForm.patchValue({
        property_name: 'Valid Name',
        property_code: 'PROP-100',
        address: 'a'.repeat(501),
        property_type: 'Residential',
      });

      component.onSubmit();

      expect(mockDialogRef.close).not.toHaveBeenCalled();
    });
  });

  describe('Edit mode', () => {
    it('should validate max length against a pre-filled long address', () => {
      createComponent({
        mode: 'edit',
        property: { ...mockProperty, address: 'a'.repeat(501) },
      });

      const addressControl = component.propertyForm.get('address');
      expect(addressControl?.hasError('maxlength')).toBe(true);
      expect(component.propertyForm.valid).toBe(false);
    });
  });
});
