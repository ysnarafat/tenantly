import { MatSnackBar } from '@angular/material/snack-bar';

/**
 * Runs a destructive/reversible action via an "undo" toast instead of a
 * blocking confirm dialog: the caller applies the change optimistically
 * (e.g. removing a row from a table) before calling this, then `commit`
 * only fires once the toast times out without the user clicking Undo. If
 * they do click Undo, `onUndo` fires instead so the caller can put the
 * optimistic change back — the server-side action never happened either way.
 */
export function actWithUndo(
  snackBar: MatSnackBar,
  message: string,
  commit: () => void,
  options?: { onUndo?: () => void; actionLabel?: string; duration?: number }
): void {
  const ref = snackBar.open(message, options?.actionLabel ?? 'Undo', {
    duration: options?.duration ?? 5000,
  });

  ref.onAction().subscribe(() => {
    options?.onUndo?.();
  });

  ref.afterDismissed().subscribe(({ dismissedByAction }) => {
    if (!dismissedByAction) {
      commit();
    }
  });
}
