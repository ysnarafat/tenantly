import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { TranslateModule } from '@ngx-translate/core';
import { LeaseCustomFields } from '../../../core/services/lease.service';

// Arbitrary org-defined key/value pairs stored on the lease (e.g. "Parking
// Slot": "B-12") — not billed like a LeaseCharge, just extra reference data.
@Component({
  selector: 'app-lease-custom-fields-editor',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatTooltipModule,
    TranslateModule,
  ],
  templateUrl: './lease-custom-fields-editor.html',
  styleUrl: './lease-custom-fields-editor.scss',
})
export class LeaseCustomFieldsEditor {
  @Input() customFields: LeaseCustomFields = {};
  @Output() customFieldsChange = new EventEmitter<LeaseCustomFields>();

  newKey = '';
  newValue = '';

  get entries(): Array<[string, string]> {
    return Object.entries(this.customFields);
  }

  add(): void {
    const key = this.newKey.trim();
    const value = this.newValue.trim();
    if (!key || !value || key in this.customFields) return;

    this.customFieldsChange.emit({ ...this.customFields, [key]: value });
    this.newKey = '';
    this.newValue = '';
  }

  updateValue(key: string, value: string): void {
    this.customFieldsChange.emit({ ...this.customFields, [key]: value });
  }

  remove(key: string): void {
    const updated = { ...this.customFields };
    delete updated[key];
    this.customFieldsChange.emit(updated);
  }
}
