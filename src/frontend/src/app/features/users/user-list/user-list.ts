import {
  Component,
  OnInit,
  ViewChild,
  inject,
  signal,
  computed,
  AfterViewInit,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatPaginatorModule, MatPaginator } from '@angular/material/paginator';
import { MatSortModule, MatSort } from '@angular/material/sort';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatChipsModule } from '@angular/material/chips';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { User } from '../../../core/services/auth.service';
import { OrganizationService } from '../../../core/services/organization.service';
import { PermissionService } from '../../../core/services/permission.service';

@Component({
  selector: 'app-user-list',
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
    MatChipsModule,
    MatSelectModule,
    MatSnackBarModule,
    MatTooltipModule,
  ],
  templateUrl: './user-list.html',
  styleUrls: ['./user-list.scss'],
})
export class UserList implements OnInit, AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  private organizationService = inject(OrganizationService);
  private permissionService = inject(PermissionService);
  private snackBar = inject(MatSnackBar);

  loading = signal(false);
  searchTerm = signal('');
  selectedRole = signal<string | null>(null);
  dataSource = new MatTableDataSource<User>();
  displayedColumns: string[] = ['username', 'email', 'role', 'active', 'created_at', 'actions'];

  roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'PropertyManager', 'Accountant'];

  filteredData = computed(() => {
    const search = this.searchTerm().toLowerCase();
    const role = this.selectedRole();

    return this.dataSource.data.filter((user) => {
      const matchesSearch =
        user.username.toLowerCase().includes(search) || user.email.toLowerCase().includes(search);
      const matchesRole = !role || user.role === role;
      return matchesSearch && matchesRole;
    });
  });

  canPromoteUsers = computed(() => this.permissionService.canManageOrgAdmins());

  ngOnInit(): void {
    this.loadUsers();
  }

  ngAfterViewInit(): void {
    if (this.paginator) {
      this.dataSource.paginator = this.paginator;
    }
    if (this.sort) {
      this.dataSource.sort = this.sort;
    }
  }

  loadUsers(): void {
    const currentOrg = this.organizationService.getCurrentOrganization();
    if (!currentOrg) {
      this.snackBar.open('No organization selected', 'Close', { duration: 3000 });
      return;
    }

    this.loading.set(true);
    this.organizationService.getOrganizationUsers(currentOrg.id).subscribe({
      next: (response) => {
        this.dataSource.data = response.users;
        this.loading.set(false);
      },
      error: (error) => {
        console.error('Error loading users:', error);
        this.snackBar.open('Failed to load users', 'Close', { duration: 3000 });
        this.loading.set(false);
      },
    });
  }

  onSearchChange(value: string): void {
    this.searchTerm.set(value);
  }

  onRoleFilterChange(role: string | null): void {
    this.selectedRole.set(role);
  }

  deactivateUser(user: User): void {
    if (confirm(`Are you sure you want to deactivate ${user.username}?`)) {
      this.snackBar.open('User deactivation would be implemented here', 'Close', {
        duration: 3000,
      });
    }
  }

  deleteUser(user: User): void {
    if (confirm(`Are you sure you want to delete ${user.username}?`)) {
      this.snackBar.open('User deletion would be implemented here', 'Close', { duration: 3000 });
    }
  }

  promoteUser(user: User): void {
    void user; // Suppress unused variable warning
    this.snackBar.open('User promotion dialog would open here', 'Close', { duration: 3000 });
  }

  getRoleColor(role: string): string {
    switch (role) {
      case 'SUPER_ADMIN':
        return 'warn';
      case 'ORG_ADMIN':
        return 'accent';
      case 'Admin':
        return 'primary';
      case 'PropertyManager':
        return 'primary';
      case 'Accountant':
        return 'accent';
      default:
        return '';
    }
  }
}
