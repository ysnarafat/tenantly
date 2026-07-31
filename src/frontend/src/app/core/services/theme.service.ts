import { Injectable, inject, signal, effect } from '@angular/core';
import { FeatureFlagService } from './feature-flag.service';

type Theme = 'light' | 'dark';
const STORAGE_KEY = 'tenantly-theme';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private _featureFlags = inject(FeatureFlagService);
  readonly darkModeEnabled = this._featureFlags.isEnabled('darkMode');

  private _theme = signal<Theme>(this.darkModeEnabled ? this._loadTheme() : 'light');
  readonly isDark = this._theme.asReadonly();

  constructor() {
    effect(() => {
      const theme = this._theme();
      document.documentElement.setAttribute('data-theme', theme);
      if (this.darkModeEnabled) localStorage.setItem(STORAGE_KEY, theme);
    });
    document.documentElement.setAttribute('data-theme', this._theme());
  }

  toggle() {
    if (!this.darkModeEnabled) return;
    this._theme.update((t) => (t === 'light' ? 'dark' : 'light'));
  }

  private _loadTheme(): Theme {
    const stored = localStorage.getItem(STORAGE_KEY) as Theme | null;
    if (stored === 'dark' || stored === 'light') return stored;
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
}
