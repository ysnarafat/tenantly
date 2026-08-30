import { Injectable, inject } from '@angular/core';
import { RouterStateSnapshot, TitleStrategy } from '@angular/router';
import { SeoService } from '../services/seo.service';

/**
 * Applies each route's `title` to the document via SeoService (which also
 * formats it as "Page · Tenantly" and refreshes canonical / Open Graph tags).
 *
 * Titles are plain strings resolved synchronously — deliberately independent of
 * the app's async translations so the tab title and link previews are always
 * correct on first paint, including on public pages a crawler may hit directly.
 */
@Injectable({ providedIn: 'root' })
export class AppTitleStrategy extends TitleStrategy {
  private readonly seo = inject(SeoService);

  override updateTitle(snapshot: RouterStateSnapshot): void {
    this.seo.updateCanonicalUrl();
    this.seo.setTitle(this.buildTitle(snapshot));
  }
}
