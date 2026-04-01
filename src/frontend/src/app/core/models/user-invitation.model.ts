import { UserRole } from './role.model';

export interface UserInvitation {
  id: number;
  organizationId: number;
  email: string;
  role: UserRole;
  token: string;
  expiresAt: Date;
  acceptedAt?: Date;
  createdAt: Date;
}

export interface InviteUserRequest {
  email: string;
  firstName: string;
  lastName: string;
  role: UserRole;
}
