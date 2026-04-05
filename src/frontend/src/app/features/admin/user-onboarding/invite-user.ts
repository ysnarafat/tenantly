import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { Store } from '@ngrx/store';
import { toSignal } from '@angular/core/rxjs-interop';
import { UserInvitationService } from '../../../core/services/user-invitation.service';
import { OrganizationService } from '../../../core/services/organization.service';
import { Organization } from '../../../core/models';
import { AppState } from '../../../store';
import * as AuthSelectors from '../../../store/auth/auth.selectors';

@Component({
  selector: 'app-invite-user',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatSnackBarModule,
    MatProgressSpinnerModule,
    MatCardModule,
    MatIconModule,
  ],
  template: `
    <div class="invite-container">
      <mat-card>
        <mat-card-header>
          <mat-card-title>
            <mat-icon>person_add</mat-icon>
            Invite User to Organization
          </mat-card-title>
          @if (!isSuperAdmin() && callerOrgName()) {
            <mat-card-subtitle>Inviting to <strong>{{ callerOrgName() }}</strong></mat-card-subtitle>
          }
        </mat-card-header>
        <mat-card-content>
          <form [formGroup]="form" (ngSubmit)="onSubmit()">
            @if (isSuperAdmin()) {
              <mat-form-field appearance="outline">
                <mat-label>Organization</mat-label>
                <mat-select formControlName="organizationId" required>
                  @for (org of organizations(); track org.id) {
                    <mat-option [value]="org.id">{{ org.name }}</mat-option>
                  }
                </mat-select>
                <mat-error>Organization is required</mat-error>
              </mat-form-field>
            }

            <mat-form-field appearance="outline">
              <mat-label>Email</mat-label>
              <input matInput type="email" formControlName="email" placeholder="user@example.com" />
              <mat-error>Valid email required</mat-error>
            </mat-form-field>

            <div class="name-row">
              <mat-form-field appearance="outline">
                <mat-label>First Name</mat-label>
                <input matInput formControlName="firstName" />
                <mat-error>First name required</mat-error>
              </mat-form-field>

              <mat-form-field appearance="outline">
                <mat-label>Last Name</mat-label>
                <input matInput formControlName="lastName" />
                <mat-error>Last name required</mat-error>
              </mat-form-field>
            </div>

            <mat-form-field appearance="outline">
              <mat-label>Role</mat-label>
              <mat-select formControlName="role" required>
                @for (role of allowedRoles; track role.value) {
                  <mat-option [value]="role.value">{{ role.label }}</mat-option>
                }
              </mat-select>
              <mat-error>Role is required</mat-error>
            </mat-form-field>

            <div class="actions">
              <button mat-button type="button" (click)="cancel()">Cancel</button>
              <button mat-raised-button color="primary" type="submit" [disabled]="submitting()">
                @if (submitting()) {
                  <mat-spinner diameter="20"></mat-spinner>
                } @else {
                  Send Invitation
                }
              </button>
            </div>
          </form>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .invite-container {
      max-width: 600px;
      margin: 24px auto;
      padding: 0 16px;
    }
    mat-card-title {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    form {
      display: flex;
      flex-direction: column;
      gap: 16px;
      padding-top: 16px;
    }
    .name-row {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 16px;
    }
    mat-form-field {
      width: 100%;
    }
    .actions {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
    }
    @media (max-width: 480px) {
      .name-row { grid-template-columns: 1fr; }
    }
  `],
})
export class InviteUserComponent implements OnInit {
  private fb = inject(FormBuilder);
  private invitationService = inject(UserInvitationService);
  private organizationService = inject(OrganizationService);
  private snackBar = inject(MatSnackBar);
  private router = inject(Router);
  private store = inject(Store<AppState>);

  submitting = signal(false);
  organizations = signal<Organization[]>([]);
  callerOrgName = signal<string>('');

  private callerUser = toSignal(this.store.select(AuthSelectors.selectUser));
  isSuperAdmin = computed(() => this.callerUser()?.role === 'SUPER_ADMIN');

  allowedRoles = [
    { value: 'Admin', label: 'Admin' },
    { value: 'PropertyManager', label: 'Property Manager' },
    { value: 'Accountant', label: 'Accountant' },
  ];

  form: FormGroup = this.fb.group({
    organizationId: [null, Validators.required],
    email: ['', [Validators.required, Validators.email]],
    firstName: ['', Validators.required],
    lastName: ['', Validators.required],
    role: ['', Validators.required],
  });

  ngOnInit(): void {
    this.store.select(AuthSelectors.selectUser).subscribe((user) => {
      if (!user) return;
      if (user.role === 'SUPER_ADMIN') {
        this.organizationService.getOrganizations().subscribe({
          next: (res) => this.organizations.set(res.organizations),
          error: () => this.snackBar.open('Failed to load organizations', 'Close', { duration: 3000 }),
        });
      } else if (user.organization_id) {
        this.form.patchValue({ organizationId: user.organization_id });
        this.organizationService.getOrganization(user.organization_id).subscribe({
          next: (org) => this.callerOrgName.set(org.name),
          error: () => this.snackBar.open('Failed to load organization', 'Close', { duration: 3000 }),
        });
      }
    });
  }

  onSubmit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const { organizationId, email, firstName, lastName, role } = this.form.value;
    this.submitting.set(true);

    this.invitationService.sendInvitation(organizationId, { email, firstName, lastName, role }).subscribe({
      next: () => {
        this.snackBar.open('Invitation sent successfully', 'Close', { duration: 3000 });
        this.router.navigate(['/admin/invitations']);
      },
      error: (err) => {
        this.snackBar.open(err.error?.error || 'Failed to send invitation', 'Close', { duration: 5000 });
        this.submitting.set(false);
      },
    });
  }

  cancel(): void {
    this.router.navigate(['/admin/invitations']);
  }
}
