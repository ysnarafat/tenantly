import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatSortModule } from '@angular/material/sort';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { TranslateModule } from '@ngx-translate/core';
import { Organization } from '../../../core/models';
import { OrganizationService } from '../../../core/services/organization.service';
import { DataTable } from '../../../shared/components/data-table/data-table';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { actWithUndo } from '../../../shared/utils/undo-toast.utils';

@Component({
  selector: 'app-organization-list',
  standalone: true,
  imports: [
    CommonModule,
    MatTableModule,
    MatSortModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatFormFieldModule,
    MatSnackBarModule,
    MatChipsModule,
    MatTooltipModule,
    TranslateModule,
    DataTable,
  ],
  templateUrl: './organization-list.html',
  styleUrls: ['./organization-list.scss'],
})
export class OrganizationListComponent implements OnInit {
  private organizationService = inject(OrganizationService);
  private router = inject(Router);
  private snackBar = inject(MatSnackBar);

  loading = signal(false);
  searchTerm = signal('');
  dataSource = new MatTableDataSource<Organization>();
  displayedColumns: string[] = [
    'name',
    'slug',
    'subscriptionTier',
    'maxUsers',
    'active',
    'actions',
  ];

  filteredData = computed(() => {
    const search = this.searchTerm().toLowerCase();
    return this.dataSource.data.filter(
      (org) => org.name.toLowerCase().includes(search) || org.slug.toLowerCase().includes(search)
    );
  });

  ngOnInit(): void {
    this.loadOrganizations();
  }

  loadOrganizations(): void {
    this.loading.set(true);
    this.organizationService.getOrganizations().subscribe({
      next: (response) => {
        this.dataSource.data = response.organizations;
        this.loading.set(false);
      },
      error: (error) => {
        console.error('Error loading organizations:', safeErrorMessage(error));
        this.snackBar.open('Failed to load organizations', 'Close', { duration: 3000 });
        this.loading.set(false);
      },
    });
  }

  onSearchChange(value: string): void {
    this.searchTerm.set(value);
    this.dataSource.filter = value.toLowerCase();
  }

  createOrganization(): void {
    this.router.navigate(['/admin/organizations/new']);
  }

  viewOrganization(org: Organization): void {
    this.router.navigate(['/admin/organizations', org.id]);
  }

  editOrganization(org: Organization): void {
    this.router.navigate(['/admin/organizations', org.id, 'edit']);
  }

  deleteOrganization(org: Organization): void {
    const previousData = this.dataSource.data;
    this.dataSource.data = previousData.filter((o) => o.id !== org.id);

    actWithUndo(
      this.snackBar,
      `Organization "${org.name}" deleted`,
      () => {
        this.organizationService.deleteOrganization(org.id).subscribe({
          error: (error) => {
            console.error('Error deleting organization:', safeErrorMessage(error));
            this.snackBar.open('Failed to delete organization', 'Close', { duration: 3000 });
            this.loadOrganizations();
          },
        });
      },
      { onUndo: () => (this.dataSource.data = previousData) }
    );
  }

  getSubscriptionTierColor(tier: string): string {
    switch (tier) {
      case 'basic':
        return 'primary';
      case 'professional':
        return 'accent';
      case 'enterprise':
        return 'warn';
      default:
        return '';
    }
  }
}
