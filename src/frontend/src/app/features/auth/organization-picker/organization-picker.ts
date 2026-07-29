import { Component, OnInit, inject, signal, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { Subject } from 'rxjs';
import { takeUntil, filter } from 'rxjs/operators';
import { AppState } from '../../../store';
import * as AuthSelectors from '../../../store/auth/auth.selectors';
import * as AuthActions from '../../../store/auth/auth.actions';
import { UserOrganization } from '../../../core/models/organization.model';

@Component({
  selector: 'app-organization-picker',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="picker-page">
      <!-- Ambient background elements -->
      <div class="bg-grid"></div>
      <div class="bg-glow bg-glow--1"></div>
      <div class="bg-glow bg-glow--2"></div>
      <div class="bg-glow bg-glow--3"></div>

      <div class="picker-wrapper">
        <!-- Header -->
        <header class="picker-header">
          <div class="brand-mark">
            <span class="brand-icon">T</span>
          </div>
          <h1 class="picker-title">Select Workspace</h1>
          <p class="picker-subtitle">
            You have access to multiple organizations.<br />
            Choose one to continue.
          </p>
        </header>

        <!-- Organization Cards Grid -->
        <div class="org-grid" [class.loading]="isSwitching()">
          @for (org of organizations(); track org.id; let i = $index) {
            <button
              class="org-card"
              [class.org-card--switching]="switchingOrgId() === org.id"
              [class.org-card--dimmed]="isSwitching() && switchingOrgId() !== org.id"
              [disabled]="isSwitching()"
              (click)="selectOrganization(org)"
              [style.animation-delay]="i * 80 + 'ms'"
            >
              <!-- Card shimmer effect (shown when switching) -->
              @if (switchingOrgId() === org.id) {
                <div class="card-shimmer"></div>
              }

              <!-- Organization Avatar -->
              <div class="org-avatar" [style.background]="getOrgGradient(org.organization.name, i)">
                <span class="org-initial">{{ getOrgInitial(org.organization.name) }}</span>
              </div>

              <!-- Org Info -->
              <div class="org-info">
                <span class="org-name">{{ org.organization.name }}</span>
                @if (org.organization.description) {
                  <span class="org-description">{{ org.organization.description }}</span>
                }
                <span class="org-role-badge" [class]="'role-' + org.role.toLowerCase()">
                  {{ org.role }}
                </span>
              </div>

              <!-- Arrow -->
              <div class="card-arrow">
                @if (switchingOrgId() === org.id) {
                  <div class="spinner"></div>
                } @else {
                  <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                    <path
                      d="M4 10H16M16 10L11 5M16 10L11 15"
                      stroke="currentColor"
                      stroke-width="1.5"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                }
              </div>
            </button>
          }

          @if (organizations().length === 0) {
            <div class="empty-state">
              <div class="empty-icon">
                <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
                  <circle
                    cx="24"
                    cy="24"
                    r="20"
                    stroke="currentColor"
                    stroke-width="1.5"
                    opacity="0.3"
                  />
                  <path
                    d="M16 24h16M24 16v16"
                    stroke="currentColor"
                    stroke-width="1.5"
                    stroke-linecap="round"
                  />
                </svg>
              </div>
              <p class="empty-text">No organizations found</p>
              <p class="empty-hint">Contact your administrator for access</p>
            </div>
          }
        </div>

        <!-- Footer -->
        <footer class="picker-footer">
          <button class="logout-link" (click)="logout()">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path
                d="M6 14H3a1 1 0 01-1-1V3a1 1 0 011-1h3M11 11l3-3-3-3M14 8H6"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
            Sign out
          </button>
        </footer>
      </div>
    </div>
  `,
  styles: [
    `
      @import url('https://fonts.googleapis.com/css2?family=Cormorant+Garamond:wght@400;500;600&family=DM+Sans:wght@300;400;500;600&display=swap');

      :host {
        display: block;
        width: 100%;
        height: 100%;
      }

      /* ─── Page Shell ─── */
      .picker-page {
        min-height: 100vh;
        background-color: #0a0f1e;
        display: flex;
        align-items: center;
        justify-content: center;
        font-family: 'DM Sans', sans-serif;
        position: relative;
        overflow: hidden;
        padding: 40px 20px;
      }

      /* ─── Background Elements ─── */
      .bg-grid {
        position: absolute;
        inset: 0;
        background-image:
          linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
          linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
        background-size: 48px 48px;
        mask-image: radial-gradient(ellipse 80% 70% at 50% 50%, black 40%, transparent 100%);
      }

      .bg-glow {
        position: absolute;
        border-radius: 50%;
        filter: blur(80px);
        pointer-events: none;
      }

      .bg-glow--1 {
        width: 500px;
        height: 500px;
        background: radial-gradient(circle, rgba(99, 102, 241, 0.12) 0%, transparent 70%);
        top: -100px;
        left: -100px;
      }

      .bg-glow--2 {
        width: 400px;
        height: 400px;
        background: radial-gradient(circle, rgba(245, 158, 11, 0.08) 0%, transparent 70%);
        bottom: -80px;
        right: -80px;
      }

      .bg-glow--3 {
        width: 300px;
        height: 300px;
        background: radial-gradient(circle, rgba(16, 185, 129, 0.06) 0%, transparent 70%);
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
      }

      /* ─── Content Wrapper ─── */
      .picker-wrapper {
        position: relative;
        z-index: 1;
        width: 100%;
        max-width: 640px;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 40px;
      }

      /* ─── Header ─── */
      .picker-header {
        text-align: center;
        animation: fadeSlideDown 0.6s cubic-bezier(0.16, 1, 0.3, 1) both;
      }

      .brand-mark {
        width: 52px;
        height: 52px;
        background: linear-gradient(135deg, #6366f1 0%, #a78bfa 100%);
        border-radius: 14px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin: 0 auto 24px;
        box-shadow: 0 8px 32px rgba(99, 102, 241, 0.4);
      }

      .brand-icon {
        font-family: 'Cormorant Garamond', serif;
        font-size: 28px;
        font-weight: 600;
        color: #ffffff;
        line-height: 1;
      }

      .picker-title {
        font-family: 'Cormorant Garamond', serif;
        font-size: 2.25rem;
        font-weight: 500;
        color: #f1f5f9;
        margin: 0 0 10px;
        letter-spacing: -0.5px;
        line-height: 1.1;
      }

      .picker-subtitle {
        font-size: 0.9375rem;
        color: #64748b;
        margin: 0;
        line-height: 1.6;
        font-weight: 300;
      }

      /* ─── Org Grid ─── */
      .org-grid {
        width: 100%;
        display: flex;
        flex-direction: column;
        gap: 12px;
        transition: opacity 0.3s ease;
      }

      .org-grid.loading {
        opacity: 0.7;
      }

      /* ─── Org Card ─── */
      .org-card {
        position: relative;
        width: 100%;
        background: rgba(255, 255, 255, 0.035);
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 16px;
        padding: 20px 24px;
        display: flex;
        align-items: center;
        gap: 18px;
        cursor: pointer;
        transition:
          background 0.25s ease,
          border-color 0.25s ease,
          transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1),
          box-shadow 0.25s ease;
        overflow: hidden;
        text-align: left;
        animation: fadeSlideUp 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
        backdrop-filter: blur(12px);
        -webkit-backdrop-filter: blur(12px);
      }

      .org-card:hover:not(:disabled) {
        background: rgba(255, 255, 255, 0.065);
        border-color: rgba(255, 255, 255, 0.16);
        transform: translateY(-2px);
        box-shadow:
          0 8px 32px rgba(0, 0, 0, 0.3),
          0 0 0 1px rgba(255, 255, 255, 0.06) inset;
      }

      .org-card:hover:not(:disabled) .card-arrow {
        color: #a78bfa;
        transform: translateX(3px);
      }

      .org-card--switching {
        border-color: rgba(99, 102, 241, 0.4) !important;
        background: rgba(99, 102, 241, 0.08) !important;
      }

      .org-card--dimmed {
        opacity: 0.45;
      }

      .org-card:disabled {
        cursor: not-allowed;
      }

      /* ─── Card Shimmer ─── */
      .card-shimmer {
        position: absolute;
        inset: 0;
        background: linear-gradient(
          105deg,
          transparent 40%,
          rgba(255, 255, 255, 0.04) 50%,
          transparent 60%
        );
        background-size: 200% 100%;
        animation: shimmer 1.5s infinite;
        border-radius: inherit;
      }

      /* ─── Avatar ─── */
      .org-avatar {
        width: 48px;
        height: 48px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
      }

      .org-initial {
        font-family: 'Cormorant Garamond', serif;
        font-size: 22px;
        font-weight: 600;
        color: #ffffff;
        line-height: 1;
        text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
      }

      /* ─── Org Info ─── */
      .org-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
      }

      .org-name {
        font-size: 1rem;
        font-weight: 500;
        color: #e2e8f0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        letter-spacing: -0.1px;
      }

      .org-description {
        font-size: 0.8125rem;
        color: #475569;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        font-weight: 300;
      }

      .org-role-badge {
        display: inline-flex;
        align-items: center;
        margin-top: 2px;
        padding: 2px 10px;
        border-radius: 999px;
        font-size: 0.6875rem;
        font-weight: 500;
        letter-spacing: 0.3px;
        width: fit-content;
      }

      .role-admin {
        background: rgba(245, 158, 11, 0.12);
        color: #fbbf24;
        border: 1px solid rgba(245, 158, 11, 0.2);
      }

      .role-propertymanager {
        background: rgba(99, 102, 241, 0.12);
        color: #a5b4fc;
        border: 1px solid rgba(99, 102, 241, 0.2);
      }

      .role-accountant {
        background: rgba(16, 185, 129, 0.12);
        color: #6ee7b7;
        border: 1px solid rgba(16, 185, 129, 0.2);
      }

      /* ─── Card Arrow ─── */
      .card-arrow {
        color: #334155;
        transition:
          color 0.25s ease,
          transform 0.25s ease;
        flex-shrink: 0;
        display: flex;
        align-items: center;
      }

      /* ─── Spinner ─── */
      .spinner {
        width: 20px;
        height: 20px;
        border: 2px solid rgba(99, 102, 241, 0.2);
        border-top-color: #6366f1;
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
      }

      /* ─── Empty State ─── */
      .empty-state {
        text-align: center;
        padding: 60px 20px;
        color: #334155;
      }

      .empty-icon {
        margin-bottom: 16px;
      }

      .empty-text {
        font-size: 1rem;
        font-weight: 500;
        color: #475569;
        margin: 0 0 6px;
      }

      .empty-hint {
        font-size: 0.875rem;
        color: #334155;
        margin: 0;
      }

      /* ─── Footer ─── */
      .picker-footer {
        animation: fadeSlideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) 0.4s both;
      }

      .logout-link {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        color: #475569;
        font-size: 0.875rem;
        font-family: 'DM Sans', sans-serif;
        font-weight: 400;
        background: none;
        border: none;
        cursor: pointer;
        padding: 8px 12px;
        border-radius: 8px;
        transition:
          color 0.2s ease,
          background 0.2s ease;
      }

      .logout-link:hover {
        color: #94a3b8;
        background: rgba(255, 255, 255, 0.04);
      }

      /* ─── Animations ─── */
      @keyframes fadeSlideDown {
        from {
          opacity: 0;
          transform: translateY(-16px);
        }
        to {
          opacity: 1;
          transform: translateY(0);
        }
      }

      @keyframes fadeSlideUp {
        from {
          opacity: 0;
          transform: translateY(16px);
        }
        to {
          opacity: 1;
          transform: translateY(0);
        }
      }

      @keyframes shimmer {
        0% {
          background-position: 200% 0;
        }
        100% {
          background-position: -200% 0;
        }
      }

      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }

      /* ─── Responsive ─── */
      @media (max-width: 768px) {
        .picker-page {
          padding: 32px 16px;
        }

        .picker-wrapper {
          gap: 32px;
        }
      }

      @media (max-width: 480px) {
        .picker-title {
          font-size: 1.75rem;
        }

        .picker-subtitle {
          font-size: 0.875rem;
        }

        .org-card {
          padding: 16px 18px;
          gap: 14px;
        }

        .org-avatar {
          width: 42px;
          height: 42px;
        }

        .org-name {
          font-size: 0.9375rem;
        }

        .logout-link {
          padding: 12px 16px;
          min-height: 40px;
        }
      }
    `,
  ],
})
export class OrganizationPicker implements OnInit, OnDestroy {
  private store = inject(Store<AppState>);
  private router = inject(Router);
  private destroy$ = new Subject<void>();

  organizations = signal<UserOrganization[]>([]);
  isSwitching = signal(false);
  switchingOrgId = signal<number | null>(null);

  private readonly ORG_GRADIENTS = [
    'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
    'linear-gradient(135deg, #f59e0b 0%, #ef4444 100%)',
    'linear-gradient(135deg, #10b981 0%, #059669 100%)',
    'linear-gradient(135deg, #3b82f6 0%, #6366f1 100%)',
    'linear-gradient(135deg, #ec4899 0%, #8b5cf6 100%)',
    'linear-gradient(135deg, #14b8a6 0%, #3b82f6 100%)',
    'linear-gradient(135deg, #f97316 0%, #ef4444 100%)',
    'linear-gradient(135deg, #a855f7 0%, #6366f1 100%)',
  ];

  ngOnInit() {
    this.store
      .select(AuthSelectors.selectUserOrganizations)
      .pipe(takeUntil(this.destroy$))
      .subscribe((orgs) => {
        this.organizations.set(orgs || []);

        // If only one org, auto-select and navigate
        if (orgs && orgs.length === 1) {
          this.selectOrganization(orgs[0]);
        }
      });

    // Listen for successful org switch to stop loading
    this.store
      .select(AuthSelectors.selectCurrentOrganizationId)
      .pipe(
        takeUntil(this.destroy$),
        filter((id) => id !== null)
      )
      .subscribe(() => {
        this.isSwitching.set(false);
        this.switchingOrgId.set(null);
      });
  }

  selectOrganization(org: UserOrganization) {
    if (this.isSwitching()) return;

    this.isSwitching.set(true);
    this.switchingOrgId.set(org.id);
    this.store.dispatch(AuthActions.switchOrganization({ organizationId: org.organization_id }));
  }

  logout() {
    this.store.dispatch(AuthActions.logout());
  }

  getOrgInitial(name: string): string {
    return name?.charAt(0)?.toUpperCase() || '?';
  }

  getOrgGradient(name: string, index: number): string {
    // Deterministic gradient based on name + index
    const charSum = name?.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0) || 0;
    return this.ORG_GRADIENTS[(charSum + index) % this.ORG_GRADIENTS.length];
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
