import { MatSnackBar } from '@angular/material/snack-bar';

/**
 * Color-coded toast helpers on top of MatSnackBar. The neutral 'app-toast'
 * class (top-right, rounded, shadowed) already applies to every
 * snackBar.open(...) call via MAT_SNACK_BAR_DEFAULT_OPTIONS in app.config.ts
 * — these add a green/red/blue surface on top of that baseline for
 * success/error/info, so panelClass must repeat 'app-toast' here rather than
 * relying on the default merging with it (MatSnackBar's config merge
 * replaces the whole array, it doesn't concatenate).
 */
export function notifySuccess(snackBar: MatSnackBar, message: string, duration = 3000): void {
  snackBar.open(message, 'Close', { duration, panelClass: ['app-toast', 'toast-success'] });
}

export function notifyError(snackBar: MatSnackBar, message: string, duration = 5000): void {
  snackBar.open(message, 'Close', { duration, panelClass: ['app-toast', 'toast-error'] });
}

export function notifyInfo(snackBar: MatSnackBar, message: string, duration = 3000): void {
  snackBar.open(message, 'Close', { duration, panelClass: ['app-toast', 'toast-info'] });
}
