import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap } from '@angular/router';
import { MatDialog, MatDialogRef } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { of, throwError } from 'rxjs';
import { PropertyDetail } from './property-detail';
import { PropertyService } from '../../../core/services/property.service';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import {
  PropertyWithStats,
  Building,
  UnitWithDetails,
  UpdatePropertyRequest,
} from '../../../core/models';

function dialogRefStub(result: unknown): MatDialogRef<unknown> {
  return { afterClosed: () => of(result) } as unknown as MatDialogRef<unknown>;
}

function makeProperty(overrides: Partial<PropertyWithStats> = {}): PropertyWithStats {
  return {
    id: 42,
    property_name: 'Sunrise Apartments',
    property_code: 'PROP-042',
    address: '12 Gulshan Ave',
    property_type: 'Residential',
    total_buildings: 2,
    active: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    building_count: 2,
    unit_count: 20,
    occupied_units: 15,
    total_revenue: 450000,
    ...overrides,
  };
}

function makeBuilding(overrides: Partial<Building> = {}): Building {
  return {
    id: 7,
    property_id: 42,
    building_name: 'Block A',
    building_code: 'BLK-A',
    building_type: 'Residential',
    has_elevator: true,
    active_status: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeUnit(overrides: Partial<UnitWithDetails> = {}): UnitWithDetails {
  return {
    id: 1,
    building_id: 7,
    property_id: 42,
    unit_number: '3B',
    unit_type: 'Apartment',
    active: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    property_name: 'Sunrise Apartments',
    building_name: 'Block A',
    building_code: 'BLK-A',
    lease_active: true,
    ...overrides,
  };
}

describe('PropertyDetail', () => {
  let component: PropertyDetail;
  let fixture: ComponentFixture<PropertyDetail>;
  let propertyService: jasmine.SpyObj<PropertyService>;
  let buildingService: jasmine.SpyObj<BuildingService>;
  let unitService: jasmine.SpyObj<UnitService>;
  let dialog: jasmine.SpyObj<MatDialog>;
  let snackBar: jasmine.SpyObj<MatSnackBar>;
  let router: jasmine.SpyObj<Router>;

  beforeEach(async () => {
    const propertyServiceSpy = jasmine.createSpyObj('PropertyService', [
      'getPropertyWithStats',
      'updateProperty',
      'deleteProperty',
    ]);
    const buildingServiceSpy = jasmine.createSpyObj('BuildingService', ['getBuildingsByProperty']);
    const unitServiceSpy = jasmine.createSpyObj('UnitService', ['getUnitsByBuilding']);
    const dialogSpy = jasmine.createSpyObj('MatDialog', ['open']);
    const snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    buildingServiceSpy.getBuildingsByProperty.and.returnValue(
      of({
        buildings: [],
        pagination: {
          current_page: 1,
          page_size: 20,
          total_items: 0,
          total_pages: 1,
          has_next: false,
          has_prev: false,
        },
      })
    );
    unitServiceSpy.getUnitsByBuilding.and.returnValue(
      of({
        units: [],
        pagination: {
          current_page: 1,
          page_size: 20,
          total_items: 0,
          total_pages: 1,
          has_next: false,
          has_prev: false,
        },
      })
    );

    await TestBed.configureTestingModule({
      imports: [PropertyDetail, TranslateModule.forRoot()],
      providers: [
        { provide: PropertyService, useValue: propertyServiceSpy },
        { provide: BuildingService, useValue: buildingServiceSpy },
        { provide: UnitService, useValue: unitServiceSpy },
        { provide: MatDialog, useValue: dialogSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
        { provide: Router, useValue: routerSpy },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id: '42' }) } },
        },
      ],
    }).compileComponents();

    propertyService = TestBed.inject(PropertyService) as jasmine.SpyObj<PropertyService>;
    buildingService = TestBed.inject(BuildingService) as jasmine.SpyObj<BuildingService>;
    unitService = TestBed.inject(UnitService) as jasmine.SpyObj<UnitService>;
    dialog = TestBed.inject(MatDialog) as jasmine.SpyObj<MatDialog>;
    snackBar = TestBed.inject(MatSnackBar) as jasmine.SpyObj<MatSnackBar>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;

    propertyService.getPropertyWithStats.and.returnValue(of(makeProperty()));

    fixture = TestBed.createComponent(PropertyDetail);
    component = fixture.componentInstance;
  });

  describe('Loading', () => {
    it('reads the property id from the route and fetches it', () => {
      fixture.detectChanges();

      expect(component.propertyId()).toBe(42);
      expect(propertyService.getPropertyWithStats).toHaveBeenCalledWith(42);
    });

    it('populates the property signal and then loads its buildings', () => {
      const property = makeProperty();
      propertyService.getPropertyWithStats.and.returnValue(of(property));

      fixture.detectChanges();

      expect(component.property()).toEqual(property);
      expect(component.loading()).toBeFalse();
      expect(buildingService.getBuildingsByProperty).toHaveBeenCalledWith(42);
    });

    it('populates the buildings signal once loaded', () => {
      const building = makeBuilding();
      buildingService.getBuildingsByProperty.and.returnValue(
        of({
          buildings: [building],
          pagination: {
            current_page: 1,
            page_size: 20,
            total_items: 1,
            total_pages: 1,
            has_next: false,
            has_prev: false,
          },
        })
      );

      fixture.detectChanges();

      expect(component.buildings()).toEqual([building]);
      expect(component.buildingsLoading()).toBeFalse();
    });

    it('sets notFound (not a generic error toast) on a 404', () => {
      propertyService.getPropertyWithStats.and.returnValue(throwError(() => ({ status: 404 })));

      fixture.detectChanges();

      expect(component.notFound()).toBeTrue();
      expect(component.loading()).toBeFalse();
      expect(snackBar.open).not.toHaveBeenCalled();
    });

    it('shows a generic error toast (not notFound) on a non-404 failure', () => {
      propertyService.getPropertyWithStats.and.returnValue(throwError(() => ({ status: 500 })));

      fixture.detectChanges();

      expect(component.notFound()).toBeFalse();
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('shows an error toast if buildings fail to load, without blocking the property view', () => {
      buildingService.getBuildingsByProperty.and.returnValue(throwError(() => ({ status: 500 })));

      fixture.detectChanges();

      expect(component.property()).toBeTruthy();
      expect(component.buildingsLoading()).toBeFalse();
      expect(snackBar.open).toHaveBeenCalled();
    });
  });

  describe('occupancyRate', () => {
    it('computes the rounded percentage of occupied units', () => {
      propertyService.getPropertyWithStats.and.returnValue(
        of(makeProperty({ unit_count: 20, occupied_units: 15 }))
      );

      fixture.detectChanges();

      expect(component.occupancyRate()).toBe(75);
    });

    it('is 0 rather than NaN/Infinity when there are no units yet', () => {
      propertyService.getPropertyWithStats.and.returnValue(
        of(makeProperty({ unit_count: 0, occupied_units: 0 }))
      );

      fixture.detectChanges();

      expect(component.occupancyRate()).toBe(0);
    });
  });

  describe('editProperty', () => {
    it('updates the property and reloads on confirmation', () => {
      fixture.detectChanges();
      const req: UpdatePropertyRequest = { property_name: 'Renamed' };
      dialog.open.and.returnValue(dialogRefStub(req));
      propertyService.updateProperty.and.returnValue(
        of(makeProperty({ property_name: 'Renamed' }))
      );
      propertyService.getPropertyWithStats.calls.reset();

      component.editProperty();

      expect(propertyService.updateProperty).toHaveBeenCalledWith(42, req);
      expect(propertyService.getPropertyWithStats).toHaveBeenCalledWith(42);
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('does nothing when the edit dialog is dismissed without a result', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(undefined));

      component.editProperty();

      expect(propertyService.updateProperty).not.toHaveBeenCalled();
    });

    it('shows an error toast when the update fails', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub({ property_name: 'Renamed' }));
      propertyService.updateProperty.and.returnValue(throwError(() => ({ status: 500 })));

      component.editProperty();

      expect(snackBar.open).toHaveBeenCalled();
    });
  });

  describe('deleteProperty', () => {
    it('deletes the property and navigates back to the list on confirmation', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(true));
      propertyService.deleteProperty.and.returnValue(of(undefined));

      component.deleteProperty();

      expect(propertyService.deleteProperty).toHaveBeenCalledWith(42);
      expect(router.navigate).toHaveBeenCalledWith(['/properties']);
    });

    it('does not delete when the confirmation dialog is cancelled', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(false));

      component.deleteProperty();

      expect(propertyService.deleteProperty).not.toHaveBeenCalled();
    });

    it('shows an error toast and does not navigate when deletion fails', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(true));
      propertyService.deleteProperty.and.returnValue(throwError(() => ({ status: 500 })));

      component.deleteProperty();

      expect(snackBar.open).toHaveBeenCalled();
      expect(router.navigate).not.toHaveBeenCalled();
    });
  });

  describe('viewBuilding', () => {
    it('navigates to the nested building detail route', () => {
      fixture.detectChanges();

      component.viewBuilding(makeBuilding({ id: 9 }));

      expect(router.navigate).toHaveBeenCalledWith(['/properties', 42, 'buildings', 9]);
    });
  });

  describe('toggleBuildingExpand', () => {
    it('expands a building and loads its units on first expand', () => {
      fixture.detectChanges();
      const building = makeBuilding({ id: 9 });
      const unit = makeUnit({ id: 1, building_id: 9 });
      unitService.getUnitsByBuilding.and.returnValue(
        of({
          units: [unit],
          pagination: {
            current_page: 1,
            page_size: 20,
            total_items: 1,
            total_pages: 1,
            has_next: false,
            has_prev: false,
          },
        })
      );

      component.toggleBuildingExpand(building);

      expect(component.isBuildingExpanded(building)).toBeTrue();
      expect(unitService.getUnitsByBuilding).toHaveBeenCalledWith(9);
      expect(component.unitsFor(building)).toEqual([unit]);
      expect(component.isBuildingUnitsLoading(building)).toBeFalse();
    });

    it('collapses an already-expanded building without refetching', () => {
      fixture.detectChanges();
      const building = makeBuilding({ id: 9 });

      component.toggleBuildingExpand(building);
      unitService.getUnitsByBuilding.calls.reset();
      component.toggleBuildingExpand(building);

      expect(component.isBuildingExpanded(building)).toBeFalse();
      expect(unitService.getUnitsByBuilding).not.toHaveBeenCalled();
    });

    it('does not refetch units on a second expand once already cached', () => {
      fixture.detectChanges();
      const building = makeBuilding({ id: 9 });

      component.toggleBuildingExpand(building); // expand (fetches)
      component.toggleBuildingExpand(building); // collapse
      unitService.getUnitsByBuilding.calls.reset();
      component.toggleBuildingExpand(building); // expand again

      expect(unitService.getUnitsByBuilding).not.toHaveBeenCalled();
      expect(component.isBuildingExpanded(building)).toBeTrue();
    });

    it('defaults units to an empty array when the backend returns null (Go nil-slice)', () => {
      fixture.detectChanges();
      const building = makeBuilding({ id: 9 });
      unitService.getUnitsByBuilding.and.returnValue(
        of({
          units: null as unknown as UnitWithDetails[],
          pagination: {
            current_page: 1,
            page_size: 20,
            total_items: 0,
            total_pages: 1,
            has_next: false,
            has_prev: false,
          },
        })
      );

      component.toggleBuildingExpand(building);

      expect(component.unitsFor(building)).toEqual([]);
    });

    it('shows an error toast and stops the loading state when unit loading fails', () => {
      fixture.detectChanges();
      const building = makeBuilding({ id: 9 });
      unitService.getUnitsByBuilding.and.returnValue(throwError(() => ({ status: 500 })));

      component.toggleBuildingExpand(building);

      expect(snackBar.open).toHaveBeenCalled();
      expect(component.isBuildingUnitsLoading(building)).toBeFalse();
    });

    it('tracks expand state and units independently per building', () => {
      fixture.detectChanges();
      const buildingA = makeBuilding({ id: 9 });
      const buildingB = makeBuilding({ id: 10 });
      const unitA = makeUnit({ id: 1, building_id: 9 });
      unitService.getUnitsByBuilding.and.callFake((id: number) =>
        of({
          units: id === 9 ? [unitA] : [],
          pagination: {
            current_page: 1,
            page_size: 20,
            total_items: id === 9 ? 1 : 0,
            total_pages: 1,
            has_next: false,
            has_prev: false,
          },
        })
      );

      component.toggleBuildingExpand(buildingA);

      expect(component.isBuildingExpanded(buildingA)).toBeTrue();
      expect(component.isBuildingExpanded(buildingB)).toBeFalse();
      expect(component.unitsFor(buildingA)).toEqual([unitA]);
      expect(component.unitsFor(buildingB)).toEqual([]);
    });
  });

  describe('viewUnitDetails', () => {
    it('opens the entity detail dialog and stops the click from bubbling to the building toggle', () => {
      fixture.detectChanges();
      const unit = makeUnit({ tenant_name: 'Rahim Uddin', lease_active: true });
      const event = new Event('click');
      const stopSpy = spyOn(event, 'stopPropagation');

      component.viewUnitDetails(unit, event);

      expect(stopSpy).toHaveBeenCalled();
      expect(dialog.open).toHaveBeenCalledWith(
        jasmine.anything(),
        jasmine.objectContaining({
          data: jasmine.objectContaining({
            title: unit.unit_number,
            rows: jasmine.arrayContaining([
              jasmine.objectContaining({ label: 'Tenant', value: 'Rahim Uddin' }),
              jasmine.objectContaining({ label: 'Lease Status', value: 'Active lease' }),
            ]),
          }),
        })
      );
    });

    it('shows "Vacant" and "No active lease" when the unit has no tenant', () => {
      fixture.detectChanges();
      const unit = makeUnit({ tenant_name: undefined, lease_active: false });

      component.viewUnitDetails(unit, new Event('click'));

      expect(dialog.open).toHaveBeenCalledWith(
        jasmine.anything(),
        jasmine.objectContaining({
          data: jasmine.objectContaining({
            rows: jasmine.arrayContaining([
              jasmine.objectContaining({ label: 'Tenant', value: 'Vacant' }),
              jasmine.objectContaining({ label: 'Lease Status', value: 'No active lease' }),
            ]),
          }),
        })
      );
    });
  });
});
