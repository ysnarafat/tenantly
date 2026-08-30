import {
  AfterViewInit,
  ChangeDetectorRef,
  Component,
  ContentChildren,
  DestroyRef,
  EventEmitter,
  Input,
  OnChanges,
  OnDestroy,
  Output,
  QueryList,
  SimpleChanges,
  ViewChild,
  inject,
} from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import {
  MatColumnDef,
  MatTable,
  MatTableDataSource,
  MatTableModule,
} from '@angular/material/table';
import { MatPaginator, MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatSort, MatSortModule, Sort } from '@angular/material/sort';
import { MatIconModule } from '@angular/material/icon';
import { LoadingSpinner } from '../loading-spinner/loading-spinner';

@Component({
  selector: 'app-data-table',
  standalone: true,
  imports: [
    CommonModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatIconModule,
    LoadingSpinner,
  ],
  templateUrl: './data-table.html',
  styleUrl: './data-table.scss',
})
export class DataTable<T = unknown> implements OnChanges, AfterViewInit, OnDestroy {
  private cdr = inject(ChangeDetectorRef);
  private destroyRef = inject(DestroyRef);
  // Consumers that use `mat-sort-header` in their projected column templates must put
  // `matSort` on the <app-data-table> tag itself, not just rely on this component's own
  // template — MatSortHeader resolves its MatSort via the injector chain of the template
  // where it's *declared*, which for ng-content-projected content is the consumer's
  // template, not this component's. Declaring matSort here on our own <table> would never
  // be visible to it.
  private injectedSort = inject(MatSort, { optional: true, self: true });
  @ContentChildren(MatColumnDef) columnDefs!: QueryList<MatColumnDef>;
  @ViewChild(MatTable) matTable!: MatTable<T>;
  @ViewChild(MatPaginator) matPaginator!: MatPaginator;

  @Input() dataSource: MatTableDataSource<T> | T[] = [];
  @Input() displayedColumns: string[] = [];
  @Input() loading = false;
  @Input() loadingMessage = 'Loading...';
  @Input() emptyIcon = 'inbox';
  @Input() emptyMessage = 'No data found';
  @Input() emptySubtitle = '';
  @Input() pageSizeOptions: number[] = [10, 25, 50];
  @Input() showPaginator = true;
  @Input() stickyHeader = true;

  @Output() pageChange = new EventEmitter<PageEvent>();
  @Output() sortChange = new EventEmitter<Sort>();

  internalDataSource = new MatTableDataSource<T>();
  renderedColumns: string[] = [];

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['dataSource']) {
      const ds = this.dataSource;
      if (ds instanceof MatTableDataSource) {
        this.internalDataSource = ds;
        // A consumer that hands us a fresh MatTableDataSource on every update (e.g. a
        // computed() wrapping a signal) swaps the instance out from under us — re-attach
        // sort/paginator, which otherwise only ever got assigned to the very first instance.
        if (this.injectedSort) {
          this.internalDataSource.sort = this.injectedSort;
        }
        if (this.showPaginator && this.matPaginator) {
          this.internalDataSource.paginator = this.matPaginator;
        }
      } else {
        this.internalDataSource.data = ds as T[];
      }
    }
    if (changes['displayedColumns'] && this.matTable) {
      this.renderedColumns = [...this.displayedColumns];
    }
  }

  ngAfterViewInit(): void {
    this.columnDefs.forEach((def) => this.matTable.addColumnDef(def));
    this.renderedColumns = [...this.displayedColumns];
    this.cdr.detectChanges();
    if (this.showPaginator && this.matPaginator) {
      this.internalDataSource.paginator = this.matPaginator;
    }
    if (this.injectedSort) {
      this.internalDataSource.sort = this.injectedSort;
      this.injectedSort.sortChange.pipe(takeUntilDestroyed(this.destroyRef)).subscribe((sort) => {
        this.sortChange.emit(sort);
      });
    }
  }

  ngOnDestroy(): void {
    if (this.matTable && this.columnDefs) {
      this.columnDefs.forEach((def) => this.matTable.removeColumnDef(def));
    }
    this.internalDataSource.paginator = undefined;
    this.internalDataSource.sort = undefined;
  }

  get isEmpty(): boolean {
    if (this.loading) return false;
    const data = this.internalDataSource.filteredData ?? this.internalDataSource.data;
    return data.length === 0;
  }
}
