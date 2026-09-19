import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { TranslateModule } from '@ngx-translate/core';
import { CreateOrgRequest } from '../../../core/models';
import { OrganizationService } from '../../../core/services/organization.service';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

@Component({
  selector: 'app-organization-create',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    TranslateModule,
  ],
  templateUrl: './organization-create.html',
  styleUrls: ['./organization-create.scss'],
})
export class OrganizationCreateComponent implements OnInit {
  private fb = inject(FormBuilder);
  private organizationService = inject(OrganizationService);
  private router = inject(Router);
  private snackBar = inject(MatSnackBar);

  form!: FormGroup;
  loading = false;

  ngOnInit(): void {
    this.initializeForm();
  }

  initializeForm(): void {
    this.form = this.fb.group({
      name: ['', [Validators.required, Validators.minLength(3)]],
      slug: ['', Validators.required],
      subscriptionTier: ['basic', Validators.required],
      maxUsers: [10, [Validators.required, Validators.min(1)]],
    });

    // Auto-generate slug from name
    this.form.get('name')?.valueChanges.subscribe((name: string) => {
      const slug = name
        .toLowerCase()
        .replace(/\s+/g, '-')
        .replace(/[^\w-]/g, '');
      this.form.patchValue({ slug }, { emitEvent: false });
    });
  }

  submit(): void {
    if (this.form.invalid) {
      return;
    }

    this.loading = true;
    const req: CreateOrgRequest = this.form.value;

    this.organizationService.createOrganization(req).subscribe({
      next: () => {
        notifySuccess(this.snackBar, 'Organization created successfully');
        this.router.navigate(['/admin/organizations']);
      },
      error: (error) => {
        console.error('Error creating organization:', safeErrorMessage(error));
        notifyError(this.snackBar, 'Failed to create organization');
        this.loading = false;
      },
    });
  }

  cancel(): void {
    this.router.navigate(['/admin/organizations']);
  }
}
