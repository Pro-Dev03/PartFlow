// API Type Definitions for PartFlow
// This file provides type safety for API requests and responses

// Generic API Response
export interface ApiResponse<T> {
  data: T;
  message?: string;
  error?: string;
}

// Pagination
export interface PaginationParams {
  page?: number;
  per_page?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

// Product Types
export interface Product {
  id: string;
  name: string;
  name_ar?: string;
  name_en?: string;
  sku?: string;
  barcode?: string;
  image_url?: string;
  description?: string;
  category_id?: string;
  preferred_supplier_id?: string;
  cost_price?: number;
  selling_price?: number;
  stock?: number;
  condition?: 'new' | 'used';
  min_stock_level?: number;
  max_stock?: number;
  sales_count?: number;
  created_at?: string;
  updated_at?: string;
}

export interface ProductCreateRequest {
  name: string;
  name_ar?: string;
  name_en?: string;
  sku?: string;
  barcode?: string;
  image_url?: string;
  description?: string;
  category_id?: string;
  cost_price?: number;
  selling_price?: number;
  stock?: number;
  condition?: 'new' | 'used';
  min_stock_level?: number;
  max_stock?: number;
}

export interface ProductUpdateRequest extends Partial<ProductCreateRequest> {}

export interface BulkProductCreateRequest {
  items: ProductCreateRequest[];
}

export interface ProductListParams extends PaginationParams {
  search?: string;
  category_id?: string;
  in_stock_only?: boolean;
}

// Category Types
export interface Category {
  id: string;
  name: string;
  name_ar?: string;
  name_en?: string;
  description?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CategoryCreateRequest {
  name: string;
  name_ar?: string;
  name_en?: string;
  description?: string;
}

export interface CategoryUpdateRequest extends Partial<CategoryCreateRequest> {}

// Customer Types
export interface Customer {
  id: string;
  name: string;
  phone?: string;
  email?: string;
  address?: string;
  notes?: string;
  debt_reason?: string;
  balance?: number;
  credit_limit?: number;
  created_at?: string;
  updated_at?: string;
}

export interface CustomerCreateRequest {
  name: string;
  phone?: string;
  email?: string;
  address?: string;
  notes?: string;
  debt_reason?: string;
  credit_limit?: number;
}

export interface CustomerUpdateRequest extends Partial<CustomerCreateRequest> {}

export interface CustomerListParams extends PaginationParams {
  search?: string;
  sort_by?: 'name' | 'total_purchases';
  sort_order?: 'asc' | 'desc';
}

// Supplier Types
export interface Supplier {
  id: string;
  name: string;
  phone?: string;
  email?: string;
  address?: string;
  balance?: number;
  created_at?: string;
  updated_at?: string;
}

export interface SupplierCreateRequest {
  name: string;
  phone?: string;
  email?: string;
  address?: string;
}

export interface SupplierUpdateRequest extends Partial<SupplierCreateRequest> {}

// Inventory Types
export interface InventoryItem {
  id: string;
  product_id: string;
  product_name?: string;
  product?: Product;
  serial_number?: string;
  condition: 'new' | 'used';
  grade?: string;
  purchase_cost: number;
  selling_price: number;
  supplier_id?: string;
  supplier_name?: string;
  purchase_date?: string;
  part_type_id?: string;
  status?: 'available' | 'sold' | 'reserved' | 'damaged';
  created_at?: string;
  updated_at?: string;
}

export interface InventoryCreateRequest {
  product_id: string;
  quantity?: number;
  serial_number?: string;
  condition: 'new' | 'used';
  grade?: string;
  purchase_cost: number;
  selling_price: number;
  supplier_id?: string;
  part_type_id?: string;
  status?: 'PURCHASED' | 'RECEIVED' | 'AVAILABLE';
  notes?: string;
}

export interface OpeningStockCreateRequest {
  product_id: string;
  mode: 'quantity' | 'individual';
  quantity: number;
  business_date: string;
  item_code?: string;
  barcode?: string;
  serial_number?: string;
  part_type_id?: string;
  condition: 'NEW' | 'USED' | 'REFURBISHED' | 'DAMAGED' | 'FOR_PARTS';
  grade?: string;
  purchase_cost: number;
  selling_price: number;
  supplier_id?: string;
  customer_id?: string;
  notes?: string;
}

export interface InventoryUpdateRequest extends Partial<InventoryCreateRequest> {
  barcode?: string;
}

export interface InventoryListParams extends PaginationParams {
  search?: string;
  condition?: string;
  supplier_id?: string;
  product_id?: string;
  status?: string;
}

// Sales Types
export interface Sale {
  id: string;
  customer_id?: string;
  customer_name?: string;
  total_amount: number;
  paid_amount: number;
  cash_received?: number;
  change_amount?: number;
  payment_method: 'cash' | 'card' | 'transfer' | 'debt';
  status: 'completed' | 'pending' | 'cancelled';
  created_at?: string;
  updated_at?: string;
}

export interface SaleItem {
  product_id: string;
  barcode?: string;
  inventory_item_id?: string;
  quantity: number;
  unit_price: number;
  is_trade_in?: boolean;
  purchase_cost?: number;
}

export interface SaleCreateRequest {
  customer_id?: string;
  items: SaleItem[];
  payment_method: 'cash' | 'card' | 'transfer' | 'debt' | 'installment';
  payment_amount: number;
  installment_months?: number;
  cash_received?: number;
  total_amount: number;
  tax_exempt?: boolean;
  discount_type?: 'percentage' | 'fixed';
  discount_value?: number;
  payment_allocations?: PaymentAllocationRequest[];
  payment_transaction_id?: string;
}

export interface PaymentAllocationRequest {
  amount: number;
  method: 'cash' | 'card' | 'transfer' | 'debt' | 'checks';
  check_number?: string;
  bank_name?: string;
  check_date?: string;
}

export interface PaymentTransactionCreateRequest {
  order_id: string;
  sale_id?: string;
  amount_minor: number;
  currency?: string;
  description?: string;
  idempotency_key: string;
  success_url?: string;
  failure_url?: string;
  cancel_url?: string;
}

export interface PaymentTransaction {
  id: string;
  order_id?: string;
  sale_id?: string;
  provider: string;
  provider_payment_id?: string;
  status: 'pending' | 'processing' | 'paid' | 'failed' | 'cancelled' | 'refunded' | 'partially_refunded';
  amount_minor: number;
  currency: string;
  checkout_url?: string;
}

// Debt Types
export interface Debt {
  id: string;
  customer_id: string;
  customer_name?: string;
  amount: number;
  paid_amount: number;
  remaining_amount: number;
  due_date?: string;
  status: 'pending' | 'partial' | 'paid' | 'overdue';
  created_at?: string;
  updated_at?: string;
}

export interface DebtPayment {
  debt_id: string;
  amount: number;
  payment_method: 'cash' | 'card' | 'transfer';
  notes?: string;
}

export interface CustomerDebtPaymentRequest {
  amount: number;
  method: 'cash' | 'credit' | 'bank_transfer' | 'check';
  reference?: string;
}

export interface PaymentCreateRequest {
  type: 'customer' | 'supplier' | 'expense';
  reference_id: string;
  amount: number;
  method: 'cash' | 'card' | 'transfer';
  notes?: string;
}

// Purchase Types
export interface Purchase {
  id: string;
  supplier_id: string;
  supplier_name?: string;
  total_amount: number;
  status: 'pending' | 'completed' | 'cancelled';
  created_at?: string;
  updated_at?: string;
}

export interface PurchaseItem {
  product_id: string;
  quantity: number;
  cost_price: number;
}

export interface PurchaseCreateRequest {
  supplier_id: string;
  items: PurchaseItem[];
  total_amount: number;
  invoice_number?: string;
  purchase_date?: string;
  expected_delivery_date?: string;
  notes?: string;
}

// Expense Types
export interface Expense {
  id: string;
  description: string;
  amount: number;
  category?: string;
  date?: string;
  created_at?: string;
}

export interface ExpenseCreateRequest {
  category_id: string;
  title: string;
  description?: string;
  amount: number;
  currency: string;
  expense_date: string;
  payment_method: 'cash' | 'card' | 'bank_transfer' | 'check';
  reference?: string;
  receipt_url?: string;
  is_recurring?: boolean;
  recurring_period?: 'daily' | 'weekly' | 'monthly' | 'yearly';
}

export interface ExpenseCategory {
  id: string;
  name: string;
  is_active?: boolean;
}

export interface ExpenseCategoryCreateRequest {
  name: string;
  description?: string;
  color?: string;
  icon?: string;
  budget?: number;
  is_active?: boolean;
}

// Part Types
export interface PartType {
  id: string;
  name_ar: string;
  name_en: string;
  color?: string;
  description?: string;
  created_at?: string;
  updated_at?: string;
}

export interface PartTypeCreateRequest {
  name_ar: string;
  name_en: string;
  color?: string;
  description?: string;
}

export interface PartTypeUpdateRequest extends Partial<PartTypeCreateRequest> {}

// Barcode Types
export interface BarcodeLookupRequest {
  barcode: string;
}

export interface BarcodeLookupResponse {
  id: string;
  name: string;
  barcode: string;
  selling_price: number;
  stock?: number;
  condition?: string;
}
