import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { UnitFormDialogComponent, UnitFormDialogData } from './unit-form-dialog';
import { Building, Property, Unit } from '../../../core/models';
import { ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDialogModule } from '@angular/material/dialog';
import { TranslateModule } from '@ngx-translate/core';
import { CommonModule } from '@angular/common';

describe('UnitFormDialogComponent - Unit Type Validation', () => {
  let component: UnitFormDialogComponent;
  let fixture: ComponentFixture<UnitFormDialogComponent>;
  let mockDialogRef: jasmine.SpyObj<MatDialogRef<UnitFormDialogComponent>>;

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

  const mockResidentialBuilding: Building = {
    id: 1,
    property_id: 1,
    building_name: 'Residential Building',
    building_code: 'RES-001',
    building_type: 'Residential',
    total_floors: 5,
    has_elevator: true,
    construction_year: 2020,
    active_status: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  const mockCommercialBuilding: Building = {
    ...mockResidentialBuilding,
    id: 2,
    building_name: 'Commercial Building',
    building_code: 'COM-001',
    building_type: 'Commercial',
  };

  const mockMixedBuilding: Building = {
    ...mockResidentialBuilding,
    id: 3,
    building_name: 'Mixed Building',
    building_code: 'MIX-001',
    building_type: 'Mixed',
  };

  beforeEach(async () => {
    mockDialogRef = jasmine.createSpyObj<MatDialogRef<UnitFormDialogComponent>>('MatDialogRef', [
      'close',
    ]);

    await TestBed.configureTestingModule({
      imports: [
        UnitFormDialogComponent,
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
      providers: [{ provide: MatDialogRef, useValue: mockDialogRef }],
    }).compileComponents();
  });

  describe('Unit Type Filtering by Building Type', () => {
    it('should show only Apartment, Parking, Storage for Residential buildings', () => {
      const dialogData: UnitFormDialogData = {
        building: mockResidentialBuilding,
        property: mockProperty,
        mode: 'create',
      };
      TestBed.inject(MAT_DIALOG_DATA);

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;

      component.ngOnInit();

      expect(component.allowedUnitTypes).toEqual(['Apartment', 'Parking', 'Storage']);
      expect(component.allowedUnitTypes.length).toBe(3);
    });

    it('should show only Shop, Office, Parking, Storage for Commercial buildings', () => {
      const dialogData: UnitFormDialogData = {
        building: mockCommercialBuilding,
        property: mockProperty,
        mode: 'create',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;

      component.ngOnInit();

      expect(component.allowedUnitTypes).toEqual(['Shop', 'Office', 'Parking', 'Storage']);
      expect(component.allowedUnitTypes.length).toBe(4);
    });

    it('should show all unit types (Shop, Apartment, Office, Parking, Storage, Other) for Mixed buildings', () => {
      const dialogData: UnitFormDialogData = {
        building: mockMixedBuilding,
        property: mockProperty,
        mode: 'create',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;

      component.ngOnInit();

      expect(component.allowedUnitTypes).toEqual([
        'Shop',
        'Apartment',
        'Office',
        'Parking',
        'Storage',
        'Other',
      ]);
      expect(component.allowedUnitTypes.length).toBe(6);
    });
  });

  describe('Form Validation', () => {
    beforeEach(() => {
      const dialogData: UnitFormDialogData = {
        building: mockResidentialBuilding,
        property: mockProperty,
        mode: 'create',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;
      component.ngOnInit();
      fixture.detectChanges();
    });

    it('should reject invalid unit types for Residential building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Shop');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(true);
    });

    it('should accept valid unit types for Residential building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Apartment');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(false);
    });

    it('should reject Office type for Residential building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Office');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(true);
    });

    it('should accept Parking for Residential building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Parking');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(false);
    });

    it('should accept Storage for Residential building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Storage');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(false);
    });
  });

  describe('Commercial Building Unit Type Validation', () => {
    beforeEach(() => {
      const dialogData: UnitFormDialogData = {
        building: mockCommercialBuilding,
        property: mockProperty,
        mode: 'create',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;
      component.ngOnInit();
      fixture.detectChanges();
    });

    it('should reject Apartment for Commercial building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Apartment');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(true);
    });

    it('should accept Shop for Commercial building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Shop');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(false);
    });

    it('should accept Office for Commercial building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Office');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(false);
    });

    it('should reject Other for Commercial building', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Other');

      expect(unitTypeControl?.hasError('invalidUnitType')).toBe(true);
    });
  });

  describe('Error Messages', () => {
    beforeEach(() => {
      const dialogData: UnitFormDialogData = {
        building: mockCommercialBuilding,
        property: mockProperty,
        mode: 'create',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;
      component.ngOnInit();
      fixture.detectChanges();
    });

    it('should show appropriate error message for invalid unit type', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Apartment');
      unitTypeControl?.markAsTouched();

      const errorMessage = component.getErrorMessage('unit_type');

      expect(errorMessage).toContain('Apartment');
      expect(errorMessage).toContain('Commercial');
      expect(errorMessage).toContain('Allowed types:');
    });

    it('should include all allowed types in error message', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Apartment');
      unitTypeControl?.markAsTouched();

      const errorMessage = component.getErrorMessage('unit_type');

      component.allowedUnitTypes.forEach((type) => {
        expect(errorMessage).toContain(type);
      });
    });

    it('should not show error message for untouched field', () => {
      const unitTypeControl = component.unitForm.get('unit_type');
      unitTypeControl?.setValue('Apartment');

      const errorMessage = component.getErrorMessage('unit_type');

      expect(errorMessage).toBe('');
    });

    it('should return empty string for null control', () => {
      const errorMessage = component.getErrorMessage('nonexistent_field');

      expect(errorMessage).toBe('');
    });
  });

  describe('Edit Mode Behavior', () => {
    it('should use existing unit type if valid for building type in edit mode', () => {
      const existingUnit: Unit = {
        id: 1,
        building_id: 1,
        property_id: 1,
        unit_number: '101',
        unit_name: 'Unit 101',
        floor: 1,
        section: 'A',
        unit_type: 'Apartment',
        metadata: {},
        active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      const dialogData: UnitFormDialogData = {
        unit: existingUnit,
        building: mockResidentialBuilding,
        property: mockProperty,
        mode: 'edit',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;
      component.ngOnInit();
      fixture.detectChanges();

      const unitTypeControl = component.unitForm.get('unit_type');
      expect(unitTypeControl?.value).toBe('Apartment');
    });

    it('should fallback to first allowed type if existing unit type is invalid for building type', () => {
      const existingUnit: Unit = {
        id: 1,
        building_id: 2,
        property_id: 1,
        unit_number: '101',
        unit_name: 'Unit 101',
        floor: 1,
        section: 'A',
        unit_type: 'Apartment',
        metadata: {},
        active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      const dialogData: UnitFormDialogData = {
        unit: existingUnit,
        building: mockCommercialBuilding,
        property: mockProperty,
        mode: 'edit',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;
      component.ngOnInit();
      fixture.detectChanges();

      const unitTypeControl = component.unitForm.get('unit_type');
      expect(unitTypeControl?.value).toBe(component.allowedUnitTypes[0]);
    });
  });

  describe('getUnitTypesForBuilding Helper Method', () => {
    beforeEach(() => {
      const dialogData: UnitFormDialogData = {
        building: mockResidentialBuilding,
        property: mockProperty,
        mode: 'create',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;
    });

    it('should return correct types for Residential', () => {
      const types = component.getUnitTypesForBuilding('Residential');
      expect(types).toEqual(['Apartment', 'Parking', 'Storage']);
    });

    it('should return correct types for Commercial', () => {
      const types = component.getUnitTypesForBuilding('Commercial');
      expect(types).toEqual(['Shop', 'Office', 'Parking', 'Storage']);
    });

    it('should return correct types for Mixed', () => {
      const types = component.getUnitTypesForBuilding('Mixed');
      expect(types).toEqual(['Shop', 'Apartment', 'Office', 'Parking', 'Storage', 'Other']);
    });
  });

  describe('Form Submission', () => {
    beforeEach(() => {
      const dialogData: UnitFormDialogData = {
        building: mockResidentialBuilding,
        property: mockProperty,
        mode: 'create',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;
      component.ngOnInit();
      fixture.detectChanges();
    });

    it('should not submit form with invalid unit type', () => {
      component.unitForm.patchValue({
        unit_number: '101',
        unit_type: 'Shop',
        monthly_rent: 1000,
      });

      component.onSubmit();

      expect(mockDialogRef.close).not.toHaveBeenCalled();
    });

    it('should submit form with valid unit type', () => {
      component.unitForm.patchValue({
        unit_number: '101',
        unit_type: 'Apartment',
        monthly_rent: 1000,
      });

      component.onSubmit();

      expect(mockDialogRef.close).toHaveBeenCalled();
    });
  });

  describe('UI Icons and Display', () => {
    beforeEach(() => {
      const dialogData: UnitFormDialogData = {
        building: mockResidentialBuilding,
        property: mockProperty,
        mode: 'create',
      };

      fixture = TestBed.createComponent(UnitFormDialogComponent);
      component = fixture.componentInstance;
      (component as any).data = dialogData;
      component.ngOnInit();
      fixture.detectChanges();
    });

    it('should return correct icon for Shop type', () => {
      expect(component.getUnitTypeIcon('Shop')).toBe('store');
    });

    it('should return correct icon for Apartment type', () => {
      expect(component.getUnitTypeIcon('Apartment')).toBe('home');
    });

    it('should return correct icon for Office type', () => {
      expect(component.getUnitTypeIcon('Office')).toBe('business');
    });

    it('should return correct icon for Parking type', () => {
      expect(component.getUnitTypeIcon('Parking')).toBe('local_parking');
    });

    it('should return correct icon for Storage type', () => {
      expect(component.getUnitTypeIcon('Storage')).toBe('inventory_2');
    });

    it('should return default icon for Other type', () => {
      expect(component.getUnitTypeIcon('Other')).toBe('meeting_room');
    });
  });
});
