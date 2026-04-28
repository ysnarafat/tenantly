export interface Organization {
  id: number;
  name: string;
  slug: string;
  subscriptionTier: 'basic' | 'professional' | 'enterprise';
  maxUsers: number;
  active: boolean;
  description?: string;
  createdAt?: Date;
  updatedAt?: Date;
  created_at?: string;
  updated_at?: string;
}

export interface CreateOrgRequest {
  name: string;
  slug: string;
  subscriptionTier: 'basic' | 'professional' | 'enterprise';
  maxUsers: number;
}

export interface UpdateOrgRequest {
  name?: string;
  subscriptionTier?: 'basic' | 'professional' | 'enterprise';
  maxUsers?: number;
  active?: boolean;
}

export interface OrgStats {
  totalUsers: number;
  usersByRole: Record<string, number>;
  pendingInvitations: number;
  activeProperties: number;
}

export interface UserOrganization {
  id: number;
  organization_id: number;
  organization: Organization;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface LoginResponseWithOrganizations {
  token: string;
  refresh_token: string;
  user: unknown;
  organizations: UserOrganization[];
  default_organization_id: number;
  expires_at: string | Date;
}

export interface SetOrganizationRequest {
  organization_id: number;
}

export interface SetOrganizationResponse {
  token: string;
  refresh_token: string;
  organization: UserOrganization;
  expires_at: string | Date;
}
