import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';

import { PropertyCardComponent, DisplayedProperty } from './property-card';

describe('PropertyCard', () => {
  let component: PropertyCardComponent;
  let fixture: ComponentFixture<PropertyCardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PropertyCardComponent],
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
});
