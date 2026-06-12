import { Component, OnInit, inject, Input } from '@angular/core';

import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatChipsModule } from '@angular/material/chips';
import { MatMenuModule } from '@angular/material/menu';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatDialog } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatInputModule } from '@angular/material/input';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import {
  AttachmentService,
  Attachment,
  AttachmentType,
} from '../../../core/services/attachment.service';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-attachment-list',
  standalone: true,
  imports: [
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatCardModule,
    MatChipsModule,
    MatMenuModule,
    MatFormFieldModule,
    MatSelectModule,
    MatInputModule,
    FormsModule,
    TranslateModule,
  ],
  templateUrl: './attachment-list.html',
  styleUrls: ['./attachment-list.scss'],
})
export class AttachmentList implements OnInit {
  @Input() entityType?: string; // 'lease', 'shop', 'tenant', 'payment', 'property'
  @Input() entityId?: number;
  @Input() showHeader = true;

  private attachmentService = inject(AttachmentService);
  private authService = inject(AuthService);
  private snackBar = inject(MatSnackBar);
  private dialog = inject(MatDialog);

  attachments: Attachment[] = [];
  filteredAttachments: Attachment[] = [];
  displayedColumns: string[] = [
    'file_name',
    'attachment_type',
    'file_size',
    'description',
    'created_at',
    'actions',
  ];
  loading = false;

  // Filters
  selectedType: AttachmentType | 'ALL' = 'ALL';
  searchQuery = '';
  attachmentTypes = Object.values(AttachmentType);

  ngOnInit() {
    this.loadAttachments();
  }

  loadAttachments() {
    this.loading = true;

    let observable;
    if (this.entityType && this.entityId) {
      observable = this.attachmentService.getAttachmentsByEntity(this.entityType, this.entityId);
    } else {
      observable = this.attachmentService.getAllAttachments();
    }

    observable.subscribe({
      next: (attachments) => {
        this.attachments = attachments;
        this.applyFilters();
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading attachments:', error);
        this.snackBar.open('Error loading attachments', 'Close', { duration: 3000 });
        this.loading = false;
      },
    });
  }

  applyFilters() {
    this.filteredAttachments = this.attachments.filter((attachment) => {
      const matchesType =
        this.selectedType === 'ALL' || attachment.attachment_type === this.selectedType;
      const matchesSearch =
        !this.searchQuery ||
        attachment.file_name.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
        attachment.description?.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
        attachment.tags?.toLowerCase().includes(this.searchQuery.toLowerCase());

      return matchesType && matchesSearch;
    });
  }

  onTypeFilterChange() {
    this.applyFilters();
  }

  onSearchChange() {
    this.applyFilters();
  }

  getTypeColor(type: AttachmentType): string {
    switch (type) {
      case AttachmentType.CONTRACT:
        return 'primary';
      case AttachmentType.RECEIPT:
        return 'accent';
      case AttachmentType.INVOICE:
        return 'warn';
      case AttachmentType.PHOTO:
        return 'primary';
      case AttachmentType.DOCUMENT:
        return 'accent';
      default:
        return 'basic';
    }
  }

  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-BD');
  }

  getFileIcon(contentType: string): string {
    if (contentType.startsWith('image/')) return 'image';
    if (contentType.includes('pdf')) return 'picture_as_pdf';
    if (contentType.includes('word') || contentType.includes('document')) return 'description';
    if (contentType.includes('excel') || contentType.includes('spreadsheet')) return 'table_chart';
    return 'insert_drive_file';
  }

  canUpload(): boolean {
    return this.authService.isAdmin() || this.authService.isPropertyManager();
  }

  canEdit(): boolean {
    return this.authService.isAdmin() || this.authService.isPropertyManager();
  }

  canDelete(): boolean {
    return this.authService.isAdmin();
  }

  uploadAttachment() {
    // TODO: Implement upload attachment dialog
    this.snackBar.open('Upload attachment functionality will be implemented', 'Close', {
      duration: 3000,
    });
  }

  downloadAttachment(attachment: Attachment) {
    this.attachmentService.downloadAttachment(attachment.id).subscribe({
      next: (blob) => {
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = attachment.file_name;
        link.click();
        window.URL.revokeObjectURL(url);
      },
      error: (error) => {
        console.error('Error downloading attachment:', error);
        this.snackBar.open('Error downloading attachment', 'Close', { duration: 3000 });
      },
    });
  }

  viewAttachment(attachment: Attachment) {
    if (attachment.content_type.startsWith('image/')) {
      // TODO: Implement image viewer
      this.snackBar.open('Image viewer will be implemented', 'Close', { duration: 3000 });
    } else {
      this.downloadAttachment(attachment);
    }
  }

  editAttachment(attachment: Attachment): void {
    // TODO: Implement edit attachment dialog
    void attachment; // Suppress unused variable warning
    this.snackBar.open('Edit attachment functionality will be implemented', 'Close', {
      duration: 3000,
    });
  }

  deleteAttachment(attachment: Attachment) {
    if (confirm(`Are you sure you want to delete "${attachment.file_name}"?`)) {
      this.attachmentService.deleteAttachment(attachment.id).subscribe({
        next: () => {
          this.loadAttachments();
          this.snackBar.open('Attachment deleted successfully', 'Close', { duration: 3000 });
          this.loadAttachments();
        },
        error: (error) => {
          console.error('Error deleting attachment:', error);
          this.snackBar.open('Error deleting attachment', 'Close', { duration: 3000 });
        },
      });
    }
  }
}
