import { ApplicationConfig, importProvidersFrom, isDevMode, APP_INITIALIZER } from '@angular/core';
import { provideRouter, TitleStrategy, withInMemoryScrolling } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { MatSnackBarModule, MAT_SNACK_BAR_DEFAULT_OPTIONS } from '@angular/material/snack-bar';
import { MAT_FORM_FIELD_DEFAULT_OPTIONS } from '@angular/material/form-field';
import { provideNativeDateAdapter } from '@angular/material/core';
import { provideStore } from '@ngrx/store';
import { provideEffects } from '@ngrx/effects';
import { provideStoreDevtools } from '@ngrx/store-devtools';
import { provideTranslateService } from '@ngx-translate/core';
import { provideTranslateHttpLoader } from '@ngx-translate/http-loader';

import { routes } from './app.routes';
import { AppTitleStrategy } from './core/seo/app-title-strategy';
import { authInterceptor } from './core/interceptors/auth.interceptor';
import { slowRequestInterceptor } from './core/interceptors/slow-request.interceptor';
import { reducers, metaReducers } from './store';
import { AuthEffects } from './store/auth/auth.effects';
import { LanguageService } from './core/services/language.service';

export const appConfig: ApplicationConfig = {
  providers: [
    // Datepickers need a DateAdapter from the root environment injector —
    // importing MatNativeDateModule only at the component level doesn't
    // reliably reach it inside a MatDialog-opened component's calendar
    // overlay (NG0201: No provider found for DateAdapter), so it's provided
    // once here for every datepicker in the app.
    provideNativeDateAdapter(),
    provideRouter(
      routes,
      withInMemoryScrolling({ anchorScrolling: 'enabled', scrollPositionRestoration: 'enabled' })
    ),
    { provide: TitleStrategy, useClass: AppTitleStrategy },
    provideHttpClient(withInterceptors([authInterceptor, slowRequestInterceptor])),
    importProvidersFrom(MatSnackBarModule),
    {
      // Every snackBar.open(...) call in the app gets this baseline unless it
      // overrides panelClass/position itself — a top-right toast rather than
      // Material's default bottom-center bar, with app-toast supplying the
      // rounded/shadowed look in styles.scss.
      provide: MAT_SNACK_BAR_DEFAULT_OPTIONS,
      useValue: {
        duration: 3500,
        horizontalPosition: 'end',
        verticalPosition: 'top',
        panelClass: ['app-toast'],
      },
    },
    {
      // Many list/filter fields pair a `mat-label` with a native `[placeholder]`.
      // With the default 'auto' float behavior, the label only floats on focus/value,
      // so it visually overlaps the placeholder text while the field is empty and unfocused.
      // Always floating the label keeps the field name pinned above the input, never overlapping.
      provide: MAT_FORM_FIELD_DEFAULT_OPTIONS,
      useValue: { floatLabel: 'always' },
    },
    provideTranslateService({
      fallbackLang: 'en',
      loader: provideTranslateHttpLoader({
        prefix: 'assets/i18n/',
        suffix: `.json?v=${Date.now()}`,
      }),
    }),
    {
      provide: APP_INITIALIZER,
      useFactory: (languageService: LanguageService) => () => languageService.init(),
      deps: [LanguageService],
      multi: true,
    },
    provideStore(reducers, {
      metaReducers,
      runtimeChecks: {
        strictStateImmutability: true,
        strictActionImmutability: true,
        strictStateSerializability: true,
        strictActionSerializability: true,
        strictActionWithinNgZone: true,
        strictActionTypeUniqueness: true,
      },
    }),
    provideEffects([AuthEffects]),
    provideStoreDevtools({
      maxAge: 25,
      logOnly: !isDevMode(),
      autoPause: true,
      trace: false,
      traceLimit: 75,
    }),
  ],
};
