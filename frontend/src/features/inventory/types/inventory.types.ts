export type ViewMode = 'products' | 'items';

export type ItemInputMethodType = 'barcode' | 'manual';

export interface Product {
  id: string;
  name: string;
  sku: string;
  sellingPrice: number;
  costPrice?: number;
  stock: number;
  current_quantity?: number;
  condition: string;
  category?: string;
  category_id?: string;
  price?: number;
  barcode?: string;
  image_url?: string;
  min_stock_level?: number;
  supplier_id?: string;
  supplier_name?: string;
}

export interface InventoryItem {
  id: string;
  product_id?: string;
  product_name?: string;
  product?: {
    name: string;
    sku?: string;
    barcode?: string;
  };
  condition: string;
  serial_number?: string;
  item_code?: string;
  selling_price: number;
  purchase_cost?: number;
  barcode?: string | null;
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
  supplier_id?: string;
  category_name?: string;
  purchase_date?: string;
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
