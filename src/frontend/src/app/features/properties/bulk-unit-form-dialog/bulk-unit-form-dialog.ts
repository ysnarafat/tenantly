import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  FormBuilder,
  FormGroup,
  Validators,
  ReactiveFormsModule,
  FormsModule,
} from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { TranslateModule } from '@ngx-translate/core';
import {
  Building,
  Property,
  UnitType,
  BulkCreateUnitItem,
  BulkCreateUnitsRequest,
} from '../../../core/models';
import { getUnitTypesForBuilding, getUnitTypeIcon } from '../unit-type.utils';

export interface BulkUnitFormDialogData {
  building: Building;
  property: Property;
}

interface BulkUnitRow {
  unit_number: string;
  unit_name: string;
}

const MIN_ROWS = 1;
const MAX_ROWS = 100;

@Component({
  selector: 'app-bulk-unit-form-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatSelectModule,
    MatIconModule,
    MatTooltipModule,
    TranslateModule,
  ],
  templateUrl: './bulk-unit-form-dialog.html',
  styleUrls: ['./bulk-unit-form-dialog.scss'],
})
export class BulkUnitFormDialogComponent implements OnInit {
  private fb = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<BulkUnitFormDialogComponent>);
  public data = inject<BulkUnitFormDialogData>(MAT_DIALOG_DATA);

  commonForm!: FormGroup;
  allowedUnitTypes: UnitType[] = [];

  rowCount = signal(1);
  rows = signal<BulkUnitRow[]>([{ unit_number: '', unit_name: '' }]);

  // Case-sensitive, matching the DB's exact UNIQUE(building_id, unit_number).
  duplicateUnitNumbers = computed(() => {
    const seen = new Set<string>();
    const duplicates = new Set<string>();
    for (const row of this.rows()) {
      const value = row.unit_number.trim();
      if (!value) continue;
      if (seen.has(value)) {
        duplicates.add(value);
      }
      seen.add(value);
    }
    return duplicates;
  });

  ngOnInit(): void {
    this.allowedUnitTypes = getUnitTypesForBuilding(this.data.building.building_type);
    this.commonForm = this.fb.group({
      unit_type: [this.allowedUnitTypes[0] ?? '', [Validators.required]],
      floor: [null, [Validators.min(0), Validators.max(200)]],
      section: ['', [Validators.maxLength(50)]],
    });
  }

  // A plain method (not computed()) — commonForm.valid isn't a signal, so a
  // computed() reading it untracked would go stale on form-only changes that
  // don't also touch the rows signal. Template calls re-evaluate this every
  // change-detection pass, which is what we want here.
  canSubmit(): boolean {
    const rows = this.rows();
    return (
      this.commonForm.valid &&
      rows.length > 0 &&
      rows.every((r) => r.unit_number.trim().length > 0) &&
      this.duplicateUnitNumbers().size === 0
    );
  }

  resizeRows(count: number): void {
    const clamped = Math.min(Math.max(Math.trunc(count) || MIN_ROWS, MIN_ROWS), MAX_ROWS);
    this.rowCount.set(clamped);
    const current = this.rows();
    if (clamped > current.length) {
      const additional = Array.from({ length: clamped - current.length }, () => ({
        unit_number: '',
        unit_name: '',
      }));
      this.rows.set([...current, ...additional]);
    } else if (clamped < current.length) {
      this.rows.set(current.slice(0, clamped));
    }
  }

  addRow(): void {
    if (this.rows().length >= MAX_ROWS) return;
    this.rows.update((rows) => [...rows, { unit_number: '', unit_name: '' }]);
    this.rowCount.set(this.rows().length);
  }

  removeRow(index: number): void {
    if (this.rows().length <= MIN_ROWS) return;
    this.rows.update((rows) => rows.filter((_, i) => i !== index));
    this.rowCount.set(this.rows().length);
  }

  // Called from the template on every per-row input change — signals need a
  // new array reference to notify computed()s, since rows are mutated in
  // place via [(ngModel)].
  touch(): void {
    this.rows.update((rows) => [...rows]);
  }

  isDuplicate(unitNumber: string): boolean {
    const value = unitNumber.trim();
    return value.length > 0 && this.duplicateUnitNumbers().has(value);
  }

  getUnitTypeIcon(type: UnitType): string {
    return getUnitTypeIcon(type);
  }

  onSubmit(): void {
    if (!this.canSubmit()) {
      Object.keys(this.commonForm.controls).forEach((key) => {
        this.commonForm.get(key)?.markAsTouched();
      });
      return;
    }

    const common = this.commonForm.getRawValue();
    const units: BulkCreateUnitItem[] = this.rows().map((row) => ({
      unit_number: row.unit_number.trim(),
      unit_name: row.unit_name.trim() || undefined,
      unit_type: common.unit_type,
      floor: common.floor ?? undefined,
      section: common.section || undefined,
    }));

    const request: BulkCreateUnitsRequest = { units };
    this.dialogRef.close(request);
  }

  onCancel(): void {
    this.dialogRef.close();
  }

  get buildingContext(): string {
    return `${this.data.property.property_name} › ${this.data.building.building_name}`;
  }
}
