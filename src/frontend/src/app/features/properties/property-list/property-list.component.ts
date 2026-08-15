import { Component, OnInit, computed, inject, signal } from '@angular/core';

import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { Store } from '@ngrx/store';
import { PropertyActions } from '../store/property.actions';
import { BuildingActions } from '../store/building.actions';
import { UnitActions } from '../store/unit.actions';
import {
  selectAllProperties,
  selectPropertyLoading,
  selectPropertyError,
} from '../store/property.selectors';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import {
  Property,
  PropertyType,
  Building,
  Unit,
  UnitWithDetails,
  BuildingListResponse,
} from '../../../core/models';
import { DisplayedProperty, PropertyCardComponent } from '../property-card/property-card';
import { PropertyFormDialogComponent } from '../property-form-dialog/property-form-dialog';
import { BuildingFormDialogComponent } from '../building-form-dialog/building-form-dialog';
import { UnitFormDialogComponent } from '../unit-form-dialog/unit-form-dialog';
import { PaymentCreateDialog } from '../../payments/payment-list/payment-list';
import { PaymentService } from '../../../core/services/payment.service';
import { LoadingSpinner } from '../../../shared/components/loading-spinner/loading-spinner';
import { ConfirmDeleteDialogComponent } from '../../../shared/components/confirm-delete-dialog/confirm-delete-dialog';
import {
  EntityDetailDialogComponent,
  DetailRow,
} from '../../../shared/components/entity-detail-dialog/entity-detail-dialog';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

interface PropertyWithHierarchy extends Property {
  buildings?: BuildingWithUnits[];
  expanded?: boolean;
}

interface BuildingWithUnits extends Building {
  units?: UnitWithDetails[];
  expanded?: boolean;
}

@Component({
  selector: 'app-property-list',
  standalone: true,
  imports: [
    RouterModule,
    FormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatTooltipModule,
    PropertyCardComponent,
    TranslateModule,
    LoadingSpinner,
  ],
  templateUrl: './property-list.component.html',
  styleUrls: ['./property-list.component.scss'],
})
export class PropertyListComponent implements OnInit {
  private store = inject(Store);
  private buildingService = inject(BuildingService);
  private unitService = inject(UnitService);
  private dialog = inject(MatDialog);
  private paymentService = inject(PaymentService);
  private snackBar = inject(MatSnackBar);

  // Store selectors
  properties = this.store.selectSignal(selectAllProperties);
  loading = this.store.selectSignal(selectPropertyLoading);
  error = this.store.selectSignal(selectPropertyError);

  // UI state
  expandedProperties = signal<Set<number>>(new Set());
  loadedBuildings = signal<Map<number, BuildingWithUnits[]>>(new Map());
  searchQuery = signal('');
  propertyTypeFilter = signal<PropertyType | null>(null);
  statusFilter = signal<'active' | 'inactive' | null>(null);
  cityFilter = signal<string | null>(null);

  propertyTypes: PropertyType[] = ['Residential', 'Commercial', 'Mixed'];

  activeCount = computed(() => this.properties().filter((p) => p.active).length);

  availableCities = computed(() => {
    const cities = new Set<string>();
    for (const p of this.properties()) {
      if (p.city) cities.add(p.city);
    }
    return [...cities].sort();
  });

  hasActiveFilters = computed(
    () =>
      !!(
        this.searchQuery() ||
        this.propertyTypeFilter() ||
        this.statusFilter() ||
        this.cityFilter()
      )
  );

  // Combined state for template
  displayedProperties = computed(() => {
    const props = this.properties();
    const expandedProps = this.expandedProperties();
    const buildingsMap = this.loadedBuildings();
    const query = this.searchQuery().trim().toLowerCase();
    const type = this.propertyTypeFilter();
    const status = this.statusFilter();
    const city = this.cityFilter();

    let withHierarchy = props.map((p) => ({
      ...p,
      expanded: expandedProps.has(p.id),
      buildings: buildingsMap.get(p.id),
    })) as DisplayedProperty[];

    if (type) {
      withHierarchy = withHierarchy.filter((p) => p.property_type === type);
    }
    if (status) {
      withHierarchy = withHierarchy.filter((p) => (status === 'active' ? p.active : !p.active));
    }
    if (city) {
      withHierarchy = withHierarchy.filter((p) => p.city === city);
    }
    if (!query) return withHierarchy;

    return withHierarchy.filter((property) => this.matchesSearch(property, query));
  });

  private matchesSearch(property: DisplayedProperty, query: string): boolean {
    const propertyMatch =
      property.property_name?.toLowerCase().includes(query) ||
      property.property_code?.toLowerCase().includes(query) ||
      property.address?.toLowerCase().includes(query);
    if (propertyMatch) return true;

    const buildings = (property as PropertyWithHierarchy).buildings ?? [];
    return buildings.some((building) => {
      const buildingMatch =
        building.building_name?.toLowerCase().includes(query) ||
        building.building_code?.toLowerCase().includes(query);
      if (buildingMatch) return true;

      const units = building.units ?? [];
      return units.some(
        (unit) =>
          unit.unit_number?.toLowerCase().includes(query) ||
          unit.unit_name?.toLowerCase().includes(query)
      );
    });
  }

  clearSearch() {
    this.searchQuery.set('');
  }

  onTypeFilterChange(type: PropertyType | null) {
    this.propertyTypeFilter.set(type);
  }

  onStatusFilterChange(status: 'active' | 'inactive' | null) {
    this.statusFilter.set(status);
  }

  onCityFilterChange(city: string | null) {
    this.cityFilter.set(city);
  }

  resetFilters() {
    this.searchQuery.set('');
    this.propertyTypeFilter.set(null);
    this.statusFilter.set(null);
    this.cityFilter.set(null);
  }

  ngOnInit() {
    this.store.dispatch(PropertyActions.loadProperties({}));
  }

  toggleProperty(property: PropertyWithHierarchy) {
    this.expandedProperties.update((expanded) => {
      const newExpanded = new Set(expanded);
      if (newExpanded.has(property.id)) {
        newExpanded.delete(property.id);
      } else {
        newExpanded.add(property.id);
      }
      return newExpanded;
    });

    if (!property.buildings) {
      this.loadBuildings(property.id);
    }
  }

  loadBuildings(propertyId: number) {
    this.buildingService.getBuildingsByProperty(propertyId).subscribe({
      next: (response: BuildingListResponse) => {
        const buildings = (response.buildings ?? []).map((b: Building) => ({
          ...b,
          expanded: false,
        }));
        this.loadedBuildings.update((map) => {
          const newMap = new Map(map);
          newMap.set(propertyId, buildings);
          return newMap;
        });
      },
      error: (err: unknown) => {
        console.error('Error loading buildings:', safeErrorMessage(err));
        this.loadedBuildings.update((map) => {
          const newMap = new Map(map);
          newMap.set(propertyId, []);
          return newMap;
        });
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
      next: (response) => {
        building.units = response.units;
      },
      error: (err: unknown) => {
        console.error('Error loading units:', safeErrorMessage(err));
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
      data: { mode: 'create' },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.store.dispatch(PropertyActions.createProperty({ property: result }));
      }
    });
  }

  editProperty(property: DisplayedProperty) {
    const dialogRef = this.dialog.open(PropertyFormDialogComponent, {
      width: '600px',
      data: { mode: 'edit', property },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.store.dispatch(
          PropertyActions.updateProperty({
            id: property.id,
            property: result,
          })
        );
      }
    });
  }

  deleteProperty(property: DisplayedProperty) {
    const dialogRef = this.dialog.open(ConfirmDeleteDialogComponent, {
      width: '480px',
      data: { entityLabel: 'property', entityName: property.property_name },
    });

    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.store.dispatch(PropertyActions.deleteProperty({ id: property.id }));
      }
    });
  }

  // Building CRUD handlers
  addBuilding(property: DisplayedProperty) {
    const dialogRef = this.dialog.open(BuildingFormDialogComponent, {
      width: '600px',
      data: { mode: 'create', property },
    });

    dialogRef.afterClosed().subscribe((result) => {
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
      data: { mode: 'edit', building, property },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.store.dispatch(
          BuildingActions.updateBuilding({
            id: building.id,
            request: result,
          })
        );
      }
    });
  }

  deleteBuilding(building: BuildingWithUnits, propertyId: number) {
    const dialogRef = this.dialog.open(ConfirmDeleteDialogComponent, {
      width: '480px',
      data: { entityLabel: 'building', entityName: building.building_name },
    });

    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.store.dispatch(BuildingActions.deleteBuilding({ id: building.id }));
        // Reload buildings for this property after deletion
        setTimeout(() => this.loadBuildings(propertyId), 500);
      }
    });
  }

  // Unit CRUD handlers
  addUnit(building: BuildingWithUnits, property: DisplayedProperty) {
    const dialogRef = this.dialog.open(UnitFormDialogComponent, {
      width: '600px',
      data: { mode: 'create', building, property },
    });

    dialogRef.afterClosed().subscribe((result) => {
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
      data: { mode: 'edit', unit, building, property },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.store.dispatch(
          UnitActions.updateUnit({
            id: unit.id,
            request: result,
          })
        );
      }
    });
  }

  deleteUnit(unit: Unit, building: BuildingWithUnits) {
    const dialogRef = this.dialog.open(ConfirmDeleteDialogComponent, {
      width: '480px',
      data: { entityLabel: 'unit', entityName: unit.unit_number },
    });

    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.store.dispatch(UnitActions.deleteUnit({ id: unit.id }));
        // Reload units for this building after deletion
        setTimeout(() => this.loadUnits(building), 500);
      }
    });
  }

  viewUnitDetails(unit: UnitWithDetails, building: BuildingWithUnits) {
    const rows: DetailRow[] = [
      { label: 'Unit Number', value: unit.unit_number },
      { label: 'Unit Name', value: unit.unit_name ?? '—' },
      { label: 'Unit Type', value: unit.unit_type },
      { label: 'Floor', value: unit.floor?.toString() ?? '—' },
      { label: 'Section', value: unit.section ?? '—' },
      { label: 'Property', value: unit.property_name },
      { label: 'Building', value: building.building_name },
      { label: 'Tenant', value: unit.tenant_name ?? 'Vacant' },
      { label: 'Lease Status', value: unit.lease_active ? 'Active lease' : 'No active lease' },
      { label: 'Status', value: unit.active ? 'Active' : 'Inactive' },
      { label: 'Created', value: this.formatDate(unit.created_at) },
      { label: 'Last Updated', value: this.formatDate(unit.updated_at) },
    ];

    this.dialog.open(EntityDetailDialogComponent, {
      width: '420px',
      data: {
        title: unit.unit_number,
        subtitle: unit.unit_name || undefined,
        icon: 'meeting_room',
        rows,
      },
    });
  }

  private formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }

  addPaymentForUnit(
    unit: UnitWithDetails,
    building: BuildingWithUnits,
    property: DisplayedProperty
  ) {
    this.paymentService.searchLeases(unit.unit_number).subscribe({
      next: (res) => {
        const lease = res.results.find((l) => l.unit_id === unit.id) ?? null;
        const ref = this.dialog.open(PaymentCreateDialog, {
          width: '560px',
          maxWidth: '95vw',
          data: {
            unit_id: unit.id,
            unit_number: unit.unit_number,
            building_name: building.building_name,
            property_name: property.property_name,
            lease,
          },
        });
        ref.afterClosed().subscribe((req) => {
          if (req) {
            this.paymentService.createPayment(req).subscribe({
              next: () => {
                notifySuccess(this.snackBar, 'Payment recorded successfully');
              },
              error: (err: unknown) => {
                console.error('Failed to create payment', safeErrorMessage(err));
                notifyError(this.snackBar, 'Failed to record payment');
              },
            });
          }
        });
      },
      error: (err: unknown) => console.error('Failed to load lease', safeErrorMessage(err)),
    });
  }
}
