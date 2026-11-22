export interface PublicPlan {
  id: number;
  code: 'basic' | 'professional' | 'enterprise';
  name_en: string;
  name_bn: string;
  description_en: string;
  description_bn: string;
  price_minor: number | null;
  currency: string;
  billing_period: 'monthly' | 'yearly';
  max_users: number;
  max_properties: number | null;
  features_en: string[];
  features_bn: string[];
}
