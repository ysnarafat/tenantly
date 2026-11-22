import { Component, input, computed } from '@angular/core';

export type StatChipType = 
  | 'total' | 'active' | 'inactive' 
  | 'expiring' | 'expired' 
  | 'pending' | 'saved' | 'failed' | 'outstanding'
  | 'success' | 'warning' | 'error' | 'info';

@Component({
  selector: 'app-stat-chip',
  standalone: true,
  templateUrl: './stat-chip.html',
  styleUrl: './stat-chip.scss',
})
export class StatChip {
  // Inputs
  readonly type = input<StatChipType>('total');
  readonly amount = input<number | string>('');
  readonly labelKey = input<string | null>(null); // For translation keys
  readonly label = input<string | null>(null); // For direct labels
  readonly showLabel = input(true);

  // Derived values
  readonly chipClass = computed(() => this.type());
  readonly fullClass = computed(() => {
    let cls = 'stat-chip ' + this.chipClass();
    if (this.showLabel()) {
      cls += ' chip-with-label';
    }
    return cls;
  });
  readonly resolvedLabel = computed(() => {
    if (this.label() !== null) return this.label();
    if (this.labelKey() !== null) return this.labelKey();
    return this.type();
  });
  readonly resolvedAmount = computed(() => {
    const amount = this.amount();
    return amount !== null && amount !== undefined ? amount : '0';
  });
}