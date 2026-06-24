export interface Facilitator {
  id: string;
  name: string;
  url: string;
  enabled: boolean;
  is_global: boolean;
  owner_user_id?: string | null;
  created_at: string;
  updated_at: string;
}
