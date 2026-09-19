import { Injectable, inject } from '@angular/core';
import { DOCUMENT } from '@angular/common';
import { Title, Meta } from '@angular/platform-browser';

/**
 * Centralizes document title and meta-tag updates for SEO and link previews.
 *
 * The app is client-rendered, so these run at runtime on each navigation. The
 * static defaults in index.html cover crawlers that only read the initial HTML;
 * this service keeps title, description, canonical, and Open Graph/Twitter tags
 * in step as the user moves between routes.
 */
@Injectable({ providedIn: 'root' })
export class SeoService {
  private readonly title = inject(Title);
  private readonly meta = inject(Meta);
  private readonly doc = inject(DOCUMENT);

  private readonly brand = 'Tenantly';
  private readonly tagline = 'Property Rental Management';

  /** Sets the document title as "Page · Tenantly", falling back to the brand tagline. */
  setTitle(pageTitle?: string): void {
    const full =
      pageTitle && pageTitle !== this.brand
        ? `${pageTitle} · ${this.brand}`
        : `${this.brand} — ${this.tagline}`;

    this.title.setTitle(full);
    this.meta.updateTag({ property: 'og:title', content: full });
    this.meta.updateTag({ name: 'twitter:title', content: full });
  }

  setDescription(description: string): void {
    this.meta.updateTag({ name: 'description', content: description });
    this.meta.updateTag({ property: 'og:description', content: description });
    this.meta.updateTag({ name: 'twitter:description', content: description });
  }

  /** Points the canonical link and og:url at the current origin + path. */
  updateCanonicalUrl(): void {
    const url = this.doc.location.origin + this.doc.location.pathname;

    let link = this.doc.querySelector<HTMLLinkElement>("link[rel='canonical']");
    if (!link) {
      link = this.doc.createElement('link');
      link.setAttribute('rel', 'canonical');
      this.doc.head.appendChild(link);
    }
    link.setAttribute('href', url);

    this.meta.updateTag({ property: 'og:url', content: url });
  }
}
