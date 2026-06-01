export interface RouteRow {
  id: string;
  project_id: string;
  project_name: string;
  project_slug: string;
  name: string;
  path_pattern: string;
  price_usd: string;
  description: string;
  free: boolean;
  bazaar?: Record<string, unknown> | null;
  proxy_url: string;
  target_url: string;
  created_at: string;
  updated_at: string;
}
