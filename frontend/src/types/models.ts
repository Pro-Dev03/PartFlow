// User Types
export interface User {
  id: string;
  name?: string;
  first_name?: string;
  last_name?: string;
  email: string;
  phone?: string;
  subscription_status?: string;
  subscription_expires_at?: string | null;
  created_at: string;
  updated_at: string;
}

// Product Types
export interface Product {
  id: string;
  name: string;
  brand_id?: string;
  category_id?: string;
  model?: string;
  sku: string;
  barcode: string;
  description?: string;
  track_serial: boolean;
  track_individual: boolean;
  min_stock_level: number;
  warranty_days: number;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface Brand {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

// Inventory Types
export interface InventoryItem {
  id: string;
  product_id?: string;
  part_type_id?: string;
  item_code: string;
  barcode: string;
  serial_number?: string;
  condition: 'NEW' | 'USED' | 'REFURBISHED' | 'DAMAGED' | 'FOR_PARTS';
  grade?: 'EXCELLENT' | 'VERY_GOOD' | 'GOOD' | 'FAIR' | 'POOR';
  purchase_cost: number;
  selling_price: number;
  status: 'PURCHASED' | 'RECEIVED' | 'INSPECTION' | 'AVAILABLE' | 'RESERVED' | 'SOLD' | 'DAMAGED' | 'IN_REPAIR' | 'RETURNED' | 'FOR_PARTS' | 'ARCHIVED';
  location_id?: string;
  supplier_id?: string;
  purchase_date?: string;
  sold_at?: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface Location {
  id: string;
  name: string;
  parent_id?: string;
  type: 'WAREHOUSE' | 'SHELF' | 'BOX' | 'DISPLAY';
  created_at: string;
  updated_at: string;
}

export interface PartType {
  id: string;
  name: string;
  name_en?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface InventoryMovement {
  id: string;
  item_id?: string;
  product_id?: string;
  movement_type: 'PURCHASE' | 'SALE' | 'RETURN' | 'ADJUSTMENT' | 'TRANSFER' | 'RESERVATION' | 'RELEASE' | 'DAMAGE' | 'REPAIR';
  quantity: number;
  before_quantity: number;
  after_quantity: number;
  reference_type?: string;
  reference_id?: string;
  reason?: string;
  created_by: string;
  created_at: string;
}

// Customer Types
export interface Customer {
  id: string;
  name: string;
  phone?: string;
  email?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CustomerLedger {
  id: string;
  customer_id: string;
  type: 'SALE' | 'PAYMENT' | 'RETURN' | 'ADJUSTMENT';
  amount: number;
  balance: number;
  reference_type?: string;
  reference_id?: string;
  description?: string;
  created_at: string;
}

// Sales Types
export interface Sale {
  id: string;
  invoice_number: string;
  customer_id?: string;
  user_id: string;
  sale_date: string;
  subtotal: number;
  tax_amount: number;
  discount_amount: number;
  total_amount: number;
  cost_amount: number;
  gross_profit: number;
  net_profit: number;
  paid_amount: number;
  payment_method?: string;
  payment_status: string;
  status: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface SaleItem {
  id: string;
  sale_id: string;
  product_id: string;
  inventory_item_id?: string;
  serial_number?: string;
  quantity: number;
  unit_price: number;
  unit_cost: number;
  discount_amount: number;
  tax_amount: number;
  total_amount: number;
  supplier_id?: string;
  created_at: string;
}

export interface Payment {
  id: string;
  reference_type: 'SALE' | 'PURCHASE' | 'CUSTOMER' | 'SUPPLIER';
  reference_id: string;
  amount: number;
  method: 'CASH' | 'CARD' | 'BANK_TRANSFER' | 'DEBT' | 'OTHER';
  status: 'PENDING' | 'COMPLETED' | 'CANCELLED';
  notes?: string;
  created_by: string;
  created_at: string;
}

// Supplier Types
export interface Supplier {
  id: string;
  name: string;
  phone?: string;
  email?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface SupplierLedger {
  id: string;
  supplier_id: string;
  type: 'PURCHASE' | 'PAYMENT' | 'RETURN' | 'ADJUSTMENT';
  amount: number;
  balance: number;
  reference_type?: string;
  reference_id?: string;
  description?: string;
  created_at: string;
}

// Purchase Types
export interface Purchase {
  id: string;
  supplier_id: string;
  subtotal: number;
  discount: number;
  tax: number;
  total: number;
  status: 'PENDING' | 'RECEIVED' | 'CANCELLED';
  created_by: string;
  created_at: string;
}

export interface PurchaseItem {
  id: string;
  purchase_id: string;
  product_id: string;
  quantity: number;
  unit_cost: number;
  total: number;
}

// Warranty Types
export interface Warranty {
  id: string;
  sale_id: string;
  item_id?: string;
  start_date: string;
  end_date: string;
  duration_months: number;
  type: 'STANDARD' | 'EXTENDED';
  status: 'ACTIVE' | 'EXPIRED' | 'CLAIMED' | 'CANCELLED';
  created_at: string;
}

export interface WarrantyClaim {
  id: string;
  warranty_id: string;
  customer_id?: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'COMPLETED';
  issue_description: string;
  resolution?: string;
  created_at: string;
}

// Other Types
export interface Expense {
  id: string;
  category_id?: string;
  amount: number;
  description: string;
  date: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED';
  created_by: string;
  created_at: string;
}

export interface Notification {
  id: string;
  user_id?: string;
  type: 'LOW_STOCK' | 'OVERDUE_DEBT' | 'WARRANTY_EXPIRING' | 'INSPECTION_REQUIRED' | 'RESERVATION_EXPIRING' | 'PAYMENT_RECEIVED' | 'PURCHASE_RECEIVED';
  title: string;
  message: string;
  read: boolean;
  created_at: string;
}

export interface AuditLog {
  id: string;
  user_id: string;
  action: string;
  entity_type: string;
  entity_id: string;
  changes?: Record<string, any>;
  created_at: string;
}