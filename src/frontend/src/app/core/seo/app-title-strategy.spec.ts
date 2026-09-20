import { TestBed } from '@angular/core/testing';
import { RouterStateSnapshot } from '@angular/router';
import { AppTitleStrategy } from './app-title-strategy';
import { SeoService } from '../services/seo.service';

describe('AppTitleStrategy', () => {
  let strategy: AppTitleStrategy;
  let seo: jasmine.SpyObj<SeoService>;

  beforeEach(() => {
    seo = jasmine.createSpyObj<SeoService>('SeoService', ['updateCanonicalUrl', 'setTitle']);
    TestBed.configureTestingModule({
      providers: [AppTitleStrategy, { provide: SeoService, useValue: seo }],
    });
    strategy = TestBed.inject(AppTitleStrategy);
  });

  it('refreshes the canonical URL and applies the resolved route title', () => {
    spyOn(strategy, 'buildTitle').and.returnValue('Dashboard');

    strategy.updateTitle({} as RouterStateSnapshot);

    expect(seo.updateCanonicalUrl).toHaveBeenCalledTimes(1);
    expect(seo.setTitle).toHaveBeenCalledOnceWith('Dashboard');
  });

  it('passes undefined to SeoService (brand fallback) when a route has no title', () => {
    spyOn(strategy, 'buildTitle').and.returnValue(undefined);

    strategy.updateTitle({} as RouterStateSnapshot);

    expect(seo.updateCanonicalUrl).toHaveBeenCalledTimes(1);
    expect(seo.setTitle).toHaveBeenCalledOnceWith(undefined);
  });
});
