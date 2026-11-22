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

  getAllAttachments(): Observable<Attachment[]> {
    return this.http.get<Attachment[]>(this.apiUrl);
  }

  getAttachmentById(id: number): Observable<Attachment> {
    return this.http.get<Attachment>(`${this.apiUrl}/${id}`);
  }

  // Tenant attachment endpoints
  getAttachmentsByTenant(tenantId: number): Observable<Attachment[]> {
    return this.http.get<Attachment[]>(`${this.apiUrl}/tenants/${tenantId}/attachments`);
  }

  uploadTenantAttachment(file: File, tenantId: number): Observable<Attachment> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('tenant_id', tenantId.toString());

    return this.http.post<Attachment>(`${this.apiUrl}/tenants/${tenantId}/attachments`, formData);
  }

  deleteTenantAttachment(attachmentId: number, tenantId: number): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(
      `${this.apiUrl}/tenants/${tenantId}/attachments/${attachmentId}`
    );
  }

  downloadTenantAttachment(attachmentId: number, tenantId: number): Observable<Blob> {
    return this.http.get(`${this.apiUrl}/tenants/${tenantId}/attachments/${attachmentId}`, {
      responseType: 'blob',
    });
  }

  // Lease attachment endpoints
  getAttachmentsByLease(leaseId: number): Observable<Attachment[]> {
    return this.http.get<Attachment[]>(`${this.apiUrl}/leases/${leaseId}/attachments`);
  }

  uploadLeaseAttachment(file: File, leaseId: number): Observable<Attachment> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('lease_id', leaseId.toString());

    return this.http.post<Attachment>(`${this.apiUrl}/leases/${leaseId}/attachments`, formData);
  }

  deleteLeaseAttachment(attachmentId: number, leaseId: number): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(
      `${this.apiUrl}/leases/${leaseId}/attachments/${attachmentId}`
    );
  }

  downloadLeaseAttachment(attachmentId: number, leaseId: number): Observable<Blob> {
    return this.http.get(`${this.apiUrl}/leases/${leaseId}/attachments/${attachmentId}`, {
      responseType: 'blob',
    });
  }

  // Generic method for attachment lists that routes to specific endpoints
  getAttachmentsByEntity(entityType: string, entityId: number): Observable<Attachment[]> {
    if (entityType === 'tenant') {
      return this.getAttachmentsByTenant(entityId);
    } else if (entityType === 'lease') {
      return this.getAttachmentsByLease(entityId);
    } else {
      // For other entity types, fallback to all attachments
      return this.getAllAttachments();
    }
  }

  // Keep existing generic attachment methods for backwards compatibility
  // and payment/document entity types
  uploadAttachment(request: CreateAttachmentRequest): Observable<Attachment> {
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
    return this.http.put<{ message: string }>(`${this.apiUrl}/${id}`, request);
  }

  deleteAttachment(id: number): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.apiUrl}/${id}`);
  }

  downloadAttachment(id: number): Observable<Blob> {
    return this.http.get(`${this.apiUrl}/${id}/download`, { responseType: 'blob' });
  }

  getAttachmentsByType(type: AttachmentType): Observable<Attachment[]> {
    return this.http.get<Attachment[]>(`${this.apiUrl}?type=${type}`);
  }

  searchAttachments(query: string): Observable<Attachment[]> {
    return this.http.get<Attachment[]>(`${this.apiUrl}?search=${encodeURIComponent(query)}`);
  }
}
