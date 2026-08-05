import { Component, OnDestroy, effect, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Meta, Title } from '@angular/platform-browser';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { LanguageService } from '../../../core/services/language.service';
import { ThemeService } from '../../../core/services/theme.service';

interface Feature {
  key: string;
  icon: string;
}

const FEATURES: Feature[] = [
  { key: 'PAYMENTS', icon: 'payments' },
  { key: 'BILINGUAL', icon: 'translate' },
  { key: 'REMINDERS', icon: 'notifications_active' },
  { key: 'ROLES', icon: 'admin_panel_settings' },
  { key: 'REPORTS', icon: 'insights' },
  { key: 'PORTFOLIO', icon: 'apartment' },
];

const FAQ_KEYS = ['Q1', 'Q2', 'Q3', 'Q4', 'Q5', 'Q6', 'Q7', 'Q8'];

@Component({
  selector: 'app-homepage',
  standalone: true,
  imports: [RouterLink, MatIconModule, MatButtonModule, TranslateModule],
  templateUrl: './homepage.html',
  styleUrls: ['./homepage.scss'],
})
export class Homepage implements OnDestroy {
  public languageService = inject(LanguageService);
  public themeService = inject(ThemeService);
  private title = inject(Title);
  private meta = inject(Meta);
  private translate = inject(TranslateService);

  readonly contactEmail = 'hello@tenantly.com';
  readonly features = FEATURES;
  readonly faqKeys = FAQ_KEYS;

  openFaq = signal<string | null>(FAQ_KEYS[0]);

  private canonicalLink?: HTMLLinkElement;

  constructor() {
    // Keep the tab title/description in step with the active language —
    // the static <title>/<meta description> in index.html (read by
    // non-JS crawlers and link-preview bots) stay English-only.
    //
    // Uses translate.get() rather than .instant(): the translation JSON is
    // fetched over HTTP and APP_INITIALIZER doesn't wait on that load, so
    // right after bootstrap the catalog may not be loaded yet — .instant()
    // would return the raw key in that window and (since nothing else
    // re-triggers it) never correct itself. get() waits for the load if
    // needed, and (since it reads whatever language .use() last set) still
    // reflects the right language because currentLang() only changes after
    // LanguageService already called .use() for it.
    effect(() => {
      this.languageService.currentLang();
      this.translate
        .get(['HOMEPAGE.SEO.TITLE', 'HOMEPAGE.SEO.DESCRIPTION'])
        .subscribe((translations) => {
          this.title.setTitle(translations['HOMEPAGE.SEO.TITLE']);
          this.meta.updateTag({
            name: 'description',
            content: translations['HOMEPAGE.SEO.DESCRIPTION'],
          });
        });
    });
    this.setCanonicalLink();
  }

  ngOnDestroy(): void {
    this.canonicalLink?.remove();
  }

  toggleFaq(key: string): void {
    this.openFaq.update((current) => (current === key ? null : key));
  }

  entryNumber(index: number): string {
    return String(index + 1).padStart(2, '0');
  }

  private setCanonicalLink(): void {
    let link = document.querySelector<HTMLLinkElement>('link[rel="canonical"]');
    if (!link) {
      link = document.createElement('link');
      link.setAttribute('rel', 'canonical');
      document.head.appendChild(link);
    }
    link.setAttribute('href', `${window.location.origin}/home`);
    this.canonicalLink = link;
  }
}
