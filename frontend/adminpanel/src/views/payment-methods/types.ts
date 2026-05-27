export interface PaymentMethod {
  id: string;
  code: string;
  protocol: string;
  name: string;
  caip2_chain_id?: string | null;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}
