import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-report-list',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  templateUrl: './report-list.html',
  styleUrls: ['./report-list.scss'],
})
export class ReportList {}
