import { TestBed } from '@angular/core/testing';
import { Title, Meta } from '@angular/platform-browser';
import { SeoService } from './seo.service';

describe('SeoService', () => {
  let service: SeoService;
  let title: Title;
  let meta: Meta;

  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [SeoService, Title, Meta] });
    service = TestBed.inject(SeoService);
    title = TestBed.inject(Title);
    meta = TestBed.inject(Meta);
  });

  afterEach(() => {
    document.querySelector("link[rel='canonical']")?.remove();
    meta.removeTag("property='og:url'");
  });

  describe('setTitle', () => {
    it('formats a page title as "Page · Tenantly" and mirrors it to og/twitter', () => {
      service.setTitle('Dashboard');

      expect(title.getTitle()).toBe('Dashboard · Tenantly');
      expect(meta.getTag("property='og:title'")?.content).toBe('Dashboard · Tenantly');
      expect(meta.getTag("name='twitter:title'")?.content).toBe('Dashboard · Tenantly');
    });

    it('falls back to the brand tagline when no page title is given', () => {
      service.setTitle();
      expect(title.getTitle()).toBe('Tenantly — Property Rental Management');
    });

    it('falls back to the brand tagline when the page title is the brand itself', () => {
      service.setTitle('Tenantly');
      expect(title.getTitle()).toBe('Tenantly — Property Rental Management');
    });
  });

  describe('setDescription', () => {
    it('updates description across name, og, and twitter tags', () => {
      service.setDescription('Manage your rentals');

      expect(meta.getTag("name='description'")?.content).toBe('Manage your rentals');
      expect(meta.getTag("property='og:description'")?.content).toBe('Manage your rentals');
      expect(meta.getTag("name='twitter:description'")?.content).toBe('Manage your rentals');
    });
  });

  describe('updateCanonicalUrl', () => {
    it('creates a canonical link and og:url pointing at origin + path', () => {
      service.updateCanonicalUrl();

      const expected = document.location.origin + document.location.pathname;
      const link = document.querySelector<HTMLLinkElement>("link[rel='canonical']");
      expect(link).toBeTruthy();
      expect(link!.getAttribute('href')).toBe(expected);
      expect(meta.getTag("property='og:url'")?.content).toBe(expected);
    });

    it('reuses the existing canonical link instead of appending duplicates', () => {
      service.updateCanonicalUrl();
      service.updateCanonicalUrl();

      expect(document.querySelectorAll("link[rel='canonical']").length).toBe(1);
    });
  });
});
