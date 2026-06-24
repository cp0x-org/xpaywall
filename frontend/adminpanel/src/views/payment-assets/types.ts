export interface PaymentMethodAsset {
  id: string;
  payment_method_id: string;
  payment_method_name: string;
  payment_method_chain?: string | null;
  symbol: string;
  contract_address?: string | null;
  decimals: number;
  is_global: boolean;
  owner_user_id?: string | null;
  created_at: string;
  updated_at: string;
}
