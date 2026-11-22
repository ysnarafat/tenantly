import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';

interface DocumentAttachment {
  id: number;
  file_name: string;
  content_type: string;
  file_size: number;
  uploaded_at: string;
}

interface FileSelectedEvent {
  file: File;
  selected: boolean;
}

interface FilesDeletedEvent {
  deletedIds: number[];
}

interface UploadProgressEvent {
  progress: number;
  status: string;
  file?: File;
}

interface AttachmentsChangeEvent {
  value: DocumentAttachment[];
}

@Component({
  selector: 'document-attachments',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonToggleModule,
    MatProgressSpinnerModule,
    MatProgressBarModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
  ],
  templateUrl: './document-attachments.component.html',
  styleUrl: './document-attachments.component.scss',
})
export class DocumentAttachmentsComponent {
  @Input() attachments: DocumentAttachment[] = [];
  @Input() isUploading = false;
  @Input() maxAttachments = 10;
  @Input() errorMessage = '';
  @Input() maxFileSize = 10 * 1024 * 1024; // 10MB

  @Output() uploadProgress = new EventEmitter<UploadProgressEvent>();
  @Output() fileSelected = new EventEmitter<FileSelectedEvent>();
  @Output() filesDeleted = new EventEmitter<FilesDeletedEvent>();
  @Output() attachmentsChange = new EventEmitter<AttachmentsChangeEvent>();

  readonly ALLOWED_CONTENT_TYPES = [
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp',
    'application/pdf',
  ];

  get MAX_ATTACHMENT_COUNT(): number {
    return this.maxAttachments;
  }

  uploadFile(file: File): Promise<void> {
    if (file.size > this.maxFileSize) {
      throw new Error('File size exceeds maximum limit of 10MB');
    }

    if (!this.ALLOWED_CONTENT_TYPES.includes(file.type)) {
      const allowedTypes = this.ALLOWED_CONTENT_TYPES.join(', ');
      throw new Error(`Only the following file types are allowed: ${allowedTypes}`);
    }

    // Simulate upload - in real implementation, send to backend
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        if (Math.random() > 0.9) {
          reject(new Error('Upload failed'));
          return;
        }
        resolve();
      }, 1000);
    });
  }

  handleFileUpload(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) {
      return;
    }

    const files = Array.from(input.files);

    if (this.attachments.length + files.length > this.maxAttachments) {
      this.uploadProgress.emit({
        progress: 0,
        status: `Maximum ${this.maxAttachments} attachments allowed per entity`,
      });
      return;
    }

    this.isUploading = true;
    this.uploadProgress.emit({ progress: 0, status: 'Starting upload...' });

    const uploadPromises = files.map((file, index) =>
      this.uploadFile(file).then(() => {
        this.uploadProgress.emit({
          progress: ((index + 1) / files.length) * 100,
          status: `Uploaded ${index + 1} of ${files.length}`,
        });
      })
    );

    Promise.all(uploadPromises)
      .then(() => {
        const newAttachments: DocumentAttachment[] = files.map((file, index) => ({
          id: Date.now() + index,
          file_name: file.name,
          content_type: file.type,
          file_size: file.size,
          uploaded_at: new Date().toISOString(),
        }));

        this.attachments.push(...newAttachments);
        input.value = '';
        this.uploadProgress.emit({ progress: 100, status: 'Upload complete' });
      })
      .catch((err) => {
        console.error('Upload error:', err);
        this.uploadProgress.emit({
          progress: 0,
          status: err.message || 'Failed to upload attachment',
        });
      })
      .finally(() => {
        this.isUploading = false;
        this.attachmentsChange.emit({ value: [...this.attachments] });
      });

    // Emit file selected event
    files.forEach((file) => {
      this.fileSelected.emit({ file, selected: true });
    });
  }

  onDragOver(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
  }

  onDragLeave(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
  }

  removeAttachment(attachmentId: number): void {
    this.attachments = this.attachments.filter((a) => a.id !== attachmentId);
    this.filesDeleted.emit({ deletedIds: [attachmentId] });
    this.attachmentsChange.emit({ value: [...this.attachments] });
  }

  formatFileSize(bytes: number): string {
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return parseFloat((bytes / Math.pow(1024, i)).toFixed(2)) + ' ' + sizes[i];
  }

  formatContentType(contentType: string): string {
    if (contentType.startsWith('image/')) {
      const imageNames: Record<string, string> = {
        'image/jpeg': 'JPEG Image',
        'image/png': 'PNG Image',
        'image/gif': 'GIF Image',
        'image/webp': 'WebP Image',
      };
      return imageNames[contentType] || contentType;
    } else if (contentType === 'application/pdf') {
      return 'PDF Document';
    }
    return contentType;
  }

  getIconForAttachmentType(contentType: string): string {
    if (contentType.startsWith('image/')) {
      return 'image';
    } else if (contentType === 'application/pdf') {
      return 'picture_as_pdf';
    }
    return 'attach_file';
  }

  getPreviewUrl(file: File): string | null {
    if (file.type.startsWith('image/')) {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      return URL.createObjectURL(file);
    }
    return null;
  }
}
