import { Component, inject } from '@angular/core';
import { Location } from '@angular/common';
import { RouterModule } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { TranslateModule } from '@ngx-translate/core';
import { ErrorPage } from '../../../shared/error-page/error-page';

@Component({
  selector: 'app-unauthorized',
  standalone: true,
  imports: [ErrorPage, RouterModule, MatButtonModule, MatIconModule, TranslateModule],
  template: `
    <app-error-page
      code="403"
      icon="lock"
      tone="warn"
      [status]="'ERROR_PAGES.UNAUTHORIZED.STATUS' | translate"
      [title]="'ERROR_PAGES.UNAUTHORIZED.TITLE' | translate"
      [message]="'ERROR_PAGES.UNAUTHORIZED.MESSAGE' | translate"
    >
      <a mat-flat-button color="primary" routerLink="/dashboard">
        <mat-icon>dashboard</mat-icon>
        {{ 'ERROR_PAGES.ACTIONS.BACK_TO_DASHBOARD' | translate }}
      </a>
      <button mat-stroked-button type="button" (click)="goBack()">
        <mat-icon>arrow_back</mat-icon>
        {{ 'ERROR_PAGES.ACTIONS.GO_BACK' | translate }}
      </button>
    </app-error-page>
  `,
})
export class Unauthorized {
  private location = inject(Location);

  goBack(): void {
    this.location.back();
  }
}
