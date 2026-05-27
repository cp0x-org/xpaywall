export interface Project {
  id: string;
  owner_user_id: string;
  owner_username: string;
  name: string;
  slug: string;
  enabled: boolean;
  base_url: string;
  payment_methods: string[];
  created_at: string;
  updated_at: string;
}

export interface FullProject extends Project {
  target_id?: string;
  server_name: string;
  base_url: string;
  auth_header_name?: string | null;
  auth_header_value?: string | null;
  is_default: boolean;
  allow_unmatched: boolean;
}

export interface ProxyUrl {
  proxy_url?: string;
}

export interface ProjectPaymentMethod {
  id: string;
  project_id: string;
  payment_method_id: string;
  asset_id: string;
  scheme: string;
  facilitator_id: string;
  payout_address?: string | null;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}
