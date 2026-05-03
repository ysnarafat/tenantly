import { Injectable, inject, signal } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

@Injectable({ providedIn: 'root' })
export class LanguageService {
  private translateService = inject(TranslateService);
  currentLang = signal<'en' | 'bn'>('en');

  init() {
    const saved = (localStorage.getItem('language') as 'en' | 'bn') || 'en';
    this.currentLang.set(saved);
    this.translateService.setDefaultLang('en');
    this.translateService.use(saved);
  }

  toggle() {
    const next = this.currentLang() === 'en' ? 'bn' : 'en';
    this.currentLang.set(next);
    localStorage.setItem('language', next);
    this.translateService.use(next);
  }
}
