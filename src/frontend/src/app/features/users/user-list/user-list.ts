import { Component } from '@angular/core';

import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-user-list',
  standalone: true,
  imports: [MatCardModule],
  templateUrl: './user-list.html',
  styleUrls: ['./user-list.scss'],
})
export class UserList {}
