import {
  Component,
  ElementRef,
  OnInit,
  computed,
  inject,
  signal,
  viewChildren,
} from '@angular/core';
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
  /** Unit numbers already present in the building — used to flag collisions
   *  before submitting, since the backend rejects the whole batch on the first
   *  one it finds. Optional: omitting it just disables that pre-flight check. */
  existingUnitNumbers?: string[];
}

interface BulkUnitRow {
  unit_number: string;
  unit_name: string;
}

/** Why a row is rejected. Blank rows are not an issue — they're skipped. */
type RowIssue = 'DUPLICATE' | 'EXISTS' | null;

const MIN_ROWS = 1;
const MAX_ROWS = 100;
const UNIT_NUMBER_MAX_LENGTH = 50;
const UNIT_NAME_MAX_LENGTH = 100;
/** Numbers shown as chips before the preview collapses to "… last". */
const PREVIEW_HEAD = 3;

const emptyRow = (): BulkUnitRow => ({ unit_number: '', unit_name: '' });

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

  private numberInputs = viewChildren<ElementRef<HTMLInputElement>>('numberInput');

  readonly maxRows = MAX_ROWS;
  readonly unitNumberMaxLength = UNIT_NUMBER_MAX_LENGTH;
  readonly unitNameMaxLength = UNIT_NAME_MAX_LENGTH;

  commonForm!: FormGroup;
  allowedUnitTypes: UnitType[] = [];

  // Generator inputs are plain fields rather than a FormGroup — they drive
  // nothing but a preview and one button, and the template re-reads them on
  // every change-detection pass anyway.
  pattern = 'A-{n}';
  startAt = '101';
  count = 10;

  rows = signal<BulkUnitRow[]>([emptyRow()]);
  /** Set when a generate/paste hit the MAX_ROWS ceiling, so the UI can say so. */
  capped = signal(false);

  private existingNumbers = new Set<string>();

  /** Case-sensitive, matching the DB's exact UNIQUE(building_id, unit_number). */
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

  /** Numbers that already exist in this building (pre-flight for the API check). */
  conflictingUnitNumbers = computed(() => {
    const conflicts = new Set<string>();
    for (const row of this.rows()) {
      const value = row.unit_number.trim();
      if (value && this.existingNumbers.has(value)) {
        conflicts.add(value);
      }
    }
    return conflicts;
  });

  filledCount = computed(
    () => this.rows().filter((row) => row.unit_number.trim().length > 0).length
  );
  blankCount = computed(() => this.rows().length - this.filledCount());
  hasIssues = computed(
    () => this.duplicateUnitNumbers().size > 0 || this.conflictingUnitNumbers().size > 0
  );

  ngOnInit(): void {
    this.allowedUnitTypes = getUnitTypesForBuilding(this.data.building.building_type);
    this.existingNumbers = new Set(
      (this.data.existingUnitNumbers ?? []).map((value) => value.trim()).filter(Boolean)
    );
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
    return this.commonForm.valid && this.filledCount() > 0 && !this.hasIssues();
  }

  // ── Generator ──────────────────────────────────────────────────────────────

  /** Rows still available before the batch ceiling. */
  remainingCapacity(): number {
    return Math.max(MAX_ROWS - this.rows().filter((row) => !this.isBlank(row)).length, 0);
  }

  canGenerate(): boolean {
    return this.buildNumbers().length > 0 && this.remainingCapacity() > 0;
  }

  /** Full sequence the current generator settings would produce. */
  private buildNumbers(): string[] {
    const start = Number.parseInt(this.startAt, 10);
    if (!Number.isFinite(start) || start < 0) return [];

    const total = Math.min(Math.max(Math.trunc(this.count) || 0, 0), MAX_ROWS);
    if (total < 1) return [];

    // Leading zeros in the start value set the width: "01" → 01, 02, … 10.
    const raw = this.startAt.trim();
    const padWidth = raw.startsWith('0') ? raw.length : 0;
    const pattern = this.pattern.trim() || '{n}';
    const template = /\{n\}/i.test(pattern) ? pattern : `${pattern}{n}`;

    return Array.from({ length: total }, (_, index) =>
      template
        .replace(/\{n\}/gi, String(start + index).padStart(padWidth, '0'))
        .slice(0, UNIT_NUMBER_MAX_LENGTH)
    );
  }

  /** Head of the sequence for the inline chip preview. */
  previewNumbers(): string[] {
    return this.buildNumbers().slice(0, PREVIEW_HEAD);
  }

  /** Last number, shown after an ellipsis when the sequence is longer than the head. */
  previewTail(): string | null {
    const numbers = this.buildNumbers();
    return numbers.length > PREVIEW_HEAD ? numbers[numbers.length - 1] : null;
  }

  /**
   * Appends the generated numbers, dropping any rows that are entirely blank
   * first — so the starter row disappears but typed-in data is never lost.
   */
  generate(): void {
    const numbers = this.buildNumbers();
    if (!numbers.length) return;

    const kept = this.rows().filter((row) => !this.isBlank(row));
    const added = numbers.slice(0, Math.max(MAX_ROWS - kept.length, 0));
    this.capped.set(added.length < numbers.length);
    this.rows.set([...kept, ...added.map((unit_number) => ({ unit_number, unit_name: '' }))]);
    this.ensureMinRows();
  }

  clearAll(): void {
    this.capped.set(false);
    this.rows.set([emptyRow()]);
  }

  // ── Rows ───────────────────────────────────────────────────────────────────

  addRow(): boolean {
    if (this.rows().length >= MAX_ROWS) {
      this.capped.set(true);
      return false;
    }
    this.rows.update((rows) => [...rows, emptyRow()]);
    return true;
  }

  addRowAndFocus(): void {
    const index = this.rows().length;
    if (this.addRow()) this.focusRow(index);
  }

  removeRow(index: number): void {
    if (this.rows().length <= MIN_ROWS) {
      this.clearAll();
      return;
    }
    this.capped.set(false);
    this.rows.update((rows) => rows.filter((_, i) => i !== index));
  }

  // Called from the template on every per-row input change — signals need a
  // new array reference to notify computed()s, since rows are mutated in
  // place via [(ngModel)].
  touch(): void {
    this.rows.update((rows) => [...rows]);
  }

  /**
   * Spreadsheet-style paste: a multi-value clipboard payload fills this row and
   * the ones below it, creating rows as needed. A single value pastes normally.
   */
  onPaste(event: ClipboardEvent, index: number): void {
    const text = event.clipboardData?.getData('text') ?? '';
    const tokens = text
      .split(/[\r\n\t,;]+/)
      .map((token) => token.trim())
      .filter((token) => token.length > 0);
    if (tokens.length < 2) return;

    event.preventDefault();
    const rows = [...this.rows()];
    for (const [offset, token] of tokens.entries()) {
      const target = index + offset;
      if (target >= MAX_ROWS) break;
      const unit_number = token.slice(0, UNIT_NUMBER_MAX_LENGTH);
      if (target < rows.length) {
        rows[target] = { ...rows[target], unit_number };
      } else {
        rows.push({ unit_number, unit_name: '' });
      }
    }
    this.capped.set(index + tokens.length > MAX_ROWS);
    this.rows.set(rows);
  }

  /** Enter moves down the unit-number column, adding a row at the bottom. */
  onEnter(index: number): void {
    if (index === this.rows().length - 1) {
      if (!this.addRow()) return;
    }
    this.focusRow(index + 1);
  }

  rowIssue(row: BulkUnitRow): RowIssue {
    const value = row.unit_number.trim();
    if (!value) return null;
    if (this.duplicateUnitNumbers().has(value)) return 'DUPLICATE';
    if (this.conflictingUnitNumbers().has(value)) return 'EXISTS';
    return null;
  }

  isBlank(row: BulkUnitRow): boolean {
    return !row.unit_number.trim() && !row.unit_name.trim();
  }

  getUnitTypeIcon(type: UnitType): string {
    return getUnitTypeIcon(type);
  }

  // ── Submit ─────────────────────────────────────────────────────────────────

  onSubmit(): void {
    if (!this.canSubmit()) {
      Object.keys(this.commonForm.controls).forEach((key) => {
        this.commonForm.get(key)?.markAsTouched();
      });
      return;
    }

    const common = this.commonForm.getRawValue();
    // Blank rows are silently skipped rather than blocking the batch.
    const units: BulkCreateUnitItem[] = this.rows()
      .filter((row) => row.unit_number.trim().length > 0)
      .map((row) => ({
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

  private ensureMinRows(): void {
    if (this.rows().length === 0) this.rows.set([emptyRow()]);
  }

  private focusRow(index: number): void {
    // One tick out, so the row added above exists in the view query.
    setTimeout(() => this.numberInputs()[index]?.nativeElement.focus());
  }
}
