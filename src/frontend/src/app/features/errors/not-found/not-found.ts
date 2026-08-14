import { Component, inject } from '@angular/core';
import { Location } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-not-found',
  standalone: true,
  imports: [RouterLink, MatIconModule],
  templateUrl: './not-found.html',
  styleUrls: ['./not-found.scss'],
})
export class NotFound {
  private location = inject(Location);
  private authService = inject(AuthService);

  homeLink = this.authService.isAuthenticated() ? '/dashboard' : '/login';

  goBack(): void {
    this.location.back();
  }
}
