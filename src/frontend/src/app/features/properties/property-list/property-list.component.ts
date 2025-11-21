import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatTooltipModule } from '@angular/material/tooltip';
import { PropertyService } from '../../../core/services/property.service';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import { Property, Building, Unit, PropertyListResponse } from '../../../core/models';

interface PropertyWithHierarchy extends Property {
  buildings?: BuildingWithUnits[];
  expanded?: boolean;
}

interface BuildingWithUnits extends Building {
  units?: Unit[];
  expanded?: boolean;
}

@Component({
  selector: 'app-property-list',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatExpansionModule,
    MatTooltipModule
  ],
  templateUrl: './property-list.component.html',
  styleUrls: ['./property-list.component.scss']
})
export class PropertyListComponent implements OnInit {
  private propertyService = inject(PropertyService);
  private buildingService = inject(BuildingService);
  private unitService = inject(UnitService);

  properties = signal<PropertyWithHierarchy[]>([]);
  loading = signal(false);
  error = signal<string | null>(null);

  ngOnInit() {
    this.loadProperties();
  }

  loadProperties() {
    this.loading.set(true);
    this.error.set(null);

    this.propertyService.getProperties({ active: true }).subscribe({
      next: (response: PropertyListResponse) => {
        this.properties.set(response.properties.map((p: Property) => ({ ...p, expanded: false })));
        this.loading.set(false);
      },
      error: (err: any) => {
        this.error.set('Failed to load properties');
        this.loading.set(false);
        console.error('Error loading properties:', err);
      }
    });
  }

  toggleProperty(property: PropertyWithHierarchy) {
    property.expanded = !property.expanded;

    if (property.expanded && !property.buildings) {
      this.loadBuildings(property);
    }
  }

  loadBuildings(property: PropertyWithHierarchy) {
    this.buildingService.getBuildingsByProperty(property.id).subscribe({
      next: (buildings) => {
        property.buildings = buildings.map(b => ({ ...b, expanded: false }));
      },
      error: (err: any) => {
        console.error('Error loading buildings:', err);
      }
    });
  }

  toggleBuilding(building: BuildingWithUnits) {
    building.expanded = !building.expanded;

    if (building.expanded && !building.units) {
      this.loadUnits(building);
    }
  }

  loadUnits(building: BuildingWithUnits) {
    this.unitService.getUnitsByBuilding(building.id).subscribe({
      next: (units) => {
        building.units = units;
      },
      error: (err: any) => {
        console.error('Error loading units:', err);
      }
    });
  }

  getPropertyTypeColor(type: string): string {
    switch (type) {
      case 'Residential': return 'primary';
      case 'Commercial': return 'accent';
      case 'Mixed': return 'warn';
      default: return '';
    }
  }

  getUnitTypeIcon(type: string): string {
    switch (type) {
      case 'Shop': return 'store';
      case 'Apartment': return 'home';
      case 'Office': return 'business';
      case 'Parking': return 'local_parking';
      case 'Storage': return 'inventory_2';
      default: return 'meeting_room';
    }
  }

  // Add handlers
  addProperty() {
    console.log('Add Property clicked');
    // TODO: Open dialog to add property
    alert('Add Property form will be implemented here');
  }

  addBuilding(property: PropertyWithHierarchy) {
    console.log('Add Building to property:', property.property_name);
    // TODO: Open dialog to add building
    alert(`Add Building to ${property.property_name} will be implemented here`);
  }

  addUnit(building: BuildingWithUnits, property: PropertyWithHierarchy) {
    console.log('Add Unit to building:', building.building_name);
    // TODO: Open dialog to add unit
    alert(`Add Unit to ${building.building_name} will be implemented here`);
  }

  editProperty(property: PropertyWithHierarchy) {
    console.log('Edit property:', property.property_name);
    // TODO: Open dialog to edit property
    alert(`Edit ${property.property_name} will be implemented here`);
  }

  deleteProperty(property: PropertyWithHierarchy) {
    console.log('Delete property:', property.property_name);
    // TODO: Confirm and delete property
    if (confirm(`Are you sure you want to delete ${property.property_name}?`)) {
      alert('Delete functionality will be implemented here');
    }
  }
}
