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

// MPP project links carry rpc_url/secret_key (and method) in a config JSONB
// instead of a facilitator. x402 links leave config unset.
export interface ProjectPaymentMethodConfig {
  method?: string;
  rpc_url?: string;
  secret_key?: string;
  [key: string]: unknown;
}

export interface ProjectPaymentMethod {
  id: string;
  project_id: string;
  payment_method_id: string;
  asset_id: string;
  scheme: string;
  // x402 only: settlement facilitator. Omitted/null for MPP.
  facilitator_id?: string | null;
  payout_address?: string | null;
  config?: ProjectPaymentMethodConfig | null;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}
