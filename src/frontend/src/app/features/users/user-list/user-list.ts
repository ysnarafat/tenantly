import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatSortModule } from '@angular/material/sort';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatChipsModule } from '@angular/material/chips';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatDialog } from '@angular/material/dialog';
import { TranslateModule } from '@ngx-translate/core';
import { Router } from '@angular/router';
import { User } from '../../../core/services/auth.service';
import { UserService } from '../../../core/services/user.service';
import { PermissionService } from '../../../core/services/permission.service';
import { DataTable } from '../../../shared/components/data-table/data-table';
import { ResetPasswordDialogComponent } from '../reset-password-dialog/reset-password-dialog';
import { ConfirmDialogComponent } from '../../../shared/components/confirm-dialog/confirm-dialog';
import { safeErrorMessage } from '../../../shared/utils/error.utils';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';
import { actWithUndo } from '../../../shared/utils/undo-toast.utils';

@Component({
  selector: 'app-user-list',
  standalone: true,
  imports: [
    CommonModule,
    MatTableModule,
    MatSortModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatFormFieldModule,
    MatChipsModule,
    MatSelectModule,
    MatSnackBarModule,
    MatTooltipModule,
    MatSlideToggleModule,
    TranslateModule,
    DataTable,
  ],
  templateUrl: './user-list.html',
  styleUrls: ['./user-list.scss'],
})
export class UserList implements OnInit {
  private userService = inject(UserService);
  private permissionService = inject(PermissionService);
  private snackBar = inject(MatSnackBar);
  private router = inject(Router);
  private dialog = inject(MatDialog);

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
  activeCount = computed(() => this.allUsers().filter((u) => u.active).length);

  ngOnInit(): void {
    this.loadUsers();
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
        console.error('Error loading users:', safeErrorMessage(error));
        notifyError(this.snackBar, 'Failed to load users');
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

  private replaceUser(id: number, updated: User): void {
    this.allUsers.update((users) => users.map((u) => (u.id === id ? updated : u)));
  }

  deactivateUser(user: User): void {
    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      width: '400px',
      data: {
        title: 'Deactivate user',
        message: `${user.username} will be immediately signed out and won't be able to log back in until reactivated. You can reactivate them anytime.`,
        confirmLabel: 'Deactivate',
        tone: 'warn',
      },
    });

    dialogRef.afterClosed().subscribe((confirmed) => {
      if (!confirmed) return;

      this.userService.updateUser(user.id, { active: false }).subscribe({
        next: () => {
          notifySuccess(this.snackBar, `${user.username} deactivated`);
          if (!this.showInactive()) {
            this.loadUsers();
          } else {
            this.replaceUser(user.id, { ...user, active: false });
          }
        },
        error: (err) => {
          notifyError(this.snackBar, err.error?.error || 'Failed to deactivate user');
        },
      });
    });
  }

  activateUser(user: User): void {
    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      width: '400px',
      data: {
        title: 'Activate user',
        message: `${user.username} will be able to log in again immediately.`,
        confirmLabel: 'Activate',
      },
    });

    dialogRef.afterClosed().subscribe((confirmed) => {
      if (!confirmed) return;

      this.userService.updateUser(user.id, { active: true }).subscribe({
        next: () => {
          notifySuccess(this.snackBar, `${user.username} activated`);
          this.replaceUser(user.id, { ...user, active: true });
        },
        error: (err) => {
          notifyError(this.snackBar, err.error?.error || 'Failed to activate user');
        },
      });
    });
  }

  deleteUser(user: User): void {
    const previousUsers = this.allUsers();
    this.allUsers.set(previousUsers.filter((u) => u.id !== user.id));

    actWithUndo(
      this.snackBar,
      `${user.username} deleted`,
      () => {
        this.userService.deleteUser(user.id).subscribe({
          error: (err) => {
            notifyError(this.snackBar, err.error?.error || 'Failed to delete user');
            this.loadUsers();
          },
        });
      },
      { onUndo: () => this.allUsers.set(previousUsers) }
    );
  }

  promoteUser(user: User): void {
    this.router.navigate(['/admin/users/promote'], { queryParams: { userId: user.id } });
  }

  resetPassword(user: User): void {
    const dialogRef = this.dialog.open(ResetPasswordDialogComponent, {
      width: '420px',
      data: { username: user.username },
    });

    dialogRef.afterClosed().subscribe((newPassword: string | null) => {
      if (!newPassword) return;

      this.userService.adminResetPassword(user.id, newPassword).subscribe({
        next: () => {
          notifySuccess(this.snackBar, `Password reset for ${user.username}`);
        },
        error: (err) => {
          notifyError(this.snackBar, err.error?.error || 'Failed to reset password');
        },
      });
    });
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
