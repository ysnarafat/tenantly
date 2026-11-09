import { UserRole } from './role.model';

export interface UserInvitation {
  id: number;
  organizationId: number;
  organizationName?: string;
  email: string;
  role: UserRole;
  token: string;
  expiresAt: Date;
  acceptedAt?: Date;
  createdAt: Date;
  email_delivery_status?: 'unknown' | 'pending' | 'sent' | 'failed';
}

export interface InviteUserRequest {
  email: string;
  firstName: string;
  lastName: string;
  role: UserRole;
}
