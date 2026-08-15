import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { PropertyService } from '../../../core/services/property.service';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import {
  PropertyWithStats,
  Building,
  UnitWithDetails,
  UpdatePropertyRequest,
  CreateBuildingRequest,
} from '../../../core/models';
import { LoadingSpinner } from '../../../shared/components/loading-spinner/loading-spinner';
import { PropertyFormDialogComponent } from '../property-form-dialog/property-form-dialog';
import { BuildingFormDialogComponent } from '../building-form-dialog/building-form-dialog';
import { ConfirmDeleteDialogComponent } from '../../../shared/components/confirm-delete-dialog/confirm-delete-dialog';
import {
  EntityDetailDialogComponent,
  DetailRow,
} from '../../../shared/components/entity-detail-dialog/entity-detail-dialog';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

@Component({
  selector: 'app-property-detail',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatTooltipModule,
    TranslateModule,
    LoadingSpinner,
  ],
  templateUrl: './property-detail.html',
  styleUrl: './property-detail.scss',
})
export class PropertyDetail implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private propertyService = inject(PropertyService);
  private buildingService = inject(BuildingService);
  private unitService = inject(UnitService);
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);

  propertyId = signal<number>(Number(this.route.snapshot.paramMap.get('id')));
  property = signal<PropertyWithStats | null>(null);
  buildings = signal<Building[]>([]);
  loading = signal(true);
  buildingsLoading = signal(false);
  notFound = signal(false);

  expandedBuildings = signal<Set<number>>(new Set());
  buildingUnits = signal<Map<number, UnitWithDetails[]>>(new Map());
  buildingUnitsLoading = signal<Set<number>>(new Set());

  occupancyRate = computed(() => {
    const p = this.property();
    if (!p || p.unit_count === 0) return 0;
    return Math.round((p.occupied_units / p.unit_count) * 100);
  });

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.notFound.set(false);
    this.propertyService.getPropertyWithStats(this.propertyId()).subscribe({
      next: (property) => {
        this.property.set(property);
        this.loading.set(false);
        this.loadBuildings();
      },
      error: (err) => {
        console.error('Error loading property:', safeErrorMessage(err));
        if (err?.status === 404) {
          this.notFound.set(true);
        } else {
          notifyError(this.snackBar, 'Failed to load property');
        }
        this.loading.set(false);
      },
    });
  }

  loadBuildings(): void {
    this.buildingsLoading.set(true);
    this.buildingService.getBuildingsByProperty(this.propertyId()).subscribe({
      next: (response) => {
        this.buildings.set(response.buildings ?? []);
        this.buildingsLoading.set(false);
      },
      error: (err) => {
        console.error('Error loading buildings:', safeErrorMessage(err));
        notifyError(this.snackBar, 'Failed to load buildings');
        this.buildingsLoading.set(false);
      },
    });
  }

  editProperty(): void {
    const property = this.property();
    if (!property) return;

    const ref = this.dialog.open(PropertyFormDialogComponent, {
      width: '600px',
      data: { mode: 'edit', property },
    });

    ref.afterClosed().subscribe((result: UpdatePropertyRequest | undefined) => {
      if (!result) return;

      this.propertyService.updateProperty(property.id, result).subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Property updated successfully');
          this.load();
        },
        error: (err) => {
          console.error('Error updating property:', safeErrorMessage(err));
          notifyError(this.snackBar, 'Failed to update property');
        },
      });
    });
  }

  deleteProperty(): void {
    const property = this.property();
    if (!property) return;

    const ref = this.dialog.open(ConfirmDeleteDialogComponent, {
      width: '480px',
      data: { entityLabel: 'property', entityName: property.property_name },
    });

    ref.afterClosed().subscribe((confirmed) => {
      if (!confirmed) return;

      this.propertyService.deleteProperty(property.id).subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Property deleted successfully');
          this.router.navigate(['/properties']);
        },
        error: (err) => {
          console.error('Error deleting property:', safeErrorMessage(err));
          notifyError(this.snackBar, 'Failed to delete property');
        },
      });
    });
  }

  addBuilding(): void {
    const property = this.property();
    if (!property) return;

    const ref = this.dialog.open(BuildingFormDialogComponent, {
      width: '600px',
      data: { mode: 'create', property },
    });

    ref.afterClosed().subscribe((result: CreateBuildingRequest | undefined) => {
      if (!result) return;

      this.buildingService.createBuilding(result).subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Building created successfully');
          this.load();
        },
        error: (err) => {
          console.error('Error creating building:', safeErrorMessage(err));
          notifyError(this.snackBar, 'Failed to create building');
        },
      });
    });
  }

  viewBuilding(building: Building): void {
    this.router.navigate(['/properties', this.propertyId(), 'buildings', building.id]);
  }

  isBuildingExpanded(building: Building): boolean {
    return this.expandedBuildings().has(building.id);
  }

  isBuildingUnitsLoading(building: Building): boolean {
    return this.buildingUnitsLoading().has(building.id);
  }

  unitsFor(building: Building): UnitWithDetails[] {
    return this.buildingUnits().get(building.id) ?? [];
  }

  toggleBuildingExpand(building: Building): void {
    const expanded = new Set(this.expandedBuildings());
    if (expanded.has(building.id)) {
      expanded.delete(building.id);
      this.expandedBuildings.set(expanded);
      return;
    }

    expanded.add(building.id);
    this.expandedBuildings.set(expanded);
    if (!this.buildingUnits().has(building.id)) {
      this.loadUnitsFor(building);
    }
  }

  private loadUnitsFor(building: Building): void {
    this.buildingUnitsLoading.update((set) => new Set(set).add(building.id));
    this.unitService.getUnitsByBuilding(building.id).subscribe({
      next: (response) => {
        const units = new Map(this.buildingUnits());
        units.set(building.id, response.units ?? []);
        this.buildingUnits.set(units);
        this.buildingUnitsLoading.update((set) => {
          const next = new Set(set);
          next.delete(building.id);
          return next;
        });
      },
      error: (err) => {
        console.error('Error loading units:', safeErrorMessage(err));
        notifyError(this.snackBar, 'Failed to load units');
        this.buildingUnitsLoading.update((set) => {
          const next = new Set(set);
          next.delete(building.id);
          return next;
        });
      },
    });
  }

  viewUnitDetails(unit: UnitWithDetails, event: Event): void {
    event.stopPropagation();
    const rows: DetailRow[] = [
      { label: 'Unit Number', value: unit.unit_number },
      { label: 'Unit Name', value: unit.unit_name ?? '—' },
      { label: 'Unit Type', value: unit.unit_type },
      { label: 'Floor', value: unit.floor?.toString() ?? '—' },
      { label: 'Section', value: unit.section ?? '—' },
      { label: 'Tenant', value: unit.tenant_name ?? 'Vacant' },
      { label: 'Lease Status', value: unit.lease_active ? 'Active lease' : 'No active lease' },
      { label: 'Status', value: unit.active ? 'Active' : 'Inactive' },
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

  formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-BD', {
      style: 'currency',
      currency: 'BDT',
      minimumFractionDigits: 0,
    }).format(amount);
  }
}
