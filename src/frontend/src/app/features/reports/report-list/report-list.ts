import { Component } from '@angular/core';

import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-report-list',
  standalone: true,
  imports: [MatCardModule],
  templateUrl: './report-list.html',
  styleUrls: ['./report-list.scss'],
})
export class ReportList {}
