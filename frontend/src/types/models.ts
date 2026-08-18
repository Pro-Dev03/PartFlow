// Organization & User Types
export interface Organization {
  id: string;
  name: string;
  store_name: string;
  currency: string;
  timezone: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  organization_id: string;
  name: string;
  email: string;
  phone?: string;
  role_id: string;
  created_at: string;
  updated_at: string;
}

export interface Role {
  id: string;
  name: string;
  description?: string;
  permissions: Permission[];
}

export interface Permission {
  id: string;
  name: string;
  description?: string;
}

// Product Types
export interface Product {
  id: string;
  organization_id: string;
  name: string;
  brand_id?: string;
  category_id?: string;
  model?: string;
  sku?: string;
  description?: string;
  product_type: 'QUANTITY' | 'INDIVIDUAL';
  default_cost?: number;
  default_price?: number;
  minimum_stock?: number;
  warranty_policy?: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface Brand {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

// Inventory Types
export interface InventoryItem {
  id: string;
  organization_id: string;
  product_id: string;
  item_code: string;
  barcode?: string;
  serial_number?: string;
  condition: 'NEW' | 'USED' | 'REFURBISHED' | 'DAMAGED' | 'FOR_PARTS';
  grade?: 'EXCELLENT' | 'VERY_GOOD' | 'GOOD' | 'FAIR' | 'POOR';
  purchase_cost?: number;
  selling_price?: number;
  status: 'PURCHASED' | 'RECEIVED' | 'INSPECTION' | 'AVAILABLE' | 'RESERVED' | 'SOLD' | 'DAMAGED' | 'IN_REPAIR' | 'RETURNED' | 'WARRANTY' | 'FOR_PARTS' | 'ARCHIVED';
  location_id?: string;
  supplier_id?: string;
  purchase_date?: string;
  sold_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Location {
  id: string;
  organization_id: string;
  name: string;
  parent_id?: string;
  type: 'WAREHOUSE' | 'SHELF' | 'BOX' | 'DISPLAY';
  created_at: string;
  updated_at: string;
}

export interface InventoryMovement {
  id: string;
  organization_id: string;
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
  organization_id: string;
  name: string;
  phone?: string;
  email?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CustomerLedger {
  id: string;
  organization_id: string;
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
  organization_id: string;
  customer_id?: string;
  subtotal: number;
  discount: number;
  tax: number;
  total: number;
  paid_amount: number;
  debt_amount: number;
  status: 'PENDING' | 'COMPLETED' | 'CANCELLED' | 'REFUNDED';
  created_by: string;
  created_at: string;
}

export interface SaleItem {
  id: string;
  sale_id: string;
  product_id: string;
  inventory_item_id?: string;
  quantity: number;
  unit_price: number;
  unit_cost: number;
  discount: number;
  total: number;
}

export interface Payment {
  id: string;
  organization_id: string;
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
  organization_id: string;
  name: string;
  phone?: string;
  email?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface SupplierLedger {
  id: string;
  organization_id: string;
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
  organization_id: string;
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
  organization_id: string;
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
  organization_id: string;
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
  organization_id: string;
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
  organization_id: string;
  user_id?: string;
  type: 'LOW_STOCK' | 'OVERDUE_DEBT' | 'WARRANTY_EXPIRING' | 'INSPECTION_REQUIRED' | 'RESERVATION_EXPIRING' | 'PAYMENT_RECEIVED' | 'PURCHASE_RECEIVED';
  title: string;
  message: string;
  read: boolean;
  created_at: string;
}

export interface AuditLog {
  id: string;
  organization_id: string;
  user_id: string;
  action: string;
  entity_type: string;
  entity_id: string;
  changes?: Record<string, any>;
  created_at: string;
}