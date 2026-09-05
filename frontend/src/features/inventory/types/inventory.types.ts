export type ViewMode = 'products' | 'items';

export type ItemInputMethodType = 'barcode' | 'camera' | 'manual';

export interface Product {
  id: string;
  name: string;
  sku: string;
  sellingPrice: number;
  costPrice?: number;
  stock: number;
  condition: string;
  category?: string;
  category_id?: string;
  price?: number;
  barcode?: string;
  image_url?: string;
  min_stock_level?: number;
}

export interface InventoryItem {
  id: string;
  product_id?: string;
  product_name?: string;
  product?: {
    name: string;
  };
  condition: string;
  selling_price: number;
  purchase_cost?: number;
  price?: number;
  stock?: number;
  current_quantity?: number;
  available_quantity?: number;
  quantity?: number;
  location?: string;
  status: string;
  created_at: string;
  purchase_date?: string;
  supplier_name?: string;
  supplier_phone?: string;
}

export interface InventoryStats {
  totalItems: number;
  totalValue: string;
  lowStock: number;
}

export interface FilterConfig {
  key: string;
  value: string;
}

export interface SortConfig {
  key: string;
  direction: 'asc' | 'desc' | null;
}