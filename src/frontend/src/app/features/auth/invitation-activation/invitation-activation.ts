import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, inject, signal } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  ValidatorFn,
  Validators,
} from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { Store } from '@ngrx/store';
import * as AuthActions from '../../../store/auth/auth.actions';
import { AppState } from '../../../store';
import { UserInvitation } from '../../../core/models';
import { UserInvitationService } from '../../../core/services/user-invitation.service';

const passwordPolicy =
  /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]).{8,128}$/;

const matchingPasswords: ValidatorFn = (control: AbstractControl): ValidationErrors | null =>
  control.get('password')?.value === control.get('confirmPassword')?.value
    ? null
    : { passwordMismatch: true };

@Component({
  selector: 'app-invitation-activation',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, TranslateModule],
  templateUrl: './invitation-activation.html',
  styleUrl: './invitation-activation.scss',
})
export class InvitationActivation implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly invitations = inject(UserInvitationService);
  private readonly store = inject(Store<AppState>);
  private readonly formBuilder = inject(FormBuilder);

  readonly invitation = signal<UserInvitation | null>(null);
  readonly loading = signal(true);
  readonly submitting = signal(false);
  readonly error = signal('');

  readonly form = this.formBuilder.group(
    {
      firstName: ['', [Validators.required, Validators.maxLength(100)]],
      lastName: ['', [Validators.required, Validators.maxLength(100)]],
      password: [
        '',
        [
          Validators.required,
          Validators.minLength(8),
          Validators.maxLength(72),
          Validators.pattern(passwordPolicy),
        ],
      ],
      confirmPassword: ['', Validators.required],
    },
    { validators: matchingPasswords }
  );

  private invitationToken = '';

  ngOnInit(): void {
    this.invitationToken = this.route.snapshot.queryParamMap.get('token') ?? '';
    if (!this.invitationToken) {
      this.error.set('INVITATION_ACTIVATION.ERROR_INVALID');
      this.loading.set(false);
      return;
    }

    const url = new URL(window.location.href);
    url.searchParams.delete('token');
    window.history.replaceState(null, '', `${url.pathname}${url.search}${url.hash}`);

    this.invitations.validateInvitationToken(this.invitationToken).subscribe({
      next: (invitation) => {
        this.invitation.set(invitation);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('INVITATION_ACTIVATION.ERROR_INVALID');
        this.loading.set(false);
      },
    });
  }

  activate(): void {
    const invitation = this.invitation();
    if (!invitation || this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.error.set('');
    this.submitting.set(true);
    this.invitations
      .registerWithInvitation({
        email: invitation.email,
        first_name: this.form.controls.firstName.value ?? '',
        last_name: this.form.controls.lastName.value ?? '',
        password: this.form.controls.password.value ?? '',
        invitation_token: this.invitationToken,
      })
      .subscribe({
        next: (response) => this.store.dispatch(AuthActions.loginSuccess({ response })),
        error: (error: HttpErrorResponse) => {
          this.submitting.set(false);
          this.error.set(error.error?.error ?? 'INVITATION_ACTIVATION.ERROR_SUBMIT');
        },
      });
  }
}
