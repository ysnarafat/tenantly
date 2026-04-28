import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export enum AttachmentType {
  CONTRACT = 'CONTRACT',
  RECEIPT = 'RECEIPT',
  INVOICE = 'INVOICE',
  PHOTO = 'PHOTO',
  DOCUMENT = 'DOCUMENT',
  OTHER = 'OTHER',
}

export enum AttachmentStatus {
  ACTIVE = 'ACTIVE',
  ARCHIVED = 'ARCHIVED',
  DELETED = 'DELETED',
}

export interface Attachment {
  id: number;
  file_name: string;
  file_path: string;
  content_type: string;
  file_size: number;
  attachment_type: AttachmentType;
  status: AttachmentStatus;
  description?: string;
  tags?: string;
  payment_id?: number;
  property_id?: number;
  shop_id?: number;
  lease_id?: number;
  tenant_id?: number;
  uploaded_by?: number;
  created_at: string;
  updated_at: string;
}

export interface CreateAttachmentRequest {
  file: File;
  attachment_type: AttachmentType;
  description?: string;
  tags?: string;
  payment_id?: number;
  property_id?: number;
  shop_id?: number;
  lease_id?: number;
  tenant_id?: number;
}

export interface UpdateAttachmentRequest {
  attachment_type?: AttachmentType;
  description?: string;
  tags?: string;
  status?: AttachmentStatus;
}

@Injectable({
  providedIn: 'root',
})
export class AttachmentService {
  private http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/attachments`;

  // DEMO MODE: Set to true to enable demo data (disable for production)
  private readonly DEMO_MODE = true;

  private demoAttachments: Attachment[] = [
    {
      id: 1,
      file_name: 'lease_contract_shop_001.pdf',
      file_path: '/uploads/contracts/lease_contract_shop_001.pdf',
      content_type: 'application/pdf',
      file_size: 2048576,
      attachment_type: AttachmentType.CONTRACT,
      status: AttachmentStatus.ACTIVE,
      description: 'Lease contract for Electronics Store',
      tags: 'contract,lease,shop-001',
      lease_id: 1,
      shop_id: 1,
      uploaded_by: 1,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 2,
      file_name: 'rent_receipt_january_2024.pdf',
      file_path: '/uploads/receipts/rent_receipt_january_2024.pdf',
      content_type: 'application/pdf',
      file_size: 512000,
      attachment_type: AttachmentType.RECEIPT,
      status: AttachmentStatus.ACTIVE,
      description: 'Rent receipt for January 2024',
      tags: 'receipt,rent,january-2024',
      payment_id: 1,
      shop_id: 1,
      uploaded_by: 1,
      created_at: '2024-01-15T00:00:00Z',
      updated_at: '2024-01-15T00:00:00Z',
    },
    {
      id: 3,
      file_name: 'shop_photo_exterior.jpg',
      file_path: '/uploads/photos/shop_photo_exterior.jpg',
      content_type: 'image/jpeg',
      file_size: 1024000,
      attachment_type: AttachmentType.PHOTO,
      status: AttachmentStatus.ACTIVE,
      description: 'Exterior photo of Fashion Boutique',
      tags: 'photo,exterior,shop-002',
      shop_id: 2,
      uploaded_by: 1,
      created_at: '2023-06-01T00:00:00Z',
      updated_at: '2023-06-01T00:00:00Z',
    },
  ];

  getAllAttachments(): Observable<Attachment[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        observer.next(this.demoAttachments);
        observer.complete();
      });
    }

    return this.http.get<Attachment[]>(this.apiUrl);
  }

  getAttachmentById(id: number): Observable<Attachment> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const attachment = this.demoAttachments.find((a) => a.id === id);
        if (attachment) {
          observer.next(attachment);
        } else {
          observer.error({ error: 'Attachment not found' });
        }
        observer.complete();
      });
    }

    return this.http.get<Attachment>(`${this.apiUrl}/${id}`);
  }

  getAttachmentsByEntity(entityType: string, entityId: number): Observable<Attachment[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        let filteredAttachments: Attachment[] = [];

        switch (entityType) {
          case 'lease':
            filteredAttachments = this.demoAttachments.filter((a) => a.lease_id === entityId);
            break;
          case 'shop':
            filteredAttachments = this.demoAttachments.filter((a) => a.shop_id === entityId);
            break;
          case 'tenant':
            filteredAttachments = this.demoAttachments.filter((a) => a.tenant_id === entityId);
            break;
          case 'payment':
            filteredAttachments = this.demoAttachments.filter((a) => a.payment_id === entityId);
            break;
          case 'property':
            filteredAttachments = this.demoAttachments.filter((a) => a.property_id === entityId);
            break;
        }

        observer.next(filteredAttachments);
        observer.complete();
      });
    }

    return this.http.get<Attachment[]>(`${this.apiUrl}?${entityType}_id=${entityId}`);
  }

  uploadAttachment(request: CreateAttachmentRequest): Observable<Attachment> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const newAttachment: Attachment = {
          id: Math.max(...this.demoAttachments.map((a) => a.id)) + 1,
          file_name: request.file.name,
          file_path: `/uploads/${request.attachment_type.toLowerCase()}/${request.file.name}`,
          content_type: request.file.type,
          file_size: request.file.size,
          attachment_type: request.attachment_type,
          status: AttachmentStatus.ACTIVE,
          description: request.description,
          tags: request.tags,
          payment_id: request.payment_id,
          property_id: request.property_id,
          shop_id: request.shop_id,
          lease_id: request.lease_id,
          tenant_id: request.tenant_id,
          uploaded_by: 1, // Demo user ID
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };

        this.demoAttachments.push(newAttachment);
        observer.next(newAttachment);
        observer.complete();
      });
    }

    const formData = new FormData();
    formData.append('file', request.file);
    formData.append('attachment_type', request.attachment_type);
    if (request.description) formData.append('description', request.description);
    if (request.tags) formData.append('tags', request.tags);
    if (request.payment_id) formData.append('payment_id', request.payment_id.toString());
    if (request.property_id) formData.append('property_id', request.property_id.toString());
    if (request.shop_id) formData.append('shop_id', request.shop_id.toString());
    if (request.lease_id) formData.append('lease_id', request.lease_id.toString());
    if (request.tenant_id) formData.append('tenant_id', request.tenant_id.toString());

    return this.http.post<Attachment>(this.apiUrl, formData);
  }

  updateAttachment(id: number, request: UpdateAttachmentRequest): Observable<{ message: string }> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const attachmentIndex = this.demoAttachments.findIndex((a) => a.id === id);
        if (attachmentIndex !== -1) {
          this.demoAttachments[attachmentIndex] = {
            ...this.demoAttachments[attachmentIndex],
            ...request,
            updated_at: new Date().toISOString(),
          };
          observer.next({ message: 'Attachment updated successfully' });
        } else {
          observer.error({ error: 'Attachment not found' });
        }
        observer.complete();
      });
    }

    return this.http.put<{ message: string }>(`${this.apiUrl}/${id}`, request);
  }

  deleteAttachment(id: number): Observable<{ message: string }> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const attachmentIndex = this.demoAttachments.findIndex((a) => a.id === id);
        if (attachmentIndex !== -1) {
          this.demoAttachments[attachmentIndex].status = AttachmentStatus.DELETED;
          observer.next({ message: 'Attachment deleted successfully' });
        } else {
          observer.error({ error: 'Attachment not found' });
        }
        observer.complete();
      });
    }

    return this.http.delete<{ message: string }>(`${this.apiUrl}/${id}`);
  }

  downloadAttachment(id: number): Observable<Blob> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        // Create a dummy blob for demo
        const blob = new Blob(['Demo file content'], { type: 'application/pdf' });
        observer.next(blob);
        observer.complete();
      });
    }

    return this.http.get(`${this.apiUrl}/${id}/download`, { responseType: 'blob' });
  }

  getAttachmentsByType(type: AttachmentType): Observable<Attachment[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const filteredAttachments = this.demoAttachments.filter((a) => a.attachment_type === type);
        observer.next(filteredAttachments);
        observer.complete();
      });
    }

    return this.http.get<Attachment[]>(`${this.apiUrl}?type=${type}`);
  }

  searchAttachments(query: string): Observable<Attachment[]> {
    if (this.DEMO_MODE) {
      return new Observable((observer) => {
        const filteredAttachments = this.demoAttachments.filter(
          (a) =>
            a.file_name.toLowerCase().includes(query.toLowerCase()) ||
            a.description?.toLowerCase().includes(query.toLowerCase()) ||
            a.tags?.toLowerCase().includes(query.toLowerCase())
        );
        observer.next(filteredAttachments);
        observer.complete();
      });
    }

    return this.http.get<Attachment[]>(`${this.apiUrl}?search=${encodeURIComponent(query)}`);
  }
}
