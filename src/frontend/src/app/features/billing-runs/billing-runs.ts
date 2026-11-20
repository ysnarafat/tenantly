import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { Building } from '../../core/models/building.model';
import { BillingRun, BillingRunItem } from '../../core/models/billing-run.model';
import { Property } from '../../core/models/property.model';
import { BillingRunService } from '../../core/services/billing-run.service';
import { BuildingService } from '../../core/services/building.service';
import { PropertyService } from '../../core/services/property.service';

@Component({
  selector: 'app-billing-runs',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, TranslateModule],
  templateUrl: './billing-runs.html',
  styleUrl: './billing-runs.scss',
})
export class BillingRuns implements OnInit {
  private readonly billingService = inject(BillingRunService);
  private readonly buildingService = inject(BuildingService);
  private readonly propertyService = inject(PropertyService);
  private readonly translate = inject(TranslateService);

  properties = signal<Property[]>([]);
  buildings = signal<Building[]>([]);
  run = signal<BillingRun | null>(null);
  loading = signal(false);
  savingItemId = signal<number | null>(null);
  error = signal('');
  selectedPropertyId: number | null = null;
  selectedBuildingId: number | null = null;
  month = new Date().getMonth() + 1;
  year = new Date().getFullYear();

  ngOnInit(): void {
    this.loadProperties();
  }

  onPropertyChange(propertyId: number | null): void {
    this.selectedPropertyId = propertyId;
    this.selectedBuildingId = null;
    this.buildings.set([]);
    this.run.set(null);
    this.error.set('');
    if (propertyId) {
      this.loadBuildings(propertyId);
    }
  }

  onBuildingChange(buildingId: number | null): void {
    this.selectedBuildingId = buildingId;
    this.loadRun();
  }

  private loadProperties(page = 1, loaded: Property[] = []): void {
    this.propertyService.getProperties({ active: true, page, page_size: 100 }).subscribe({
      next: (response) => {
        const properties = [...loaded, ...(response.properties ?? [])];
        if (response.pagination?.has_next) {
          this.loadProperties(page + 1, properties);
        } else {
          this.properties.set(properties);
        }
      },
      error: () => this.error.set(this.translate.instant('BILLING_RUN.ERRORS.LOAD_PROPERTIES')),
    });
  }

  private loadBuildings(propertyId: number, page = 1, loaded: Building[] = []): void {
    this.buildingService.getBuildingsByProperty(propertyId, { page, page_size: 100 }).subscribe({
      next: (response) => {
        const buildings = [...loaded, ...(response.buildings ?? [])];
        if (response.pagination?.has_next) {
          this.loadBuildings(propertyId, page + 1, buildings);
        } else {
          this.buildings.set(buildings.filter((building) => building.active_status));
        }
      },
      error: () => this.error.set(this.translate.instant('BILLING_RUN.ERRORS.LOAD_BUILDINGS')),
    });
  }

  loadRun(): void {
    if (!this.selectedBuildingId || !this.validPeriod()) return;
    this.loading.set(true);
    this.error.set('');
    this.run.set(null);
    this.billingService.getByPeriod(this.selectedBuildingId, this.month, this.year).subscribe({
      next: (run) => {
        this.run.set(run);
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);
        if (err.status !== 404) {
          this.error.set(err.error?.error ?? this.translate.instant('BILLING_RUN.ERRORS.LOAD_RUN'));
        }
      },
    });
  }

  createRun(): void {
    if (!this.selectedBuildingId || !this.validPeriod()) return;
    this.loading.set(true);
    this.error.set('');
    this.billingService
      .create({ building_id: this.selectedBuildingId, month: this.month, year: this.year })
      .subscribe({
        next: (run) => {
          this.run.set(run);
          this.loading.set(false);
        },
        error: (err) => {
          this.loading.set(false);
          this.error.set(
            err.error?.error ?? this.translate.instant('BILLING_RUN.ERRORS.CREATE_RUN')
          );
        },
      });
  }

  saveElectricity(item: BillingRunItem): void {
    const run = this.run();
    const amount = item.electricity_amount;
    if (!run || run.status !== 'Draft' || amount === null) return;
    if (!Number.isFinite(amount) || amount < 0) {
      this.error.set(this.translate.instant('BILLING_RUN.ERRORS.INVALID_AMOUNT'));
      return;
    }
    this.error.set('');
    this.savingItemId.set(item.id);
    this.billingService.updateElectricity(run.id, item.id, item.meter_number, amount).subscribe({
      next: (updated) => {
        const current = this.run();
        this.run.set({
          ...updated,
          items: updated.items.map((savedItem) => {
            const localItem = current?.items.find((item) => item.id === savedItem.id);
            return savedItem.id === item.id || !localItem
              ? savedItem
              : {
                  ...savedItem,
                  meter_number: localItem.meter_number,
                  electricity_amount: localItem.electricity_amount,
                };
          }),
        });
        this.savingItemId.set(null);
      },
      error: (err) => {
        this.savingItemId.set(null);
        this.error.set(
          err.error?.error ?? this.translate.instant('BILLING_RUN.ERRORS.SAVE_ELECTRICITY')
        );
        this.loadRun();
      },
    });
  }

  finalize(): void {
    const run = this.run();
    if (!run || run.status !== 'Draft') return;
    if (this.hasMissingAmounts()) {
      this.error.set(this.translate.instant('BILLING_RUN.ERRORS.MISSING_AMOUNTS'));
      return;
    }
    if (!window.confirm(this.translate.instant('BILLING_RUN.CONFIRM_FINALIZE'))) {
      return;
    }
    this.loading.set(true);
    this.billingService.finalize(run.id).subscribe({
      next: (finalized) => {
        this.run.set(finalized);
        this.loading.set(false);
      },
      error: (err) => {
        this.error.set(
          err.error?.error ?? this.translate.instant('BILLING_RUN.ERRORS.FINALIZE_RUN')
        );
        this.loading.set(false);
      },
    });
  }

  printRegister(): void {
    if (this.hasMissingAmounts()) {
      this.error.set(this.translate.instant('BILLING_RUN.ERRORS.MISSING_AMOUNTS'));
      return;
    }
    this.error.set('');
    window.print();
  }

  itemsForColumn(column: number): BillingRunItem[] {
    const items = this.run()?.items ?? [];
    const midpoint = Math.ceil(items.length / 2);
    return column === 0 ? items.slice(0, midpoint) : items.slice(midpoint);
  }

  hasMissingAmounts(): boolean {
    const items = this.run()?.items ?? [];
    return items.length === 0 || items.some((item) => item.electricity_amount === null);
  }

  totalAmount(): number {
    return (this.run()?.items ?? []).reduce(
      (total, item) => total + (item.electricity_amount ?? 0),
      0
    );
  }

  monthName(month: number): string {
    const monthKeys = [
      'JAN',
      'FEB',
      'MAR',
      'APR',
      'MAY',
      'JUN',
      'JUL',
      'AUG',
      'SEP',
      'OCT',
      'NOV',
      'DEC',
    ];
    return this.translate.instant(`PAYMENT_COLLECT.MONTHS.${monthKeys[month - 1]}`);
  }

  preparedDate(): string {
    return new Intl.DateTimeFormat(this.translate.currentLang || 'en').format(new Date());
  }

  private validPeriod(): boolean {
    return this.month >= 1 && this.month <= 12 && this.year >= 2020;
  }
}
