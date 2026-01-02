import { Component } from '@angular/core';

import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-tenant-list',
  standalone: true,
  imports: [MatCardModule],
  templateUrl: './tenant-list.html',
  styleUrls: ['./tenant-list.scss'],
})
export class TenantList {}
