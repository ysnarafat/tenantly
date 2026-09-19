import { Component, effect, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Meta, Title } from '@angular/platform-browser';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { LanguageService } from '../../../core/services/language.service';
import { ThemeService } from '../../../core/services/theme.service';

@Component({
  selector: 'app-contact',
  standalone: true,
  imports: [RouterLink, MatIconModule, MatButtonModule, TranslateModule],
  templateUrl: './contact.html',
  styleUrls: ['./contact.scss'],
})
export class Contact {
  readonly languageService = inject(LanguageService);
  readonly themeService = inject(ThemeService);
  private readonly titleService = inject(Title);
  private readonly meta = inject(Meta);
  private readonly translate = inject(TranslateService);

  readonly contactEmail = 'hello@tenantly.com';
  readonly currentYear = new Date().getFullYear();
  emailCopied = signal(false);

  constructor() {
    effect(() => {
      this.languageService.currentLang();
      this.translate
        .get(['CONTACT_PAGE.SEO.TITLE', 'CONTACT_PAGE.SEO.DESCRIPTION'])
        .subscribe((t) => {
          this.titleService.setTitle(t['CONTACT_PAGE.SEO.TITLE']);
          this.meta.updateTag({ name: 'description', content: t['CONTACT_PAGE.SEO.DESCRIPTION'] });
        });
    });
  }

  copyEmail(): void {
    navigator.clipboard.writeText(this.contactEmail).then(() => {
      this.emailCopied.set(true);
      setTimeout(() => this.emailCopied.set(false), 2000);
    });
  }
}
