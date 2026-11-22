import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { OperationsApp } from './operations-app';
import { environment } from '../environments/environment';

describe('OperationsApp support conversation', () => {
  let fixture: ComponentFixture<OperationsApp>;
  let component: OperationsApp;
  let http: HttpTestingController;

  const supportCase = {
    id: 8,
    organization_id: 3,
    organization: 'Example Organization',
    subject: 'Login issue',
    description: 'The login page does not load.',
    status: 'open',
    priority: 'normal',
    assigned_to: null as number | null,
    created_at: '2026-11-08T10:00:00Z',
    updated_at: '2026-11-08T10:00:00Z',
  };

  beforeEach(async () => {
    sessionStorage.removeItem('tenantly_ops_token');
    await TestBed.configureTestingModule({
      imports: [OperationsApp, HttpClientTestingModule],
    }).compileComponents();
    fixture = TestBed.createComponent(OperationsApp);
    component = fixture.componentInstance;
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    sessionStorage.removeItem('tenantly_ops_token');
  });

  function signInForTest(): void {
    sessionStorage.setItem('tenantly_ops_token', 'staff-token');
    fixture.detectChanges();
    http.expectOne(`${environment.apiUrl}/ops/plans`).flush({ plans: [] });
    http.expectOne(`${environment.apiUrl}/ops/organizations`).flush({ organizations: [] });
    http.expectOne(`${environment.apiUrl}/ops/support/cases`).flush({ cases: [supportCase] });
    http.expectOne(`${environment.apiUrl}/ops/support/agents`).flush({ agents: [] });
    http.expectOne(`${environment.apiUrl}/ops/invitations`).flush({ invitations: [] });
    fixture.detectChanges();
  }

  it('signs in using the operations-only authentication endpoint', () => {
    component.username = 'platform-admin';
    component.password = 'secret';
    component.login();

    const request = http.expectOne(`${environment.apiUrl}/ops/auth/login`);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ username: 'platform-admin', password: 'secret' });
    request.flush({
      token: 'operations-token',
      user: { role: 'SUPER_ADMIN', email: 'ops@example.com' },
    });

    expect(sessionStorage.getItem('tenantly_ops_token')).toBe('operations-token');
    http.expectOne(`${environment.apiUrl}/ops/plans`).flush({ plans: [] });
    http.expectOne(`${environment.apiUrl}/ops/organizations`).flush({ organizations: [] });
    http.expectOne(`${environment.apiUrl}/ops/support/cases`).flush({ cases: [] });
    http.expectOne(`${environment.apiUrl}/ops/support/agents`).flush({ agents: [] });
    http.expectOne(`${environment.apiUrl}/ops/invitations`).flush({ invitations: [] });
  });

  it('renders customer-visible messages separately from private staff notes', () => {
    signInForTest();
    component.selectTab('support');
    component.openCase(supportCase);
    http.expectOne(`${environment.apiUrl}/ops/support/cases/8/notes`).flush({
      notes: [
        {
          id: 1,
          author: 'Staff',
          body: 'Internal escalation',
          created_at: '2026-11-08T10:02:00Z',
        },
      ],
    });
    http.expectOne(`${environment.apiUrl}/ops/support/cases/8/messages`).flush({
      messages: [
        {
          id: 2,
          case_id: 8,
          author: 'Org Admin',
          body: 'Still blocked',
          created_at: '2026-11-08T10:03:00Z',
        },
      ],
    });
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Customer conversation');
    expect(text).toContain('Still blocked');
    expect(text).toContain('Internal notes');
    expect(text).toContain('Internal escalation');
    expect(component.caseMessages.length).toBe(1);
    expect(component.caseNotes.length).toBe(1);
  });

  it('assigns a case using the staff-only assignment endpoint', () => {
    signInForTest();
    component.assignCase(supportCase, 5);
    const request = http.expectOne(`${environment.apiUrl}/ops/support/cases/8/assignment`);

    expect(request.request.method).toBe('PUT');
    expect(request.request.body).toEqual({ assigned_to: 5 });
    request.flush(null);
    expect(supportCase.assigned_to).toBe(5);
  });
});
