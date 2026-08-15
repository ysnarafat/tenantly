import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap } from '@angular/router';
import { MatDialog, MatDialogRef } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { of, throwError } from 'rxjs';
import { BuildingDetail } from './building-detail';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import { BuildingWithStats, UnitWithDetails, UpdateBuildingRequest } from '../../../core/models';

function dialogRefStub(result: unknown): MatDialogRef<unknown> {
  return { afterClosed: () => of(result) } as unknown as MatDialogRef<unknown>;
}

function makeBuilding(overrides: Partial<BuildingWithStats> = {}): BuildingWithStats {
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
    property_name: 'Sunrise Apartments',
    unit_count: 10,
    occupied_units: 8,
    total_revenue: 200000,
    occupancy_rate: 80,
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

describe('BuildingDetail', () => {
  let component: BuildingDetail;
  let fixture: ComponentFixture<BuildingDetail>;
  let buildingService: jasmine.SpyObj<BuildingService>;
  let unitService: jasmine.SpyObj<UnitService>;
  let dialog: jasmine.SpyObj<MatDialog>;
  let snackBar: jasmine.SpyObj<MatSnackBar>;
  let router: jasmine.SpyObj<Router>;

  beforeEach(async () => {
    const buildingServiceSpy = jasmine.createSpyObj('BuildingService', [
      'getBuildingWithStats',
      'updateBuilding',
      'deleteBuilding',
    ]);
    const unitServiceSpy = jasmine.createSpyObj('UnitService', ['getUnitsByBuilding']);
    const dialogSpy = jasmine.createSpyObj('MatDialog', ['open']);
    const snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);

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
      imports: [BuildingDetail, TranslateModule.forRoot()],
      providers: [
        { provide: BuildingService, useValue: buildingServiceSpy },
        { provide: UnitService, useValue: unitServiceSpy },
        { provide: MatDialog, useValue: dialogSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
        { provide: Router, useValue: routerSpy },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: { paramMap: convertToParamMap({ propertyId: '42', buildingId: '7' }) },
          },
        },
      ],
    }).compileComponents();

    buildingService = TestBed.inject(BuildingService) as jasmine.SpyObj<BuildingService>;
    unitService = TestBed.inject(UnitService) as jasmine.SpyObj<UnitService>;
    dialog = TestBed.inject(MatDialog) as jasmine.SpyObj<MatDialog>;
    snackBar = TestBed.inject(MatSnackBar) as jasmine.SpyObj<MatSnackBar>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;

    buildingService.getBuildingWithStats.and.returnValue(of(makeBuilding()));

    fixture = TestBed.createComponent(BuildingDetail);
    component = fixture.componentInstance;
  });

  describe('Loading', () => {
    it('reads the property and building ids from the route and fetches the building', () => {
      fixture.detectChanges();

      expect(component.propertyId()).toBe(42);
      expect(component.buildingId()).toBe(7);
      expect(buildingService.getBuildingWithStats).toHaveBeenCalledWith(7);
    });

    it('populates the building signal and then loads its units', () => {
      const building = makeBuilding();
      buildingService.getBuildingWithStats.and.returnValue(of(building));

      fixture.detectChanges();

      expect(component.building()).toEqual(building);
      expect(component.loading()).toBeFalse();
      expect(unitService.getUnitsByBuilding).toHaveBeenCalledWith(7);
    });

    it('populates the units signal once loaded', () => {
      const unit = makeUnit();
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

      fixture.detectChanges();

      expect(component.units()).toEqual([unit]);
      expect(component.unitsLoading()).toBeFalse();
    });

    it('sets notFound (not a generic error toast) on a 404', () => {
      buildingService.getBuildingWithStats.and.returnValue(throwError(() => ({ status: 404 })));

      fixture.detectChanges();

      expect(component.notFound()).toBeTrue();
      expect(snackBar.open).not.toHaveBeenCalled();
    });

    it('shows a generic error toast (not notFound) on a non-404 failure', () => {
      buildingService.getBuildingWithStats.and.returnValue(throwError(() => ({ status: 500 })));

      fixture.detectChanges();

      expect(component.notFound()).toBeFalse();
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('shows an error toast if units fail to load, without blocking the building view', () => {
      unitService.getUnitsByBuilding.and.returnValue(throwError(() => ({ status: 500 })));

      fixture.detectChanges();

      expect(component.building()).toBeTruthy();
      expect(component.unitsLoading()).toBeFalse();
      expect(snackBar.open).toHaveBeenCalled();
    });
  });

  describe('editBuilding', () => {
    it('updates the building and reloads on confirmation, passing a minimal property context', () => {
      fixture.detectChanges();
      const req: UpdateBuildingRequest = { building_name: 'Renamed' };
      dialog.open.and.returnValue(dialogRefStub(req));
      buildingService.updateBuilding.and.returnValue(
        of(makeBuilding({ building_name: 'Renamed' }))
      );
      buildingService.getBuildingWithStats.calls.reset();

      component.editBuilding();

      expect(dialog.open).toHaveBeenCalledWith(
        jasmine.anything(),
        jasmine.objectContaining({
          data: jasmine.objectContaining({
            mode: 'edit',
            property: jasmine.objectContaining({ id: 42, property_name: 'Sunrise Apartments' }),
          }),
        })
      );
      expect(buildingService.updateBuilding).toHaveBeenCalledWith(7, req);
      expect(buildingService.getBuildingWithStats).toHaveBeenCalledWith(7);
      expect(snackBar.open).toHaveBeenCalled();
    });

    it('does nothing when the edit dialog is dismissed without a result', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(undefined));

      component.editBuilding();

      expect(buildingService.updateBuilding).not.toHaveBeenCalled();
    });

    it('shows an error toast when the update fails', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub({ building_name: 'Renamed' }));
      buildingService.updateBuilding.and.returnValue(throwError(() => ({ status: 500 })));

      component.editBuilding();

      expect(snackBar.open).toHaveBeenCalled();
    });
  });

  describe('deleteBuilding', () => {
    it('deletes the building and navigates back to the parent property on confirmation', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(true));
      buildingService.deleteBuilding.and.returnValue(of(undefined));

      component.deleteBuilding();

      expect(buildingService.deleteBuilding).toHaveBeenCalledWith(7);
      expect(router.navigate).toHaveBeenCalledWith(['/properties', 42]);
    });

    it('does not delete when the confirmation dialog is cancelled', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(false));

      component.deleteBuilding();

      expect(buildingService.deleteBuilding).not.toHaveBeenCalled();
    });

    it('shows an error toast and does not navigate when deletion fails', () => {
      fixture.detectChanges();
      dialog.open.and.returnValue(dialogRefStub(true));
      buildingService.deleteBuilding.and.returnValue(throwError(() => ({ status: 500 })));

      component.deleteBuilding();

      expect(snackBar.open).toHaveBeenCalled();
      expect(router.navigate).not.toHaveBeenCalled();
    });
  });

  describe('viewUnitDetails', () => {
    it('opens the entity detail dialog with the unit and lease/tenant info', () => {
      fixture.detectChanges();
      const unit = makeUnit({ tenant_name: 'Rahim Uddin', lease_active: true });

      component.viewUnitDetails(unit);

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

      component.viewUnitDetails(unit);

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
