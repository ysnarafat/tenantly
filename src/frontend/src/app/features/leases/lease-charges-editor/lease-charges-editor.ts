import { Component, EventEmitter, Input, Output, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { LeaseService, ChargeType } from '../../../core/services/lease.service';
import { notifyError } from '../../../shared/utils/notify.utils';

// A charge row this editor works with — `id` is present once it's actually
// persisted (either it already existed, or it's been saved in "live" mode).
// A draft row created while leaseId is unset (the lease doesn't exist yet,
// e.g. inside the create dialog) has no id until the parent creates the
// lease and persists it.
export interface LeaseChargeDraft {
  id?: number;
  charge_type: ChargeType;
  label: string;
  amount: number;
  active: boolean;
}

const CHARGE_TYPES: ChargeType[] = ['Utility', 'ServiceCharge', 'Maintenance', 'Parking', 'Other'];

@Component({
  selector: 'app-lease-charges-editor',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIconModule,
    MatTooltipModule,
    TranslateModule,
  ],
  templateUrl: './lease-charges-editor.html',
  styleUrl: './lease-charges-editor.scss',
})
export class LeaseChargesEditor {
  private fb = inject(FormBuilder);
  private leaseService = inject(LeaseService);
  private snackBar = inject(MatSnackBar);

  // When set, every add/edit/remove/toggle persists immediately against this
  // lease's API endpoints ("live" mode — used by the edit dialog, where the
  // lease already exists). When unset, the editor only maintains an in-memory
  // draft list and relies on the parent to persist it once the lease itself
  // is created (used by the create dialog).
  @Input() leaseId?: number;
  @Input() charges: LeaseChargeDraft[] = [];
  @Output() chargesChange = new EventEmitter<LeaseChargeDraft[]>();

  readonly chargeTypes = CHARGE_TYPES;
  editingIndex: number | null = null;
  saving = false;

  form = this.fb.group({
    charge_type: ['Utility' as ChargeType, Validators.required],
    label: ['', [Validators.required, Validators.maxLength(100)]],
    amount: [null as number | null, [Validators.required, Validators.min(0)]],
  });

  get totalActive(): number {
    return this.charges.filter((c) => c.active).reduce((sum, c) => sum + c.amount, 0);
  }

  chargeTypeLabel(type: ChargeType): string {
    return type === 'ServiceCharge' ? 'Service Charge' : type;
  }

  startEdit(index: number): void {
    const c = this.charges[index];
    this.editingIndex = index;
    this.form.setValue({ charge_type: c.charge_type, label: c.label, amount: c.amount });
  }

  cancelEdit(): void {
    this.resetForm();
  }

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const v = this.form.value;
    const payload = {
      charge_type: v.charge_type as ChargeType,
      label: (v.label ?? '').trim(),
      amount: Number(v.amount),
    };

    if (this.editingIndex !== null) {
      this.updateAt(this.editingIndex, payload);
    } else {
      this.addNew(payload);
    }
  }

  toggleActive(index: number): void {
    const existing = this.charges[index];
    const active = !existing.active;

    if (!this.leaseId || existing.id === undefined) {
      this.emit(this.replaceAt(index, { ...existing, active }));
      return;
    }

    this.leaseService.updateLeaseCharge(this.leaseId, existing.id, { active }).subscribe({
      next: (updated) => this.emit(this.replaceAt(index, updated)),
      error: (err) => notifyError(this.snackBar, err?.error?.error || 'Failed to update charge'),
    });
  }

  remove(index: number): void {
    const existing = this.charges[index];

    if (!this.leaseId || existing.id === undefined) {
      this.emit(this.charges.filter((_, i) => i !== index));
      return;
    }

    this.leaseService.deleteLeaseCharge(this.leaseId, existing.id).subscribe({
      next: () => this.emit(this.charges.filter((_, i) => i !== index)),
      error: (err) => notifyError(this.snackBar, err?.error?.error || 'Failed to remove charge'),
    });
  }

  private addNew(payload: { charge_type: ChargeType; label: string; amount: number }): void {
    if (!this.leaseId) {
      this.emit([...this.charges, { ...payload, active: true }]);
      this.resetForm();
      return;
    }

    this.saving = true;
    this.leaseService.addLeaseCharge(this.leaseId, payload).subscribe({
      next: (created) => {
        this.saving = false;
        this.emit([...this.charges, created]);
        this.resetForm();
      },
      error: (err) => {
        this.saving = false;
        notifyError(this.snackBar, err?.error?.error || 'Failed to add charge');
      },
    });
  }

  private updateAt(
    index: number,
    payload: { charge_type: ChargeType; label: string; amount: number }
  ): void {
    const existing = this.charges[index];

    if (!this.leaseId || existing.id === undefined) {
      this.emit(this.replaceAt(index, { ...existing, ...payload }));
      this.resetForm();
      return;
    }

    this.saving = true;
    this.leaseService.updateLeaseCharge(this.leaseId, existing.id, payload).subscribe({
      next: (updated) => {
        this.saving = false;
        this.emit(this.replaceAt(index, updated));
        this.resetForm();
      },
      error: (err) => {
        this.saving = false;
        notifyError(this.snackBar, err?.error?.error || 'Failed to update charge');
      },
    });
  }

  private replaceAt(index: number, value: LeaseChargeDraft): LeaseChargeDraft[] {
    const updated = [...this.charges];
    updated[index] = value;
    return updated;
  }

  private emit(updated: LeaseChargeDraft[]): void {
    this.chargesChange.emit(updated);
  }

  private resetForm(): void {
    this.editingIndex = null;
    this.form.reset({ charge_type: 'Utility', label: '', amount: null });
  }
}
