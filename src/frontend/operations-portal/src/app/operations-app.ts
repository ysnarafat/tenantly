import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { environment } from '../environments/environment';

interface LoginResponse {
  token: string;
  user: { role: string; email: string };
}

interface Plan {
  id: number;
  code: string;
  name_en: string;
  name_bn: string;
  description_en: string;
  description_bn: string;
  price_minor: number | null;
  currency: string;
  billing_period: 'monthly' | 'yearly';
  max_users: number;
  max_properties: number | null;
  features_en: string[];
  features_bn: string[];
  is_public: boolean;
  is_active: boolean;
  sort_order: number;
}

interface PlanDraft extends Plan {
  price_taka: number | null;
  features_en_text: string;
  features_bn_text: string;
  saving?: boolean;
}

interface CustomerOrganization {
  id: number;
  name: string;
  slug: string;
  subscription_tier: string;
  max_users: number;
  active: boolean;
  contact_name: string;
  contact_email: string;
  subscription_status: string;
  subscription_plan_code: string;
  subscription_renews_at: string | null;
  edit_plan_code: string;
  edit_status: string;
  edit_renews_at: string;
  saving?: boolean;
}

interface OrganizationInvitation {
  id: number;
  organization_id: number;
  organization_name: string;
  email: string;
  role: string;
  expires_at: string;
  email_delivery_status: 'unknown' | 'pending' | 'sent' | 'failed';
  last_sent_at: string | null;
}

interface OrganizationOnboardingResponse {
  organization: { id: number; name: string };
  invitation: OrganizationInvitation;
}

interface SupportCase {
  id: number;
  organization_id: number;
  organization: string;
  subject: string;
  description: string;
  status: string;
  priority: string;
  assigned_to: number | null;
  created_at: string;
  updated_at: string;
}

interface SupportAgent {
  id: number;
  name: string;
}

interface SupportMessage {
  id: number;
  case_id: number;
  author: string;
  body: string;
  created_at: string;
}

interface SupportNote {
  id: number;
  author: string;
  body: string;
  created_at: string;
}

type PortalTab = 'plans' | 'customers' | 'onboarding' | 'support';

@Component({
  selector: 'ops-root',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './operations-app.html',
  styleUrl: './operations-app.scss',
})
export class OperationsApp implements OnInit {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl;
  readonly tokenKey = 'tenantly_ops_token';

  username = '';
  password = '';
  loginError = '';
  error = '';
  notice = '';
  busy = false;
  activeTab: PortalTab = 'plans';
  plans: PlanDraft[] = [];
  organizations: CustomerOrganization[] = [];
  pendingInvitations: OrganizationInvitation[] = [];
  agents: SupportAgent[] = [];
  cases: SupportCase[] = [];
  selectedCase: SupportCase | null = null;
  caseMessages: SupportMessage[] = [];
  caseNotes: SupportNote[] = [];
  messageBody = '';
  noteBody = '';
  newCase = { organization_id: 0, subject: '', description: '', priority: 'normal' };
  newOrganization = { name: '', slug: '', plan_code: 'basic', admin_email: '' };
  onboarding = false;
  invitationActionID: number | null = null;

  get authenticated(): boolean {
    return !!sessionStorage.getItem(this.tokenKey);
  }

  ngOnInit(): void {
    if (this.authenticated) {
      this.loadPortal();
    }
  }

  login(): void {
    this.loginError = '';
    this.busy = true;
    this.http
      .post<LoginResponse>(`${this.apiUrl}/auth/login`, {
        username: this.username,
        password: this.password,
      })
      .subscribe({
        next: (response) => {
          this.busy = false;
          if (response.user.role !== 'SUPER_ADMIN') {
            this.loginError = 'This portal is restricted to platform administrators.';
            return;
          }
          sessionStorage.setItem(this.tokenKey, response.token);
          this.password = '';
          this.loadPortal();
        },
        error: (error: HttpErrorResponse) => {
          this.busy = false;
          this.loginError =
            error.error?.error ?? 'Sign-in failed. Check your credentials and retry.';
        },
      });
  }

  logout(): void {
    sessionStorage.removeItem(this.tokenKey);
    this.plans = [];
    this.organizations = [];
    this.pendingInvitations = [];
    this.agents = [];
    this.cases = [];
    this.selectedCase = null;
    this.caseMessages = [];
    this.error = '';
    this.notice = '';
  }

  selectTab(tab: PortalTab): void {
    this.activeTab = tab;
    this.error = '';
    this.notice = '';
    if (tab === 'plans' && this.plans.length === 0) this.loadPlans();
    if (tab === 'customers' && this.organizations.length === 0) this.loadOrganizations();
    if (tab === 'onboarding') {
      if (this.plans.length === 0) this.loadPlans();
      this.loadPendingInvitations();
    }
    if (tab === 'support' && this.cases.length === 0) this.loadCases();
  }

  savePlan(plan: PlanDraft): void {
    plan.saving = true;
    this.clearMessages();
    const body = {
      name_en: plan.name_en,
      name_bn: plan.name_bn,
      description_en: plan.description_en,
      description_bn: plan.description_bn,
      price_minor: plan.price_taka === null ? null : Math.round(plan.price_taka * 100),
      currency: plan.currency,
      billing_period: plan.billing_period,
      max_users: Number(plan.max_users),
      max_properties: plan.max_properties === null ? null : Number(plan.max_properties),
      features_en: plan.features_en_text.split(/\r?\n/),
      features_bn: plan.features_bn_text.split(/\r?\n/),
      is_public: plan.is_public,
      is_active: plan.is_active,
      sort_order: Number(plan.sort_order),
    };
    this.http.put<Plan>(`${this.apiUrl}/ops/plans/${plan.code}`, body).subscribe({
      next: () => {
        plan.saving = false;
        this.notice = `${plan.name_en} plan saved.`;
        this.loadPlans();
      },
      error: (error: HttpErrorResponse) => {
        plan.saving = false;
        this.showError(error);
      },
    });
  }

  saveSubscription(organization: CustomerOrganization): void {
    organization.saving = true;
    this.clearMessages();
    const renewsAt = organization.edit_renews_at
      ? new Date(`${organization.edit_renews_at}T00:00:00Z`).toISOString()
      : null;
    this.http
      .put(`${this.apiUrl}/ops/organizations/${organization.id}/subscription`, {
        plan_code: organization.edit_plan_code,
        status: organization.edit_status,
        renews_at: renewsAt,
      })
      .subscribe({
        next: () => {
          organization.saving = false;
          this.notice = `Subscription updated for ${organization.name}.`;
          this.loadOrganizations();
        },
        error: (error: HttpErrorResponse) => {
          organization.saving = false;
          this.showError(error);
        },
      });
  }

  onboardOrganization(): void {
    this.clearMessages();
    this.onboarding = true;
    this.http
      .post<OrganizationOnboardingResponse>(`${this.apiUrl}/ops/organizations/onboard`, {
        ...this.newOrganization,
      })
      .subscribe({
        next: ({ organization, invitation }) => {
          this.onboarding = false;
          this.newOrganization = { name: '', slug: '', plan_code: 'basic', admin_email: '' };
          this.notice =
            invitation.email_delivery_status === 'sent'
              ? `${organization.name} created; activation email sent to ${invitation.email}.`
              : `${organization.name} created; invite email failed. Check SMTP and retry below.`;
          this.loadOrganizations();
          this.loadPendingInvitations();
        },
        error: (error: HttpErrorResponse) => {
          this.onboarding = false;
          this.showError(error);
        },
      });
  }

  resendOrganizationInvitation(invitation: OrganizationInvitation): void {
    this.invitationActionID = invitation.id;
    this.clearMessages();
    this.http
      .post<OrganizationInvitation>(
        `${this.apiUrl}/ops/invitations/${invitation.id}/resend`,
        {}
      )
      .subscribe({
        next: (updatedInvitation) => {
          this.invitationActionID = null;
          this.pendingInvitations = this.pendingInvitations.map((item) =>
            item.id === updatedInvitation.id ? updatedInvitation : item
          );
          this.notice =
            updatedInvitation.email_delivery_status === 'sent'
              ? `Activation email sent to ${updatedInvitation.email}.`
              : `Email delivery failed for ${updatedInvitation.email}.`;
        },
        error: (error: HttpErrorResponse) => {
          this.invitationActionID = null;
          this.showError(error);
        },
      });
  }

  revokeOrganizationInvitation(invitation: OrganizationInvitation): void {
    this.invitationActionID = invitation.id;
    this.clearMessages();
    this.http.delete(`${this.apiUrl}/ops/invitations/${invitation.id}`).subscribe({
      next: () => {
        this.invitationActionID = null;
        this.pendingInvitations = this.pendingInvitations.filter((item) => item.id !== invitation.id);
        this.notice = `Invitation for ${invitation.email} revoked.`;
      },
      error: (error: HttpErrorResponse) => {
        this.invitationActionID = null;
        this.showError(error);
      },
    });
  }

  createCase(): void {
    this.clearMessages();
    if (this.newCase.organization_id < 1) {
      this.error = 'Select an organization before creating a support case.';
      return;
    }
    this.http.post(`${this.apiUrl}/ops/support/cases`, this.newCase).subscribe({
      next: () => {
        this.notice = 'Support case added.';
        this.newCase = { organization_id: 0, subject: '', description: '', priority: 'normal' };
        this.loadCases();
      },
      error: (error: HttpErrorResponse) => this.showError(error),
    });
  }

  openCase(supportCase: SupportCase): void {
    this.selectedCase = supportCase;
    this.caseNotes = [];
    this.caseMessages = [];
    this.http
      .get<{ notes: SupportNote[] }>(`${this.apiUrl}/ops/support/cases/${supportCase.id}/notes`)
      .subscribe({
        next: ({ notes }) => (this.caseNotes = notes),
        error: (error: HttpErrorResponse) => this.showError(error),
      });
    this.http
      .get<{
        messages: SupportMessage[];
      }>(`${this.apiUrl}/ops/support/cases/${supportCase.id}/messages`)
      .subscribe({
        next: ({ messages }) => (this.caseMessages = messages),
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  changeCaseStatus(supportCase: SupportCase, status: string): void {
    this.http
      .put<void>(`${this.apiUrl}/ops/support/cases/${supportCase.id}/status`, { status })
      .subscribe({
        next: () => {
          supportCase.status = status;
          this.notice = 'Support case status updated.';
        },
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  assignCase(supportCase: SupportCase, assignedTo: number | string | null): void {
    const assignedToID = assignedTo === '' || assignedTo === null ? null : Number(assignedTo);
    this.http
      .put<void>(`${this.apiUrl}/ops/support/cases/${supportCase.id}/assignment`, {
        assigned_to: assignedToID,
      })
      .subscribe({
        next: () => {
          supportCase.assigned_to = assignedToID;
          this.notice = 'Support case assignment updated.';
        },
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  addMessage(): void {
    if (!this.selectedCase || !this.messageBody.trim()) return;
    this.http
      .post<SupportMessage>(`${this.apiUrl}/ops/support/cases/${this.selectedCase.id}/messages`, {
        body: this.messageBody,
      })
      .subscribe({
        next: (message) => {
          this.caseMessages = [...this.caseMessages, message];
          this.messageBody = '';
        },
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  addNote(): void {
    if (!this.selectedCase || !this.noteBody.trim()) return;
    this.http
      .post<SupportNote>(`${this.apiUrl}/ops/support/cases/${this.selectedCase.id}/notes`, {
        body: this.noteBody,
      })
      .subscribe({
        next: (note) => {
          this.caseNotes = [...this.caseNotes, note];
          this.noteBody = '';
        },
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  formatDate(date: string): string {
    return new Date(date).toLocaleDateString();
  }

  private loadPortal(): void {
    this.loadPlans();
    this.loadOrganizations();
    this.loadCases();
    this.loadSupportAgents();
    this.loadPendingInvitations();
  }

  private loadPlans(): void {
    this.http.get<{ plans: Plan[] }>(`${this.apiUrl}/ops/plans`).subscribe({
      next: ({ plans }) => {
        this.plans = plans.map((plan) => ({
          ...plan,
          price_taka: plan.price_minor === null ? null : plan.price_minor / 100,
          features_en_text: plan.features_en.join('\n'),
          features_bn_text: plan.features_bn.join('\n'),
        }));
      },
      error: (error: HttpErrorResponse) => this.showError(error),
    });
  }

  private loadOrganizations(): void {
    this.http
      .get<{ organizations: CustomerOrganization[] }>(`${this.apiUrl}/ops/organizations`)
      .subscribe({
        next: ({ organizations }) => {
          this.organizations = organizations.map((organization) => ({
            ...organization,
            edit_plan_code: organization.subscription_plan_code || organization.subscription_tier,
            edit_status:
              organization.subscription_status === 'none'
                ? 'active'
                : organization.subscription_status,
            edit_renews_at: organization.subscription_renews_at
              ? organization.subscription_renews_at.slice(0, 10)
              : '',
          }));
        },
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  private loadPendingInvitations(): void {
    this.http.get<{ invitations: OrganizationInvitation[] }>(`${this.apiUrl}/ops/invitations`).subscribe({
      next: ({ invitations }) => (this.pendingInvitations = invitations),
      error: (error: HttpErrorResponse) => this.showError(error),
    });
  }

  private loadCases(): void {
    this.http.get<{ cases: SupportCase[] }>(`${this.apiUrl}/ops/support/cases`).subscribe({
      next: ({ cases }) => (this.cases = cases),
      error: (error: HttpErrorResponse) => this.showError(error),
    });
  }

  private loadSupportAgents(): void {
    this.http.get<{ agents: SupportAgent[] }>(`${this.apiUrl}/ops/support/agents`).subscribe({
      next: ({ agents }) => (this.agents = agents),
      error: (error: HttpErrorResponse) => this.showError(error),
    });
  }

  private showError(error: HttpErrorResponse): void {
    this.error = error.error?.message ?? error.error?.error ?? 'Request failed. Please retry.';
  }

  private clearMessages(): void {
    this.error = '';
    this.notice = '';
  }
}
