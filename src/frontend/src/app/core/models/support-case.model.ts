export type SupportCaseStatus = 'open' | 'in_progress' | 'waiting' | 'resolved' | 'closed';

export interface SupportCaseMessage {
  id: number;
  case_id: number;
  author: string;
  body: string;
  created_at: string;
}

export interface SupportCase {
  id: number;
  subject: string;
  description: string;
  status: SupportCaseStatus;
  priority: string;
  created_at: string;
  updated_at: string;
  messages?: SupportCaseMessage[];
}

export interface CreateSupportCaseRequest {
  subject: string;
  description: string;
}
