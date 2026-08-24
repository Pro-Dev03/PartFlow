export interface Purchase {
  id: string;
  supplier_id: string;
  supplier?: {
    id: string;
    name: string;
  };
  invoice_number: string;
  purchase_date: string;
  expected_date?: string;
  status: 'pending' | 'ordered' | 'received' | 'cancelled';
  items: PurchaseItem[];
  total_cost: number;
  paid_amount: number;
  remaining_amount: number;
}

export interface PurchaseItem {
  product_id: string;
  product_name: string;
  quantity: number;
  unit_cost: number;
  condition: 'new' | 'used';
}

export interface PurchaseFormData {
  supplier_id: string;
  invoice_number: string;
  purchase_date: string;
  items: PurchaseItem[];
}

export interface PurchaseStats {
  totalPurchases: number;
  pendingCount: number;
  receivedCount: number;
  totalCost: number;
}