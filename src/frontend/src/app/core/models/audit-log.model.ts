export interface AuditLog {
  id: number;
  organizationId: number;
  userId: number;
  userEmail: string;
  action: string;
  resourceType: string;
  resourceId?: number;
  resourceName?: string;
  ipAddress: string;
  status: 'success' | 'failure';
  changes?: Record<string, unknown>;
  createdAt: Date;
}

export interface AuditFilter {
  action?: string;
  resourceType?: string;
  userId?: number;
  startDate?: Date;
  endDate?: Date;
}
