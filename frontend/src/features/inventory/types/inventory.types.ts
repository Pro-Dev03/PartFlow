export type ViewMode = 'products' | 'items';

export type ItemInputMethodType = 'barcode' | 'camera' | 'manual';

export interface Product {
  id: string;
  name: string;
  sku: string;
  sellingPrice: number;
  stock: number;
  condition: string;
  category?: string;
  category_id?: string;
  price?: number;
  barcode?: string;
}

export interface InventoryItem {
  id: string;
  product_name?: string;
  product?: {
    name: string;
  };
  condition: string;
  selling_price: number;
  price?: number;
  stock?: number;
  location?: string;
  status: string;
  created_at: string;
}

export interface InventoryStats {
  totalItems: number;
  totalValue: string;
  lowStock: number;
  usedItems: number;
}

export interface FilterConfig {
  key: string;
  value: string;
}

export interface SortConfig {
  key: string;
  direction: 'asc' | 'desc' | null;
}