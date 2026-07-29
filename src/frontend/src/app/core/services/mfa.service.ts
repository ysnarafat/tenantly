import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { MatDialog } from '@angular/material/dialog';
import { Observable, of, throwError } from 'rxjs';
import { switchMap, tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface MfaStatus {
  enabled: boolean;
}

export interface MfaEnrollResponse {
  secret: string;
  otpauth_uri: string;
}

export interface MfaVerifyResponse {
  step_up_token: string;
  expires_at: string; // ISO 8601
}

/**
 * Manages TOTP MFA: enrollment, verification, and the short-lived step-up token
 * required for sensitive actions (e.g. revealing a full NID). The step-up token
 * is held in memory only — never in localStorage/sessionStorage — and cleared on
 * expiry, matching the app's no-client-side-caching rule for sensitive material.
 */
@Injectable({ providedIn: 'root' })
export class MfaService {
  private http = inject(HttpClient);
  private dialog = inject(MatDialog);
  private apiUrl = `${environment.apiUrl}/auth/mfa`;

  private stepUpToken: string | null = null;
  private stepUpExpiresAt = 0;

  getStatus(): Observable<MfaStatus> {
    return this.http.get<MfaStatus>(`${this.apiUrl}/status`);
  }

  enroll(): Observable<MfaEnrollResponse> {
    return this.http.post<MfaEnrollResponse>(`${this.apiUrl}/enroll`, {});
  }

  verify(code: string): Observable<MfaVerifyResponse> {
    return this.http.post<MfaVerifyResponse>(`${this.apiUrl}/verify`, { code }).pipe(
      tap((res) => {
        this.stepUpToken = res.step_up_token;
        // Expire a little early to avoid using a token the server would reject.
        this.stepUpExpiresAt = new Date(res.expires_at).getTime() - 5000;
      })
    );
  }

  /** Returns a currently-valid step-up token, or null if none/expired. */
  getValidStepUpToken(): string | null {
    if (this.stepUpToken && Date.now() < this.stepUpExpiresAt) {
      return this.stepUpToken;
    }
    this.clearStepUp();
    return null;
  }

  clearStepUp(): void {
    this.stepUpToken = null;
    this.stepUpExpiresAt = 0;
  }

  /**
   * Ensures a valid step-up token exists, prompting the user to enroll and/or
   * enter a TOTP code via a dialog when needed. Emits the token, or errors if
   * the user cancels.
   */
  ensureStepUp(): Observable<string> {
    const existing = this.getValidStepUpToken();
    if (existing) {
      return of(existing);
    }
    return this.getStatus().pipe(
      switchMap((status) => this.openChallenge(status.enabled)),
      switchMap((token) =>
        token ? of(token) : throwError(() => new Error('MFA verification cancelled'))
      )
    );
  }

  private openChallenge(enrolled: boolean): Observable<string | null> {
    // Imported lazily to avoid a circular dependency between the service and the
    // dialog component (which itself uses this service).
    return new Observable<string | null>((subscriber) => {
      import('../../features/auth/mfa/mfa-dialog').then(({ MfaDialogComponent }) => {
        const ref = this.dialog.open(MfaDialogComponent, {
          width: '440px',
          disableClose: false,
          data: { enrolled },
        });
        ref.afterClosed().subscribe((token: string | null | undefined) => {
          subscriber.next(token ?? null);
          subscriber.complete();
        });
      });
    });
  }
}
