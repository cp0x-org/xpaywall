export interface PaymentMethod {
  id: string;
  code: string;
  protocol: string;
  name: string;
  caip2_chain_id?: string | null;
  // MPP-only: method (e.g. "tempo") and scheme (e.g. "charge"). x402 leaves these unset.
  method?: string | null;
  scheme?: string | null;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}
