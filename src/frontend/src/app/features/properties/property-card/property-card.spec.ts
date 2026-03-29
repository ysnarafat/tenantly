import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PropertyCardComponent } from './property-card';

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
    } as any;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
