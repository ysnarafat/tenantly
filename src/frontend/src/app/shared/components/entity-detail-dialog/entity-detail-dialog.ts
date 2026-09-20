import { Component, inject } from '@angular/core';
import { MatDialogModule, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';

export interface DetailRow {
  label: string;
  value: string;
}

export interface EntityDetailDialogData {
  title: string;
  subtitle?: string;
  icon?: string;
  rows: DetailRow[];
}

@Component({
  selector: 'app-entity-detail-dialog',
  standalone: true,
  imports: [MatDialogModule, MatIconModule, MatButtonModule],
  templateUrl: './entity-detail-dialog.html',
  styleUrl: './entity-detail-dialog.scss',
})
export class EntityDetailDialogComponent {
  data = inject<EntityDetailDialogData>(MAT_DIALOG_DATA);
}
