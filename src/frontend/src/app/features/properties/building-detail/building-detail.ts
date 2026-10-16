import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { BuildingService } from '../../../core/services/building.service';
import { UnitService } from '../../../core/services/unit.service';
import {
  BuildingWithStats,
  UnitWithDetails,
  Property,
  UpdateBuildingRequest,
  CreateUnitRequest,
  BulkCreateUnitsRequest,
} from '../../../core/models';
import { LoadingSpinner } from '../../../shared/components/loading-spinner/loading-spinner';
import { BuildingFormDialogComponent } from '../building-form-dialog/building-form-dialog';
import { UnitFormDialogComponent } from '../unit-form-dialog/unit-form-dialog';
import { BulkUnitFormDialogComponent } from '../bulk-unit-form-dialog/bulk-unit-form-dialog';
import { ConfirmDeleteDialogComponent } from '../../../shared/components/confirm-delete-dialog/confirm-delete-dialog';
import {
  EntityDetailDialogComponent,
  DetailRow,
} from '../../../shared/components/entity-detail-dialog/entity-detail-dialog';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

@Component({
  selector: 'app-building-detail',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatTooltipModule,
    MatFormFieldModule,
    MatInputModule,
    TranslateModule,
    LoadingSpinner,
  ],
  templateUrl: './building-detail.html',
  styleUrl: './building-detail.scss',
})
export class BuildingDetail implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private buildingService = inject(BuildingService);
  private unitService = inject(UnitService);
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);

  propertyId = signal<number>(Number(this.route.snapshot.paramMap.get('propertyId')));
  buildingId = signal<number>(Number(this.route.snapshot.paramMap.get('buildingId')));
  building = signal<BuildingWithStats | null>(null);
  units = signal<UnitWithDetails[]>([]);
  loading = signal(true);
  unitsLoading = signal(false);
  notFound = signal(false);
  searchQuery = signal('');

  filteredUnits = computed(() => {
    const query = this.searchQuery().trim().toLowerCase();
    const units = this.units();
    if (!query) return units;

    return units.filter(
      (u) =>
        u.unit_number?.toLowerCase().includes(query) ||
        u.unit_name?.toLowerCase().includes(query) ||
        u.unit_type?.toLowerCase().includes(query)
    );
  });

  ngOnInit(): void {
    this.load();
  }

  clearSearch(): void {
    this.searchQuery.set('');
  }

  load(): void {
    this.loading.set(true);
    this.notFound.set(false);
    this.buildingService.getBuildingWithStats(this.buildingId()).subscribe({
      next: (building) => {
        this.building.set(building);
        this.loading.set(false);
        this.loadUnits();
      },
      error: (err) => {
        console.error('Error loading building:', safeErrorMessage(err));
        if (err?.status === 404) {
          this.notFound.set(true);
        } else {
          notifyError(this.snackBar, 'Failed to load building');
        }
        this.loading.set(false);
      },
    });
  }

  loadUnits(): void {
    this.unitsLoading.set(true);
    this.unitService.getUnitsByBuilding(this.buildingId()).subscribe({
      next: (response) => {
        // Go serializes a nil slice as JSON null (not []) when there are zero
        // rows — guard against that rather than crash the template on .length.
        this.units.set(response.units ?? []);
        this.unitsLoading.set(false);
      },
      error: (err) => {
        console.error('Error loading units:', safeErrorMessage(err));
        notifyError(this.snackBar, 'Failed to load units');
        this.unitsLoading.set(false);
      },
    });
  }

  editBuilding(): void {
    const building = this.building();
    if (!building) return;

    const ref = this.dialog.open(BuildingFormDialogComponent, {
      width: '600px',
      data: {
        mode: 'edit',
        building,
        property: { id: building.property_id, property_name: building.property_name } as Property,
      },
    });

    ref.afterClosed().subscribe((result: UpdateBuildingRequest | undefined) => {
      if (!result) return;

      this.buildingService.updateBuilding(building.id, result).subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Building updated successfully');
          this.load();
        },
        error: (err) => {
          console.error('Error updating building:', safeErrorMessage(err));
          notifyError(this.snackBar, 'Failed to update building');
        },
      });
    });
  }

  deleteBuilding(): void {
    const building = this.building();
    if (!building) return;

    const ref = this.dialog.open(ConfirmDeleteDialogComponent, {
      width: '480px',
      data: { entityLabel: 'building', entityName: building.building_name },
    });

    ref.afterClosed().subscribe((confirmed) => {
      if (!confirmed) return;

      this.buildingService.deleteBuilding(building.id).subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Building deleted successfully');
          this.router.navigate(['/properties', this.propertyId()]);
        },
        error: (err) => {
          console.error('Error deleting building:', safeErrorMessage(err));
          notifyError(this.snackBar, 'Failed to delete building');
        },
      });
    });
  }

  addUnit(): void {
    const building = this.building();
    if (!building) return;

    const ref = this.dialog.open(UnitFormDialogComponent, {
      width: '600px',
      data: {
        mode: 'create',
        building,
        property: { id: building.property_id, property_name: building.property_name } as Property,
      },
    });

    ref.afterClosed().subscribe((result: CreateUnitRequest | undefined) => {
      if (!result) return;

      this.unitService.createUnit(result).subscribe({
        next: () => {
          notifySuccess(this.snackBar, 'Unit created successfully');
          this.loadUnits();
        },
        error: (err) => {
          console.error('Error creating unit:', safeErrorMessage(err));
          notifyError(this.snackBar, 'Failed to create unit');
        },
      });
    });
  }

  bulkAddUnits(): void {
    const building = this.building();
    if (!building) return;

    const ref = this.dialog.open(BulkUnitFormDialogComponent, {
      width: '960px',
      maxWidth: '95vw',
      maxHeight: '90vh',
      data: {
        building,
        property: { id: building.property_id, property_name: building.property_name } as Property,
        // Lets the dialog flag collisions up front — the API rejects the whole
        // batch on the first unit number that already exists.
        existingUnitNumbers: this.units().map((unit) => unit.unit_number),
      },
    });

    ref.afterClosed().subscribe((result: BulkCreateUnitsRequest | undefined) => {
      if (!result) return;

      this.unitService.bulkCreateUnits(building.id, result).subscribe({
        next: (response) => {
          notifySuccess(
            this.snackBar,
            `${response.summary.units_created} units created successfully`
          );
          this.loadUnits();
        },
        error: (err) => {
          console.error('Error bulk creating units:', safeErrorMessage(err));
          notifyError(this.snackBar, 'Failed to bulk create units');
        },
      });
    });
  }

  viewUnitDetails(unit: UnitWithDetails): void {
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
