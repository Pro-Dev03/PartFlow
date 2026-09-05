export interface Purchase {
  id: string;
  supplier_id: string;
  supplier?: {
    id: string;
    name: string;
  };
  invoice_number: string;
  purchase_date: string;
  expected_delivery_date?: string;
  status: 'draft' | 'pending' | 'ordered' | 'received' | 'cancelled' | 'reversed' | 'partially_received';
  items: PurchaseItem[];
  subtotal?: number;
  tax_amount?: number;
  total_amount: number;
  paid_amount: number;
  remaining: number;
  total_items?: number;
  supplier_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface PurchaseItem {
  product_id: string;
  product_name: string;
  quantity: number;
  unit_cost: number;
  condition: 'new' | 'used' | 'refurbished';
}

export interface PurchaseFormData {
  supplier_id: string;
  invoice_number: string;
  purchase_date: string;
  expected_delivery_date?: string;
  notes?: string;
  items: PurchaseItem[];
}

export interface PurchaseStats {
  totalPurchases: number;
  pendingCount: number;
  receivedCount: number;
  reversedCount: number;
  pendingCost: number;
  receivedCost: number;
  outstandingAmount: number;
  untaxedCount: number;
  untaxedCost: number;
  taxedCount: number;
  taxedCost: number;
}