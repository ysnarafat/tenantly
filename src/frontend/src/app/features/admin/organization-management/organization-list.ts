import {
  Component,
  OnInit,
  AfterViewInit,
  ViewChild,
  inject,
  signal,
  computed,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatPaginatorModule, MatPaginator } from '@angular/material/paginator';
import { MatSortModule, MatSort } from '@angular/material/sort';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Organization } from '../../../core/models';
import { OrganizationService } from '../../../core/services/organization.service';

@Component({
  selector: 'app-organization-list',
  standalone: true,
  imports: [
    CommonModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatFormFieldModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    MatChipsModule,
    MatTooltipModule,
  ],
  templateUrl: './organization-list.html',
  styleUrls: ['./organization-list.scss'],
})
export class OrganizationListComponent implements OnInit, AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

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

  ngAfterViewInit(): void {
    if (this.paginator) {
      this.dataSource.paginator = this.paginator;
    }
    if (this.sort) {
      this.dataSource.sort = this.sort;
    }
  }

  loadOrganizations(): void {
    this.loading.set(true);
    this.organizationService.getOrganizations().subscribe({
      next: (response) => {
        this.dataSource.data = response.organizations;
        this.loading.set(false);
      },
      error: (error) => {
        console.error('Error loading organizations:', error);
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
    if (confirm(`Are you sure you want to delete organization "${org.name}"?`)) {
      this.organizationService.deleteOrganization(org.id).subscribe({
        next: () => {
          this.snackBar.open('Organization deleted successfully', 'Close', { duration: 3000 });
          this.loadOrganizations();
        },
        error: (error) => {
          console.error('Error deleting organization:', error);
          this.snackBar.open('Failed to delete organization', 'Close', { duration: 3000 });
        },
      });
    }
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
