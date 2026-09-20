import { Component, inject } from '@angular/core';
import { Location } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-unauthorized',
  standalone: true,
  imports: [RouterLink, MatIconModule, MatButtonModule],
  templateUrl: './unauthorized.html',
  styleUrls: ['./unauthorized.scss'],
})
export class Unauthorized {
  private location = inject(Location);
  private authService = inject(AuthService);

  homeLink = this.authService.isAuthenticated() ? '/dashboard' : '/login';
  homeLabel = this.authService.isAuthenticated() ? 'Go to dashboard' : 'Go to login';

  goBack(): void {
    this.location.back();
  }
}
