import { Component, inject } from '@angular/core';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { TenantService } from '../../../core/services/tenant.service';
import { TenantFormDialogComponent } from '../tenant-form-dialog/tenant-form-dialog';
import { AuthFacade } from '../../../store/auth/auth.facade';
import { take } from 'rxjs';

@Component({
  selector: 'app-tenant-list',
  standalone: true,
  imports: [MatCardModule, MatButtonModule, MatIconModule, MatDialogModule],
  templateUrl: './tenant-list.html',
  styleUrls: ['./tenant-list.scss'],
})
export class TenantList {
  private dialog = inject(MatDialog);
  private tenantService = inject(TenantService);
  private authFacade = inject(AuthFacade);

  addTenant() {
    const dialogRef = this.dialog.open(TenantFormDialogComponent, {
      width: '600px',
      data: { mode: 'create' },
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.authFacade.user$.pipe(take(1)).subscribe((user) => {
          const userID = user?.id || 0;
          this.tenantService.createTenant(result).subscribe({
            next: (tenant) => {
              console.log('Tenant created successfully:', tenant);
              // In a real scenario, we might want to refresh the list or show a snackbar
            },
            error: (err) => {
              console.error('Error creating tenant:', err);
            },
          });
        });
      }
    });
  }
}
