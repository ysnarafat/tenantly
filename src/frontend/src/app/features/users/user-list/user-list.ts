import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-user-list',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  templateUrl: './user-list.html',
  styleUrls: ['./user-list.scss'],
})
export class UserList {}
