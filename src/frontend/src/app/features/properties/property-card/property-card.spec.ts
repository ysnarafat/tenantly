import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { provideRouter, Router, RouterLink } from '@angular/router';

import { PropertyCardComponent, DisplayedProperty, BuildingWithUnits } from './property-card';

describe('PropertyCard', () => {
  let component: PropertyCardComponent;
  let fixture: ComponentFixture<PropertyCardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PropertyCardComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(PropertyCardComponent);
    component = fixture.componentInstance;
    component.property = {
      id: 1,
      property_name: 'Test Property',
      address: '123 Test St',
      property_type: 'Residential',
      expanded: false,
    } as DisplayedProperty;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('Delete action consistency (matching Tenant/Lease)', () => {
    it('renders the delete button with color="warn"', () => {
      const deleteButton = fixture.debugElement.query(
        By.css('mat-card-actions button[color="warn"]')
      );

      expect(deleteButton).withContext('delete button should exist').not.toBeNull();
    });

    it('renders the delete button with the shared "delete" icon', () => {
      const deleteButton = fixture.debugElement.query(
        By.css('mat-card-actions button[color="warn"]')
      );
      const icon = deleteButton.query(By.css('mat-icon'));

      expect(icon.nativeElement.textContent.trim()).toBe('delete');
    });
  });

  describe('Navigation to detail pages', () => {
    // RouterLink's `routerLink` input is a write-only setter (no getter), so
    // reading it back off the directive instance always yields undefined —
    // asserting via an actual navigation call is the reliable way to test it.
    function expectNavigatesTo(button: Element, path: string) {
      const router = TestBed.inject(Router);
      const navigateSpy = spyOn(router, 'navigateByUrl').and.resolveTo(true);

      (button as HTMLElement).click();

      expect(navigateSpy).toHaveBeenCalled();
      expect(navigateSpy.calls.mostRecent().args[0].toString()).toBe(path);
    }

    it('navigates "View Details" to the property detail route', () => {
      const links = fixture.debugElement.queryAll(By.directive(RouterLink));
      const viewDetailsLink = links.find((l) =>
        l.nativeElement.textContent.includes('View Details')
      );

      expect(viewDetailsLink).withContext('View Details link should exist').toBeTruthy();
      expectNavigatesTo(viewDetailsLink!.nativeElement, '/properties/1');
    });

    it('navigates the building info button to the building detail route once expanded', () => {
      component.property = {
        ...component.property,
        expanded: true,
        buildings: [
          {
            id: 7,
            building_name: 'Block A',
            building_code: 'BLK-A',
            expanded: false,
          } as BuildingWithUnits,
        ],
      };
      fixture.detectChanges();

      const infoButton = fixture.debugElement.query(By.css('.building-info-btn'));
      expect(infoButton).withContext('building info button should exist').not.toBeNull();
      expectNavigatesTo(infoButton.nativeElement, '/properties/1/buildings/7');
    });

    it('stops the building info button click from also toggling the building expand state', () => {
      component.property = {
        ...component.property,
        expanded: true,
        buildings: [
          {
            id: 7,
            building_name: 'Block A',
            building_code: 'BLK-A',
            expanded: false,
          } as BuildingWithUnits,
        ],
      };
      fixture.detectChanges();

      const toggleSpy = spyOn(component, 'onToggleBuilding');
      const infoButton = fixture.debugElement.query(By.css('.building-info-btn'));
      infoButton.nativeElement.click();

      expect(toggleSpy).not.toHaveBeenCalled();
    });
  });
});
