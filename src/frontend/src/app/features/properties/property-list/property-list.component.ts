import { Component, OnInit, inject, signal } from '@angular/core';

import { RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDialog } from '@angular/material/dialog';
import { Store } from '@ngrx/store';
import { PropertyActions } from '../store/property.actions';
import { BuildingActions } from '../store/building.actions';
import { UnitActions } from '../store/unit.actions';
import { selectAllProperties, selectPropertyLoading, selectPropertyError } from '../store/property.selectors';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import {
  Property,
  Building,
  Unit,
  PropertyListResponse,
  BuildingListResponse,
  UnitListResponse,
} from '../../../core/models';

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
    RouterModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatExpansionModule,
    MatTooltipModule,
  ],
  templateUrl: './property-list.component.html',
  styleUrls: ['./property-list.component.scss'],
})
export class PropertyListComponent implements OnInit {
  private store = inject(Store);
  private buildingService = inject(BuildingService);
  private unitService = inject(UnitService);
  private dialog = inject(MatDialog);

  // Store selectors
  properties = this.store.selectSignal(selectAllProperties);
  loading = this.store.selectSignal(selectPropertyLoading);
  error = this.store.selectSignal(selectPropertyError);

  // UI state
  expandedProperties = signal<Set<number>>(new Set());
  loadedBuildings = signal<Map<number, BuildingWithUnits[]>>(new Map());

  // Combined state for template
  displayedProperties = computed(() => {
    const props = this.properties();
    const expandedProps = this.expandedProperties();
    const buildingsMap = this.loadedBuildings();

    return props.map(p => ({
      ...p,
      expanded: expandedProps.has(p.id),
      buildings: buildingsMap.get(p.id)
    })) as DisplayedProperty[];
  });

  ngOnInit() {
    this.store.dispatch(PropertyActions.loadProperties({ active: true }));
  }

  toggleProperty(property: Property) {
    const expanded = this.expandedProperties();
    const newExpanded = new Set(expanded);

    this.propertyService.getProperties({ active: true }).subscribe({
      next: (response: PropertyListResponse) => {
        this.properties.set(response.properties.map((p: Property) => ({ ...p, expanded: false })));
        this.loading.set(false);
      },
      error: (err: any) => {
        this.error.set('Failed to load properties');
        this.loading.set(false);
        console.error('Error loading properties:', err);
      },
    });
  }

  toggleProperty(property: PropertyWithHierarchy) {
    property.expanded = !property.expanded;

    if (property.expanded && !property.buildings) {
      this.loadBuildings(property);
    }
    this.expandedProperties.set(newExpanded);
  }

  loadBuildings(propertyId: number) {
    this.buildingService.getBuildingsByProperty(propertyId).subscribe({
      next: (response: BuildingListResponse) => {
        const buildings = response.buildings.map((b: Building) => ({ ...b, expanded: false }));
        this.loadedBuildings.update(map => {
          const newMap = new Map(map);
          newMap.set(propertyId, buildings);
          return newMap;
        });
      },
      error: (err: any) => {
        console.error('Error loading buildings:', err);
      },
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
      next: (response: UnitListResponse) => {
        building.units = response.units;
      },
      error: (err: any) => {
        console.error('Error loading units:', err);
      },
    });
  }

  getPropertyTypeColor(type: string): string {
    switch (type) {
      case 'Residential':
        return 'primary';
      case 'Commercial':
        return 'accent';
      case 'Mixed':
        return 'warn';
      default:
        return '';
    }
  }

  getUnitTypeIcon(type: string): string {
    switch (type) {
      case 'Shop':
        return 'store';
      case 'Apartment':
        return 'home';
      case 'Office':
        return 'business';
      case 'Parking':
        return 'local_parking';
      case 'Storage':
        return 'inventory_2';
      default:
        return 'meeting_room';
    }
  }

  // Property CRUD handlers
  addProperty() {
    const dialogRef = this.dialog.open(PropertyFormDialogComponent, {
      width: '600px',
      data: { mode: 'create' }
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        this.store.dispatch(PropertyActions.createProperty({ property: result }));
      }
    });
  }

  editProperty(property: DisplayedProperty) {
    const dialogRef = this.dialog.open(PropertyFormDialogComponent, {
      width: '600px',
      data: { mode: 'edit', property }
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        this.store.dispatch(PropertyActions.updateProperty({
          id: property.id,
          property: result
        }));
      }
    });
  }

  deleteProperty(property: DisplayedProperty) {
    if (confirm(`Are you sure you want to delete ${property.property_name}?`)) {
      this.store.dispatch(PropertyActions.deleteProperty({ id: property.id }));
    }
  }

  // Building CRUD handlers
  addBuilding(property: DisplayedProperty) {
    const dialogRef = this.dialog.open(BuildingFormDialogComponent, {
      width: '600px',
      data: { mode: 'create', property }
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        this.store.dispatch(BuildingActions.createBuilding({ request: result }));
        // Reload buildings after creation
        setTimeout(() => {
          this.loadBuildings(property.id);
        }, 500);
      }
    });
  }

  editBuilding(building: BuildingWithUnits, property: DisplayedProperty) {
    const dialogRef = this.dialog.open(BuildingFormDialogComponent, {
      width: '600px',
      data: { mode: 'edit', building, property }
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        this.store.dispatch(BuildingActions.updateBuilding({
          id: building.id,
          request: result
        }));
      }
    });
  }

  deleteBuilding(building: BuildingWithUnits, propertyId: number) {
    if (confirm(`Are you sure you want to delete ${building.building_name}?`)) {
      this.store.dispatch(BuildingActions.deleteBuilding({ id: building.id }));
      // Reload buildings for this property after deletion
      setTimeout(() => this.loadBuildings(propertyId), 500);
    }
  }

  // Unit CRUD handlers
  addUnit(building: BuildingWithUnits, property: DisplayedProperty) {
    const dialogRef = this.dialog.open(UnitFormDialogComponent, {
      width: '600px',
      data: { mode: 'create', building, property }
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        this.store.dispatch(UnitActions.createUnit({ request: result }));
        // Reload units for this building after creation
        setTimeout(() => this.loadUnits(building), 500);
      }
    });
  }

  editUnit(unit: Unit, building: BuildingWithUnits, property: DisplayedProperty) {
    const dialogRef = this.dialog.open(UnitFormDialogComponent, {
      width: '600px',
      data: { mode: 'edit', unit, building, property }
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        this.store.dispatch(UnitActions.updateUnit({
          id: unit.id,
          request: result
        }));
      }
    });
  }

  deleteUnit(unit: Unit, building: BuildingWithUnits) {
    if (confirm(`Are you sure you want to delete unit ${unit.unit_number}?`)) {
      this.store.dispatch(UnitActions.deleteUnit({ id: unit.id }));
      // Reload units for this building after deletion
      setTimeout(() => this.loadUnits(building), 500);
    }
  }
}
