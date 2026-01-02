import { Component } from '@angular/core';

import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-payment-list',
  standalone: true,
  imports: [MatCardModule],
  templateUrl: './payment-list.html',
  styleUrls: ['./payment-list.scss'],
})
export class PaymentList {}
