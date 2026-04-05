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
import { MatDividerModule } from '@angular/material/divider';
import { Store } from '@ngrx/store';
import { toSignal } from '@angular/core/rxjs-interop';
import { UserService } from '../../../core/services/user.service';
import { OrganizationService } from '../../../core/services/organization.service';
import { Organization } from '../../../core/models';
import { AppState } from '../../../store';
import * as AuthSelectors from '../../../store/auth/auth.selectors';

@Component({
  selector: 'app-create-user',
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
    MatDividerModule,
  ],
  template: `
    <div class="create-user-container">
      <mat-card>
        <mat-card-header>
          <mat-card-title>
            <mat-icon>person_add</mat-icon>
            Create New User
          </mat-card-title>
          <mat-card-subtitle>
            @if (isSuperAdmin()) {
              Create account directly in selected organization.
            } @else {
              Creating account in <strong>{{ callerOrgName() }}</strong>
            }
          </mat-card-subtitle>
        </mat-card-header>
        <mat-divider></mat-divider>
        <mat-card-content>
          <form [formGroup]="form" (ngSubmit)="onSubmit()">
            <h3 class="section-title">Personal Information</h3>
            <div class="name-row">
              <mat-form-field appearance="outline">
                <mat-label>First Name</mat-label>
                <input matInput formControlName="first_name" />
                <mat-error>First name required</mat-error>
              </mat-form-field>
              <mat-form-field appearance="outline">
                <mat-label>Last Name</mat-label>
                <input matInput formControlName="last_name" />
                <mat-error>Last name required</mat-error>
              </mat-form-field>
            </div>

            <h3 class="section-title">Account Details</h3>
            <mat-form-field appearance="outline">
              <mat-label>Username</mat-label>
              <input matInput formControlName="username" placeholder="alphanumeric and underscore" />
              <mat-error>Username required (min 3 chars, alphanumeric + underscore)</mat-error>
            </mat-form-field>

            <mat-form-field appearance="outline">
              <mat-label>Email</mat-label>
              <input matInput type="email" formControlName="email" />
              <mat-error>Valid email required</mat-error>
            </mat-form-field>

            <mat-form-field appearance="outline">
              <mat-label>Password</mat-label>
              <input matInput type="password" formControlName="password" />
              <mat-hint>Min 8 chars, uppercase, lowercase, digit, special character</mat-hint>
              <mat-error>Password required</mat-error>
            </mat-form-field>

            <h3 class="section-title">Role & Organization</h3>
            <mat-form-field appearance="outline">
              <mat-label>Role</mat-label>
              <mat-select formControlName="role" required>
                @for (role of availableRoles(); track role.value) {
                  <mat-option [value]="role.value">{{ role.label }}</mat-option>
                }
              </mat-select>
              <mat-error>Role required</mat-error>
            </mat-form-field>

            @if (isSuperAdmin()) {
              <mat-form-field appearance="outline">
                <mat-label>Organization (optional)</mat-label>
                <mat-select formControlName="organization_id">
                  <mat-option [value]="null">None</mat-option>
                  @for (org of organizations(); track org.id) {
                    <mat-option [value]="org.id">{{ org.name }}</mat-option>
                  }
                </mat-select>
              </mat-form-field>
            }

            <div class="actions">
              <button mat-button type="button" (click)="cancel()">Cancel</button>
              <button mat-raised-button color="primary" type="submit" [disabled]="submitting()">
                @if (submitting()) {
                  <mat-spinner diameter="20"></mat-spinner>
                } @else {
                  Create User
                }
              </button>
            </div>
          </form>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .create-user-container {
      max-width: 640px;
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
    .section-title {
      margin: 8px 0 0;
      font-size: 14px;
      font-weight: 600;
      color: var(--text-secondary, #666);
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    .name-row {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 16px;
    }
    mat-form-field { width: 100%; }
    .actions {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
      padding-top: 8px;
    }
    @media (max-width: 480px) {
      .name-row { grid-template-columns: 1fr; }
    }
  `],
})
export class CreateUserComponent implements OnInit {
  private fb = inject(FormBuilder);
  private userService = inject(UserService);
  private organizationService = inject(OrganizationService);
  private snackBar = inject(MatSnackBar);
  private router = inject(Router);
  private store = inject(Store<AppState>);

  submitting = signal(false);
  organizations = signal<Organization[]>([]);
  callerOrgName = signal<string>('');

  private callerUser = toSignal(this.store.select(AuthSelectors.selectUser));
  isSuperAdmin = computed(() => this.callerUser()?.role === 'SUPER_ADMIN');

  allRoles = [
    { value: 'SUPER_ADMIN', label: 'Super Admin' },
    { value: 'ORG_ADMIN', label: 'Organization Admin' },
    { value: 'Admin', label: 'Admin' },
    { value: 'PropertyManager', label: 'Property Manager' },
    { value: 'Accountant', label: 'Accountant' },
  ];

  availableRoles = signal(this.allRoles);

  form: FormGroup = this.fb.group({
    first_name: ['', Validators.required],
    last_name: ['', Validators.required],
    username: ['', [Validators.required, Validators.minLength(3), Validators.pattern(/^[a-zA-Z0-9_]+$/)]],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
    role: ['', Validators.required],
    organization_id: [null],
  });

  ngOnInit(): void {
    this.store.select(AuthSelectors.selectUser).subscribe((user) => {
      if (!user) return;
      const filtered = this.allRoles.filter((r) => this.canAssignRole(user.role, r.value));
      this.availableRoles.set(filtered);

      if (user.role !== 'SUPER_ADMIN' && user.organization_id) {
        this.form.patchValue({ organization_id: user.organization_id });
      }
    });

    this.store.select(AuthSelectors.selectIsSuperAdmin).subscribe((isSA) => {
      if (isSA) {
        this.organizationService.getOrganizations().subscribe({
          next: (res) => this.organizations.set(res.organizations),
          error: () => this.snackBar.open('Failed to load organizations', 'Close', { duration: 3000 }),
        });
      } else {
        this.store.select(AuthSelectors.selectUserOrganizationId).subscribe((orgId) => {
          if (orgId) {
            this.organizationService.getOrganization(orgId).subscribe({
              next: (org) => this.callerOrgName.set(org.name),
              error: () => this.snackBar.open('Failed to load organization', 'Close', { duration: 3000 }),
            });
          }
        });
      }
    });
  }

  canAssignRole(callerRole: string, targetRole: string): boolean {
    if (callerRole === 'SUPER_ADMIN') return true;
    if (callerRole === 'ORG_ADMIN') return ['Admin', 'PropertyManager', 'Accountant'].includes(targetRole);
    if (callerRole === 'Admin') return ['PropertyManager', 'Accountant'].includes(targetRole);
    return false;
  }

  onSubmit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.submitting.set(true);
    this.userService.createUser(this.form.value).subscribe({
      next: () => {
        this.snackBar.open('User created successfully', 'Close', { duration: 3000 });
        this.router.navigate(['/admin/users']);
      },
      error: (err) => {
        this.snackBar.open(err.error?.error || 'Failed to create user', 'Close', { duration: 5000 });
        this.submitting.set(false);
      },
    });
  }

  cancel(): void {
    this.router.navigate(['/admin/users']);
  }
}
