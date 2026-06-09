import { Injectable, inject } from '@angular/core';
import { MatSnackBar, MatSnackBarRef, TextOnlySnackBar } from '@angular/material/snack-bar';

@Injectable({ providedIn: 'root' })
export class SlowRequestService {
  private snackBar = inject(MatSnackBar);
  private pendingCount = 0;
  private snackBarRef: MatSnackBarRef<TextOnlySnackBar> | null = null;

  markSlow(): void {
    this.pendingCount++;
    if (this.pendingCount === 1) {
      this.snackBarRef = this.snackBar.open('⏳ This is taking longer than expected…', undefined, {
        panelClass: 'slow-request-snackbar',
      });
    }
  }

  markDone(): void {
    this.pendingCount = Math.max(0, this.pendingCount - 1);
    if (this.pendingCount === 0 && this.snackBarRef) {
      this.snackBarRef.dismiss();
      this.snackBarRef = null;
    }
  }
}
