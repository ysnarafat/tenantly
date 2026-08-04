import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ErrorPage } from './error-page';

describe('ErrorPage', () => {
  let fixture: ComponentFixture<ErrorPage>;
  let component: ErrorPage;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [ErrorPage] }).compileComponents();
    fixture = TestBed.createComponent(ErrorPage);
    component = fixture.componentInstance;
  });

  const el = (sel: string): HTMLElement | null => fixture.nativeElement.querySelector(sel);

  it('renders the code, status, title, message, and badge icon', () => {
    component.code = '404';
    component.status = 'Not found';
    component.icon = 'wrong_location';
    component.title = 'We can’t find that page';
    component.message = 'The address leads nowhere.';
    fixture.detectChanges();

    expect(el('.plate-code')?.textContent).toContain('404');
    expect(el('.plate-status')?.textContent).toContain('Not found');
    expect(el('.error-title')?.textContent).toContain('We can’t find that page');
    expect(el('.error-message')?.textContent).toContain('The address leads nowhere.');
    expect(el('.plate-badge mat-icon')?.textContent).toContain('wrong_location');
  });

  it('applies the warn tone class only when tone is "warn"', () => {
    component.code = '403';
    component.status = 'Forbidden';
    component.icon = 'lock';
    component.title = 'No access';
    component.message = 'Nope';
    component.tone = 'warn';
    fixture.detectChanges();
    expect(el('.error-page')?.classList).toContain('tone-warn');
  });

  it('defaults to the primary tone (no warn class)', () => {
    component.code = '404';
    component.status = 'Not found';
    component.icon = 'wrong_location';
    component.title = 'Missing';
    component.message = 'Gone';
    fixture.detectChanges();
    expect(el('.error-page')?.classList).not.toContain('tone-warn');
  });
});
