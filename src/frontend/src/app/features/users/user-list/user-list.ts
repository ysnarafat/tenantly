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
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { TranslateModule } from '@ngx-translate/core';
import { Router } from '@angular/router';
import { User } from '../../../core/services/auth.service';
import { UserService } from '../../../core/services/user.service';
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
    MatSlideToggleModule,
    TranslateModule,
  ],
  templateUrl: './user-list.html',
  styleUrls: ['./user-list.scss'],
})
export class UserList implements OnInit, AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  private userService = inject(UserService);
  private permissionService = inject(PermissionService);
  private snackBar = inject(MatSnackBar);
  private router = inject(Router);

  loading = signal(false);
  searchTerm = signal('');
  selectedRole = signal<string | null>(null);
  showInactive = signal(false);
  allUsers = signal<User[]>([]);
  dataSource = new MatTableDataSource<User>();
  displayedColumns: string[] = ['username', 'email', 'role', 'active', 'created_at', 'actions'];

  roles = ['SUPER_ADMIN', 'ORG_ADMIN', 'Admin', 'PropertyManager', 'Accountant'];

  filteredData = computed(() => {
    const search = this.searchTerm().toLowerCase();
    const role = this.selectedRole();

    return this.allUsers().filter((user) => {
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
    this.loading.set(true);
    this.userService.getUsers(!this.showInactive()).subscribe({
      next: (response) => {
        this.allUsers.set(response.users);
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

  toggleShowInactive(checked: boolean): void {
    this.showInactive.set(checked);
    this.loadUsers();
  }

  onSearchChange(value: string): void {
    this.searchTerm.set(value);
  }

  onRoleFilterChange(role: string | null): void {
    this.selectedRole.set(role);
  }

  deactivateUser(user: User): void {
    if (!confirm(`Deactivate ${user.username}?`)) return;
    this.userService.updateUser(user.id, { active: false }).subscribe({
      next: () => {
        this.snackBar.open('User deactivated', 'Close', { duration: 3000 });
        this.loadUsers();
      },
      error: (err) =>
        this.snackBar.open(err.error?.error || 'Failed to deactivate user', 'Close', {
          duration: 5000,
        }),
    });
  }

  activateUser(user: User): void {
    if (!confirm(`Activate ${user.username}?`)) return;
    this.userService.updateUser(user.id, { active: true }).subscribe({
      next: () => {
        this.snackBar.open('User activated', 'Close', { duration: 3000 });
        this.loadUsers();
      },
      error: (err) =>
        this.snackBar.open(err.error?.error || 'Failed to activate user', 'Close', {
          duration: 5000,
        }),
    });
  }

  deleteUser(user: User): void {
    if (!confirm(`Permanently delete ${user.username}? This cannot be undone.`)) return;
    this.userService.deleteUser(user.id).subscribe({
      next: () => {
        this.snackBar.open('User deleted', 'Close', { duration: 3000 });
        this.loadUsers();
      },
      error: (err) =>
        this.snackBar.open(err.error?.error || 'Failed to delete user', 'Close', {
          duration: 5000,
        }),
    });
  }

  promoteUser(user: User): void {
    this.router.navigate(['/admin/users/promote'], { queryParams: { userId: user.id } });
  }

  createUser(): void {
    this.router.navigate(['/admin/users/new']);
  }

  inviteUser(): void {
    this.router.navigate(['/admin/invitations/new']);
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
