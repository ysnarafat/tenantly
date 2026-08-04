import { Component, inject } from '@angular/core';
import { Location } from '@angular/common';
import { RouterModule } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { TranslateModule } from '@ngx-translate/core';
import { ErrorPage } from '../../../shared/error-page/error-page';

@Component({
  selector: 'app-not-found',
  standalone: true,
  imports: [ErrorPage, RouterModule, MatButtonModule, MatIconModule, TranslateModule],
  template: `
    <app-error-page
      code="404"
      icon="wrong_location"
      tone="primary"
      [status]="'ERROR_PAGES.NOT_FOUND.STATUS' | translate"
      [title]="'ERROR_PAGES.NOT_FOUND.TITLE' | translate"
      [message]="'ERROR_PAGES.NOT_FOUND.MESSAGE' | translate"
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
export class NotFound {
  private location = inject(Location);

  goBack(): void {
    this.location.back();
  }
}
