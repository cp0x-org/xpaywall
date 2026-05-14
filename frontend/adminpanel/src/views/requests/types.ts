export interface RequestLog {
  id: string;
  project_id: string;
  outbound_route_id?: string;
  request_id: string;
  method: string;
  path: string;
  client_ip?: string;
  user_agent?: string;
  status: string;
  payment_required: boolean;
  payment_requested_at?: string;
  payment_completed: boolean;
  payment_completed_at?: string;
  payment_channel_id?: string;
  payment_channel_asset_id?: string;
  amount_paid?: number;
  amount_usd?: string;
  upstream_url?: string;
  upstream_status_code?: number;
  upstream_response_time_ms?: number;
  final_status_code?: number;
  error_type?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}
