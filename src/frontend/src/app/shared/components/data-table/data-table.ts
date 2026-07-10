import {
  AfterViewInit,
  Component,
  ContentChildren,
  EventEmitter,
  Input,
  OnChanges,
  Output,
  QueryList,
  SimpleChanges,
  ViewChild,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  MatColumnDef,
  MatTable,
  MatTableDataSource,
  MatTableModule,
} from '@angular/material/table';
import { MatPaginator, MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatSort, MatSortModule, Sort } from '@angular/material/sort';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-data-table',
  standalone: true,
  imports: [
    CommonModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatProgressSpinnerModule,
    MatIconModule,
  ],
  templateUrl: './data-table.html',
  styleUrl: './data-table.scss',
})
export class DataTable<T = unknown> implements OnChanges, AfterViewInit {
  @ContentChildren(MatColumnDef) columnDefs!: QueryList<MatColumnDef>;
  @ViewChild(MatTable) matTable!: MatTable<T>;
  @ViewChild(MatPaginator) matPaginator!: MatPaginator;
  @ViewChild(MatSort) matSort!: MatSort;

  @Input() dataSource: MatTableDataSource<T> | T[] = [];
  @Input() displayedColumns: string[] = [];
  @Input() loading = false;
  @Input() emptyIcon = 'inbox';
  @Input() emptyMessage = 'No data found';
  @Input() pageSizeOptions: number[] = [10, 25, 50];
  @Input() showPaginator = true;
  @Input() stickyHeader = true;

  @Output() pageChange = new EventEmitter<PageEvent>();
  @Output() sortChange = new EventEmitter<Sort>();

  internalDataSource = new MatTableDataSource<T>();

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['dataSource']) {
      const ds = this.dataSource;
      if (ds instanceof MatTableDataSource) {
        this.internalDataSource = ds;
      } else {
        this.internalDataSource.data = ds as T[];
      }
    }
  }

  ngAfterViewInit(): void {
    this.columnDefs.forEach(def => this.matTable.addColumnDef(def));
    if (this.showPaginator && this.matPaginator) {
      this.internalDataSource.paginator = this.matPaginator;
    }
    if (this.matSort) {
      this.internalDataSource.sort = this.matSort;
    }
  }

  get isEmpty(): boolean {
    if (this.loading) return false;
    const data = this.internalDataSource.filteredData ?? this.internalDataSource.data;
    return data.length === 0;
  }
}
