import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { TranslateModule } from '@ngx-translate/core';

export interface ConfirmDeleteDialogData {
  entityLabel: string;
  entityName: string;
}

@Component({
  selector: 'app-confirm-delete-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    TranslateModule,
  ],
  templateUrl: './confirm-delete-dialog.html',
  styleUrl: './confirm-delete-dialog.scss',
})
export class ConfirmDeleteDialogComponent {
  private dialogRef = inject(MatDialogRef<ConfirmDeleteDialogComponent>);
  public data = inject<ConfirmDeleteDialogData>(MAT_DIALOG_DATA);

  confirmationInput = new FormControl('', { nonNullable: true });

  get isMatch(): boolean {
    return this.confirmationInput.value.trim() === this.data.entityName;
  }

  onConfirm() {
    if (this.isMatch) {
      this.dialogRef.close(true);
    }
  }

  onCancel() {
    this.dialogRef.close(false);
  }
}
