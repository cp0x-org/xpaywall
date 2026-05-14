export interface PaymentChannel {
  id: string;
  protocol: string;
  method: string;
  scheme: string;
  enabled: boolean;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}
