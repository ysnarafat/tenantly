export interface Organization {
  id: number;
  name: string;
  slug: string;
  subscriptionTier: 'basic' | 'professional' | 'enterprise';
  maxUsers: number;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
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
