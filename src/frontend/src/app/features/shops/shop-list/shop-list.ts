import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-shop-list',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  templateUrl: './shop-list.html',
  styleUrls: ['./shop-list.scss'],
})
export class ShopList {}
